/* tslint:disable */
/* eslint-disable */

/* auto-generated by NAPI-RS */

export class AsyncClient {
  static CreateConnection(connectionAddress: string): Promise<AsyncClient>
  get(key: string): Promise<string | null>
  set(key: string, value: string): Promise<void>
}
