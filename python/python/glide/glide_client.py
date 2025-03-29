# Copyright Valkey GLIDE Project Contributors - SPDX Identifier: Apache-2.0

import sys
import threading
from contextlib import AsyncExitStack
from types import TracebackType
from typing import Any, Dict, List, Optional, Tuple, Type, Union, cast

import anyio
from anyio.abc import ByteStream, TaskGroup
from typing_extensions import Self

from glide.async_commands.cluster_commands import ClusterCommands
from glide.async_commands.command_args import ObjectType
from glide.async_commands.core import CoreCommands
from glide.async_commands.standalone_commands import StandaloneCommands
from glide.config import BaseClientConfiguration, ServerCredentials
from glide.constants import DEFAULT_READ_BYTES_SIZE, OK, TEncodable, TRequest, TResult
from glide.exceptions import (
    ClosingError,
    ConfigurationError,
    ConnectionError,
    ExecAbortError,
    RequestError,
    TimeoutError,
)
from glide.logger import Level as LogLevel
from glide.logger import Logger as ClientLogger
from glide.protobuf.command_request_pb2 import Command, CommandRequest, RequestType
from glide.protobuf.connection_request_pb2 import ConnectionRequest
from glide.protobuf.response_pb2 import RequestErrorType, Response
from glide.protobuf_codec import PartialMessageException, ProtobufCodec
from glide.routes import Route, set_protobuf_route

from .glide import (
    DEFAULT_TIMEOUT_IN_MILLISECONDS,
    MAX_REQUEST_ARGS_LEN,
    ClusterScanCursor,
    create_leaked_bytes_vec,
    get_statistics,
    start_socket_listener_external,
    value_from_pointer,
)


def get_request_error_class(
    error_type: Optional[RequestErrorType.ValueType],
) -> Type[RequestError]:
    if error_type == RequestErrorType.Disconnect:
        return ConnectionError
    if error_type == RequestErrorType.ExecAbort:
        return ExecAbortError
    if error_type == RequestErrorType.Timeout:
        return TimeoutError
    if error_type == RequestErrorType.Unspecified:
        return RequestError
    return RequestError


class Result:
    def __init__(self) -> None:
        self._is_done = anyio.Event()
        self._value: Any = None
        self._exception: Optional[Exception] = None

    def success(self, value: Any) -> None:
        self._value = value
        self._is_done.set()

    def failure(self, exception: Exception) -> None:
        self._exception = exception
        self._is_done.set()

    def is_done(self) -> bool:
        return self._is_done.is_set()

    async def wait(self) -> Any:
        await self._is_done.wait()

        if self._exception is not None:
            raise self._exception
        return self._value


class BaseClient(CoreCommands):
    def __init__(self, config: BaseClientConfiguration):
        """Creates a Glide client.

        Args:
            config (ClientConfiguration): The client configurations.
                If no configuration is provided, a default client to "localhost":6379 will be created.

        Returns:
            Self: a Glide Client instance.
        """
        self.config: BaseClientConfiguration = config

        self._task_exitstack: Optional[AsyncExitStack] = None
        self._task_group: Optional[TaskGroup] = None

        self.socket_path: Optional[str] = None
        self._stream: Optional[ByteStream] = None
        self._writer_lock = threading.Lock()

        self._pending_results: Dict[int, Result] = {}
        self._available_callback_indexes: List[int] = list()
        self._buffered_requests: List[TRequest] = list()

        self._pubsub_results: List[Result] = []
        self._pubsub_lock = threading.Lock()
        self._pending_push_notifications: List[Response] = list()

        self._is_closed: bool = False

    async def __aenter__(self) -> Self:
        init_event = threading.Event()

        def init_callback(socket_path: Optional[str], err: Optional[str]):
            if err is not None:
                raise ClosingError(err)
            elif socket_path is None:
                raise ClosingError(
                    "Socket initialization error: Missing valid socket path."
                )
            else:
                # Received socket path
                self.socket_path = socket_path
                init_event.set()

        start_socket_listener_external(init_callback=init_callback)
        await anyio.to_thread.run_sync(init_event.wait)

        # Create UDS connection
        await self._create_uds_connection()

        # will log if the logger was created (wrapper or costumer) on info
        # level or higher
        ClientLogger.log(
            LogLevel.INFO,
            "connection info",
            "new connection established",
        )

        async with AsyncExitStack() as stack:
            self._task_group = await stack.enter_async_context(
                anyio.create_task_group()
            )

            # Start the reader loop as a background task
            self._task_group.start_soon(self._stream_demuxer)

            # Set the client configurations
            await self._set_connection_configurations()

            self._task_exitstack = stack.pop_all()

        return self

    async def __aexit__(
        self,
        exc_type: Optional[Type[BaseException]],
        exc_val: Optional[BaseException],
        exc_tb: Optional[TracebackType],
    ) -> None:
        """
        Terminate the client by closing all associated resources, including the socket and any pending requests.
        """
        self._is_closed = True
        err_message = str(exc_val)

        for result in self._pending_results.values():
            if not result.is_done():
                result.failure(ClosingError(err_message))
        try:
            self._pubsub_lock.acquire()
            for pubsub_result in self._pubsub_results:
                if not pubsub_result.is_done():
                    pubsub_result.failure(ClosingError(err_message))
        finally:
            self._pubsub_lock.release()

        if self._stream:
            await self._stream.aclose()

        return await self._task_exitstack.__aexit__(exc_type, exc_val, exc_tb)

    async def _create_uds_connection(self) -> None:
        # Open an UDS connection
        with anyio.fail_after(DEFAULT_TIMEOUT_IN_MILLISECONDS):
            self._stream = await anyio.connect_unix(self.socket_path)

    async def _stream_demuxer(self) -> None:
        # socket response reading & demuxer
        remaining_read_bytes = bytearray()

        while True:
            try:
                read_bytes = await self._stream.receive(DEFAULT_READ_BYTES_SIZE)
            except anyio.EndOfStream:
                raise ClosingError("The communication layer was unexpectedly closed.")
            except anyio.ClosedResourceError:
                if self._is_closed:
                    break
                raise

            read_bytes = remaining_read_bytes + bytearray(read_bytes)
            read_bytes_view = memoryview(read_bytes)
            offset = 0
            while offset <= len(read_bytes):
                try:
                    response, offset = ProtobufCodec.decode_delimited(
                        read_bytes, read_bytes_view, offset, Response
                    )
                except PartialMessageException:
                    # Received only partial response, break the inner loop
                    remaining_read_bytes = read_bytes[offset:]
                    break

                response = cast(Response, response)
                if response.is_push:
                    await self._process_push(response=response)
                else:
                    await self._process_response(response=response)

    async def _set_connection_configurations(self) -> None:
        conn_request = self._get_protobuf_conn_request()
        result = await self._write_request_await_response(conn_request)
        if result is not OK:
            raise ClosingError(result)

    def _get_protobuf_conn_request(self) -> ConnectionRequest:
        return self.config._create_a_protobuf_conn_request()

    async def _write_request_await_response(self, request: TRequest):
        result = self._get_result(getattr(request, "callback_idx", 0))
        self._send_request(request)
        return await result.wait()

    def _get_result(self, callback_idx: int) -> Result:
        # Create a pending result for this request and add it to the pending
        # results map
        result: Result = Result()
        self._pending_results[callback_idx] = result
        return result

    def _send_request(self, request: TRequest):
        self._task_group.start_soon(self._write_or_buffer_request, request)

    # TODO: does this actually batch requests?
    async def _write_or_buffer_request(self, request: TRequest):
        self._buffered_requests.append(request)
        if self._writer_lock.acquire(False):
            try:
                while len(self._buffered_requests) > 0:
                    await self._write_buffered_requests_to_socket()

            finally:
                self._writer_lock.release()

    async def _write_buffered_requests_to_socket(self) -> None:
        requests = self._buffered_requests
        self._buffered_requests = list()
        b_arr = bytearray()
        for request in requests:
            ProtobufCodec.encode_delimited(b_arr, request)
        await self._stream.send(b_arr)

    async def _process_push(self, response: Response) -> None:
        if response.HasField("closing_error") or not response.HasField("resp_pointer"):
            err_msg = (
                response.closing_error
                if response.HasField("closing_error")
                else "Client Error - push notification without resp_pointer"
            )
            # await self.aclose(err_msg)
            raise ClosingError(err_msg)

        try:
            self._pubsub_lock.acquire()
            callback, context = self.config._get_pubsub_callback_and_context()
            if callback:
                pubsub_message = self._notification_to_pubsub_message_safe(response)
                if pubsub_message:
                    callback(pubsub_message, context)
            else:
                self._pending_push_notifications.append(response)
                self._complete_pubsub_results_safe()
        finally:
            self._pubsub_lock.release()

    def _notification_to_pubsub_message_safe(
        self, response: Response
    ) -> Optional[CoreCommands.PubSubMsg]:
        pubsub_message = None
        push_notification = cast(
            Dict[str, Any], value_from_pointer(response.resp_pointer)
        )
        message_kind = push_notification["kind"]
        if message_kind == "Disconnection":
            ClientLogger.log(
                LogLevel.WARN,
                "disconnect notification",
                "Transport disconnected, messages might be lost",
            )
        elif (
            message_kind == "Message"
            or message_kind == "PMessage"
            or message_kind == "SMessage"
        ):
            values: List = push_notification["values"]
            if message_kind == "PMessage":
                pubsub_message = BaseClient.PubSubMsg(
                    message=values[2], channel=values[1], pattern=values[0]
                )
            else:
                pubsub_message = BaseClient.PubSubMsg(
                    message=values[1], channel=values[0], pattern=None
                )
        elif (
            message_kind == "PSubscribe"
            or message_kind == "Subscribe"
            or message_kind == "SSubscribe"
            or message_kind == "Unsubscribe"
            or message_kind == "PUnsubscribe"
            or message_kind == "SUnsubscribe"
        ):
            pass
        else:
            ClientLogger.log(
                LogLevel.WARN,
                "unknown notification",
                f"Unknown notification message: '{message_kind}'",
            )

        return pubsub_message

    def _complete_pubsub_results_safe(self):
        while len(self._pending_push_notifications) and len(self._pubsub_results):
            next_push_notification = self._pending_push_notifications.pop(0)
            pubsub_message = self._notification_to_pubsub_message_safe(
                next_push_notification
            )
            if pubsub_message:
                self._pubsub_results.pop(0).set_result(pubsub_message)

    async def _process_response(self, response: Response) -> None:
        result = self._pending_results.pop(response.callback_idx, None)
        if not result or response.HasField("closing_error"):
            err_msg = (
                response.closing_error
                if response.HasField("closing_error")
                else f"Client Error - closing due to unknown error. callback index:  {response.callback_idx}"
            )
            if result is not None:
                result.failure(ClosingError(err_msg))
            else:
                # await self.aclose(err_msg)
                raise ClosingError(err_msg)
        else:
            self._available_callback_indexes.append(response.callback_idx)
            if response.HasField("request_error"):
                error_type = get_request_error_class(response.request_error.type)
                result.failure(error_type(response.request_error.message))
            elif response.HasField("resp_pointer"):
                result.success(value_from_pointer(response.resp_pointer))
            elif response.HasField("constant_response"):
                result.success(OK)
            else:
                result.success(None)

    def try_get_pubsub_message(self) -> Optional[CoreCommands.PubSubMsg]:
        if self._is_closed:
            raise ClosingError(
                "Unable to execute requests; the client is closed. Please create a new client."
            )

        if not self.config._is_pubsub_configured():
            raise ConfigurationError(
                "The operation will never succeed since there was no pubsbub subscriptions applied to the client."
            )

        if self.config._get_pubsub_callback_and_context()[0] is not None:
            raise ConfigurationError(
                "The operation will never succeed since messages will be passed to the configured callback."
            )

        # locking might not be required
        msg: Optional[CoreCommands.PubSubMsg] = None
        try:
            self._pubsub_lock.acquire()
            self._complete_pubsub_results_safe()
            while len(self._pending_push_notifications) and not msg:
                push_notification = self._pending_push_notifications.pop(0)
                msg = self._notification_to_pubsub_message_safe(push_notification)
        finally:
            self._pubsub_lock.release()
        return msg

    async def get_pubsub_message(self) -> CoreCommands.PubSubMsg:
        if self._is_closed:
            raise ClosingError(
                "Unable to execute requests; the client is closed. Please create a new client."
            )

        if not self.config._is_pubsub_configured():
            raise ConfigurationError(
                "The operation will never complete since there was no pubsub subscriptions applied to the client."
            )

        if self.config._get_pubsub_callback_and_context()[0] is not None:
            raise ConfigurationError(
                "The operation will never complete since messages will be passed to the configured callback."
            )

        # locking might not be required
        result = self._get_result(0)
        try:
            self._pubsub_lock.acquire()
            self._pubsub_results.append(result)
            self._complete_pubsub_results_safe()
        finally:
            self._pubsub_lock.release()
        return await result.wait()

    async def _execute_command(
        self,
        request_type: RequestType.ValueType,
        args: List[TEncodable],
        route: Optional[Route] = None,
    ) -> TResult:
        if self._is_closed:
            raise ClosingError(
                "Unable to execute requests; the client is closed. Please create a new client."
            )
        request = CommandRequest()
        request.callback_idx = self._get_callback_index()
        request.single_command.request_type = request_type
        request.single_command.args_array.args[:] = [
            bytes(elem, encoding="utf8") if isinstance(elem, str) else elem
            for elem in args
        ]
        (encoded_args, args_size) = self._encode_and_sum_size(args)
        if args_size < MAX_REQUEST_ARGS_LEN:
            request.single_command.args_array.args[:] = encoded_args
        else:
            request.single_command.args_vec_pointer = create_leaked_bytes_vec(
                encoded_args
            )
        set_protobuf_route(request, route)
        return await self._write_request_await_response(request)

    def _get_callback_index(self) -> int:
        try:
            return self._available_callback_indexes.pop()
        except IndexError:
            # The list is empty
            return len(self._pending_results)

    def _encode_and_sum_size(
        self,
        args_list: Optional[List[TEncodable]],
    ) -> Tuple[List[bytes], int]:
        """
        Encodes the list and calculates the total memory size.

        Args:
            args_list (Optional[List[TEncodable]]): A list of strings to be converted to bytes.
                                                           If None or empty, returns ([], 0).

        Returns:
            int: The total memory size of the encoded arguments in bytes.
        """
        args_size = 0
        encoded_args_list: List[bytes] = []
        if not args_list:
            return (encoded_args_list, args_size)
        for arg in args_list:
            encoded_arg = self._encode_arg(arg) if isinstance(arg, str) else arg
            encoded_args_list.append(encoded_arg)
            args_size += sys.getsizeof(encoded_arg)
        return (encoded_args_list, args_size)

    def _encode_arg(self, arg: TEncodable) -> bytes:
        """
        Converts a string argument to bytes.

        Args:
            arg (str): An encodable argument.

        Returns:
            bytes: The encoded argument as bytes.
        """
        if isinstance(arg, str):
            # TODO: Allow passing different encoding options
            return bytes(arg, encoding="utf8")
        return arg

    async def _execute_transaction(
        self,
        commands: List[Tuple[RequestType.ValueType, List[TEncodable]]],
        route: Optional[Route] = None,
    ) -> List[TResult]:
        if self._is_closed:
            raise ClosingError(
                "Unable to execute requests; the client is closed. Please create a new client."
            )
        request = CommandRequest()
        request.callback_idx = self._get_callback_index()
        transaction_commands = []
        for requst_type, args in commands:
            command = Command()
            command.request_type = requst_type
            # For now, we allow the user to pass the command as array of strings
            # we convert them here into bytes (the datatype that our rust core expects)
            (encoded_args, args_size) = self._encode_and_sum_size(args)
            if args_size < MAX_REQUEST_ARGS_LEN:
                command.args_array.args[:] = encoded_args
            else:
                command.args_vec_pointer = create_leaked_bytes_vec(encoded_args)
            transaction_commands.append(command)
        request.transaction.commands.extend(transaction_commands)
        set_protobuf_route(request, route)
        return await self._write_request_await_response(request)

    async def _execute_script(
        self,
        hash: str,
        keys: Optional[List[Union[str, bytes]]] = None,
        args: Optional[List[Union[str, bytes]]] = None,
        route: Optional[Route] = None,
    ) -> TResult:
        if self._is_closed:
            raise ClosingError(
                "Unable to execute requests; the client is closed. Please create a new client."
            )
        request = CommandRequest()
        request.callback_idx = self._get_callback_index()
        (encoded_keys, keys_size) = self._encode_and_sum_size(keys)
        (encoded_args, args_size) = self._encode_and_sum_size(args)
        if (keys_size + args_size) < MAX_REQUEST_ARGS_LEN:
            request.script_invocation.hash = hash
            request.script_invocation.keys[:] = encoded_keys
            request.script_invocation.args[:] = encoded_args

        else:
            request.script_invocation_pointers.hash = hash
            request.script_invocation_pointers.keys_pointer = create_leaked_bytes_vec(
                encoded_keys
            )
            request.script_invocation_pointers.args_pointer = create_leaked_bytes_vec(
                encoded_args
            )
        set_protobuf_route(request, route)
        return await self._write_request_await_response(request)

    async def _update_connection_password(
        self, password: Optional[str], immediate_auth: bool
    ) -> TResult:
        request = CommandRequest()
        request.callback_idx = self._get_callback_index()
        if password is not None:
            request.update_connection_password.password = password
        request.update_connection_password.immediate_auth = immediate_auth
        response = await self._write_request_await_response(request)
        # Update the client binding side password if managed to change core configuration password
        if response is OK:
            if self.config.credentials is None:
                self.config.credentials = ServerCredentials(password=password or "")
                self.config.credentials.password = password or ""
        return response

    async def get_statistics(self) -> Dict[Any, Any]:
        return get_statistics()


class GlideClusterClient(BaseClient, ClusterCommands):
    """
    Client used for connection to cluster servers.
    For full documentation, see
    https://github.com/valkey-io/valkey-glide/wiki/Python-wrapper#cluster
    """

    async def _cluster_scan(
        self,
        cursor: ClusterScanCursor,
        match: Optional[TEncodable] = None,
        count: Optional[int] = None,
        type: Optional[ObjectType] = None,
        allow_non_covered_slots: bool = False,
    ) -> List[Union[ClusterScanCursor, List[bytes]]]:
        if self._is_closed:
            raise ClosingError(
                "Unable to execute requests; the client is closed. Please create a new client."
            )
        request = CommandRequest()
        request.callback_idx = self._get_callback_index()
        # Take out the id string from the wrapping object
        cursor_string = cursor.get_cursor()
        request.cluster_scan.cursor = cursor_string
        request.cluster_scan.allow_non_covered_slots = allow_non_covered_slots
        if match is not None:
            request.cluster_scan.match_pattern = (
                self._encode_arg(match) if isinstance(match, str) else match
            )
        if count is not None:
            request.cluster_scan.count = count
        if type is not None:
            request.cluster_scan.object_type = type.value
        response = await self._write_request_await_response(request)
        return [ClusterScanCursor(bytes(response[0]).decode()), response[1]]

    def _get_protobuf_conn_request(self) -> ConnectionRequest:
        return self.config._create_a_protobuf_conn_request(cluster_mode=True)


class GlideClient(BaseClient, StandaloneCommands):
    """
    Client used for connection to standalone servers.
    For full documentation, see
    https://github.com/valkey-io/valkey-glide/wiki/Python-wrapper#standalone
    """


TGlideClient = Union[GlideClient, GlideClusterClient]
