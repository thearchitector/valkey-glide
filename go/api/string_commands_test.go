// Copyright Valkey GLIDE Project Contributors - SPDX Identifier: Apache-2.0

package api

import (
	"context"
	"fmt"

	"github.com/valkey-io/valkey-glide/go/api/options"
)

func ExampleGlideClient_Set() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	result, err := client.Set(context.Background(), "my_key", "my_value")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: OK
}

func ExampleGlideClusterClient_Set() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	result, err := client.Set(context.Background(), "my_key", "my_value")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: OK
}

func ExampleGlideClient_SetWithOptions() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	options := options.NewSetOptions().
		SetExpiry(options.NewExpiry().
			SetType(options.Seconds).
			SetCount(5))
	result, err := client.SetWithOptions(context.Background(), "my_key", "my_value", *options)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result.Value())

	// Output: OK
}

func ExampleGlideClusterClient_SetWithOptions() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	options := options.NewSetOptions().
		SetExpiry(options.NewExpiry().
			SetType(options.Seconds).
			SetCount(uint64(5)))
	result, err := client.SetWithOptions(context.Background(), "my_key", "my_value", *options)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result.Value())

	// Output: OK
}

func ExampleGlideClient_Get_keyexists() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key", "my_value")
	result, err := client.Get(context.Background(), "my_key")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result.Value())

	// Output: my_value
}

func ExampleGlideClusterClient_Get_keyexists() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "my_key", "my_value")
	result, err := client.Get(context.Background(), "my_key")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result.Value())

	// Output: my_value
}

func ExampleGlideClient_Get_keynotexists() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	result, err := client.Get(context.Background(), "my_key")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result.IsNil()) // missing key returns nil result

	// Output: true
}

func ExampleGlideClusterClient_Get_keynotexists() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	result, err := client.Get(context.Background(), "my_key")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result.IsNil()) // missing key returns nil result

	// Output: true
}

func ExampleGlideClient_GetEx() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key", "my_value")
	result, err := client.GetEx(context.Background(), "my_key")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result.Value())
	ttl, _ := client.TTL(context.Background(), "my_key")
	fmt.Println(ttl)

	// Output:
	// my_value
	// -1
}

func ExampleGlideClusterClient_GetEx() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "my_key", "my_value")
	result, err := client.GetEx(context.Background(), "my_key")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result.Value())
	ttl, _ := client.TTL(context.Background(), "my_key")
	fmt.Println(ttl)

	// Output:
	// my_value
	// -1
}

func ExampleGlideClient_GetExWithOptions() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key", "my_value")
	options := options.NewGetExOptions().
		SetExpiry(options.NewExpiry().
			SetType(options.Seconds).
			SetCount(5))
	result, err := client.GetExWithOptions(context.Background(), "my_key", *options)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result.Value())
	ttl, _ := client.TTL(context.Background(), "my_key")
	fmt.Println(ttl)

	// Output:
	// my_value
	// 5
}

func ExampleGlideClusterClient_GetExWithOptions() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "my_key", "my_value")
	options := options.NewGetExOptions().
		SetExpiry(options.NewExpiry().
			SetType(options.Seconds).
			SetCount(uint64(5)))
	result, err := client.GetExWithOptions(context.Background(), "my_key", *options)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result.Value())
	ttl, _ := client.TTL(context.Background(), "my_key")
	fmt.Println(ttl)

	// Output:
	// my_value
	// 5
}

func ExampleGlideClient_MSet() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	keyValueMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	result, err := client.MSet(context.Background(), keyValueMap)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: OK
}

func ExampleGlideClusterClient_MSet() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	keyValueMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	result, err := client.MSet(context.Background(), keyValueMap)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: OK
}

func ExampleGlideClient_MGet() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.MSet(
		context.Background(),
		map[string]string{"my_key1": "my_value1", "my_key2": "my_value2", "my_key3": "my_value3"},
	)
	keys := []string{"my_key1", "my_key2", "my_key3"}
	result, err := client.MGet(context.Background(), keys)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	for _, res := range result {
		fmt.Println(res.Value())
	}

	// Output:
	// my_value1
	// my_value2
	// my_value3
}

func ExampleGlideClusterClient_MGet() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.MSet(
		context.Background(),
		map[string]string{"my_key1": "my_value1", "my_key2": "my_value2", "my_key3": "my_value3"},
	)
	keys := []string{"my_key1", "my_key2", "my_key3"}
	result, err := client.MGet(context.Background(), keys)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	for _, res := range result {
		fmt.Println(res.Value())
	}

	// Output:
	// my_value1
	// my_value2
	// my_value3
}

func ExampleGlideClient_MSetNX() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	keyValueMap := map[string]string{"my_key1": "my_value1", "my_key2": "my_value2"}
	result, err := client.MSetNX(context.Background(), keyValueMap)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)
	client.Set(context.Background(), "my_key3", "my_value3")
	result, _ = client.MSetNX(context.Background(), map[string]string{"my_key3": "my_value3"})
	fmt.Println(result)

	// Output:
	// true
	// false
}

func ExampleGlideClusterClient_MSetNX() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	keyValueMap := map[string]string{"{my_key}1": "my_value1", "{my_key}2": "my_value2"}
	result, err := client.MSetNX(context.Background(), keyValueMap)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)
	client.Set(context.Background(), "{my_key}3", "my_value3")
	result, _ = client.MSetNX(context.Background(), map[string]string{"{my_key}3": "my_value3"})
	fmt.Println(result)

	// Output:
	// true
	// false
}

func ExampleGlideClient_Incr() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key", "1")
	result, err := client.Incr(context.Background(), "my_key")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: 2
}

func ExampleGlideClusterClient_Incr() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "my_key", "1")
	result, err := client.Incr(context.Background(), "my_key")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: 2
}

func ExampleGlideClient_IncrBy() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key", "5")
	result, err := client.IncrBy(context.Background(), "my_key", 5)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)
	// Output: 10
}

func ExampleGlideClusterClient_IncrBy() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "my_key", "5")
	result, err := client.IncrBy(context.Background(), "my_key", 5)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)
	// Output: 10
}

func ExampleGlideClient_IncrByFloat() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key", "1")
	result, err := client.IncrByFloat(context.Background(), "my_key", 5.5)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: 6.5
}

func ExampleGlideClusterClient_IncrByFloat() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "my_key", "1")
	result, err := client.IncrByFloat(context.Background(), "my_key", 5.5)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: 6.5
}

func ExampleGlideClient_Decr() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key", "0")
	result, err := client.Decr(context.Background(), "my_key")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: -1
}

func ExampleGlideClusterClient_Decr() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "my_key", "0")
	result, err := client.Decr(context.Background(), "my_key")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: -1
}

func ExampleGlideClient_DecrBy() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key", "20")
	result, err := client.DecrBy(context.Background(), "my_key", 5)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: 15
}

func ExampleGlideClusterClient_DecrBy() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "my_key", "20")
	result, err := client.DecrBy(context.Background(), "my_key", 5)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: 15
}

func ExampleGlideClient_Strlen() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key", "my_value")
	result, err := client.Strlen(context.Background(), "my_key")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: 8
}

func ExampleGlideClusterClient_Strlen() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "my_key", "my_value")
	result, err := client.Strlen(context.Background(), "my_key")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: 8
}

func ExampleGlideClient_SetRange_one() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key", "my_value")
	result, err := client.SetRange(context.Background(), "my_key", 3, "example")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)
	value, _ := client.Get(context.Background(), "my_key")
	fmt.Println(value.Value())

	// Output:
	// 10
	// my_example
}

func ExampleGlideClusterClient_SetRange_one() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "my_key", "my_value")
	result, err := client.SetRange(context.Background(), "my_key", 3, "example")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)
	value, _ := client.Get(context.Background(), "my_key")
	fmt.Println(value.Value())

	// Output:
	// 10
	// my_example
}

func ExampleGlideClient_SetRange_two() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key", "愛") // "愛" is a single character in UTF-8, but 3 bytes long
	result, err := client.SetRange(context.Background(), "my_key", 1, "a")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: 3
}

func ExampleGlideClusterClient_SetRange_two() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "my_key", "愛") // "愛" is a single character in UTF-8, but 3 bytes long
	result, err := client.SetRange(context.Background(), "my_key", 1, "a")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: 3
}

func ExampleGlideClient_GetRange_one() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key", "Welcome to Valkey Glide!")
	result, err := client.GetRange(context.Background(), "my_key", 0, 7)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: Welcome
}

func ExampleGlideClusterClient_GetRange_one() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "my_key", "Welcome to Valkey Glide!")
	result, err := client.GetRange(context.Background(), "my_key", 0, 7)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// Output: Welcome
}

func ExampleGlideClient_GetRange_two() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key", "愛")
	fmt.Println([]byte("愛")) // "愛" is a single character in UTF-8, but 3 bytes long
	result, err := client.GetRange(context.Background(), "my_key", 0, 1)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println([]byte(result))

	// Output:
	// [230 132 155]
	// [230 132]
}

func ExampleGlideClusterClient_GetRange_two() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "my_key", "愛")
	fmt.Println([]byte("愛")) // "愛" is a single character in UTF-8, but 3 bytes long
	result, err := client.GetRange(context.Background(), "my_key", 0, 1)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println([]byte(result))

	// Output:
	// [230 132 155]
	// [230 132]
}

func ExampleGlideClient_Append() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key", "my_valu")
	result, err := client.Append(context.Background(), "my_key", "e")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)
	value, _ := client.Get(context.Background(), "my_key")
	fmt.Println(value.Value())

	// Output:
	// 8
	// my_value
}

func ExampleGlideClusterClient_Append() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "my_key", "my_valu")
	result, err := client.Append(context.Background(), "my_key", "e")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)
	value, _ := client.Get(context.Background(), "my_key")
	fmt.Println(value.Value())

	// Output:
	// 8
	// my_value
}

func ExampleGlideClient_LCS() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.MSet(context.Background(), map[string]string{"my_key1": "oh my gosh", "my_key2": "hello world"})
	result, err := client.LCS(context.Background(), "my_key1", "my_key2")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// LCS is only available in 7.0 and above. It will fail in any server < 7.0
	// Output: h o
}

func ExampleGlideClusterClient_LCS() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.MSet(context.Background(), map[string]string{"{my_key}1": "oh my gosh", "{my_key}2": "hello world"})
	result, err := client.LCS(context.Background(), "{my_key}1", "{my_key}2")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// LCS is only available in 7.0 and above. It will fail in any release < 7.0
	// Output: h o
}

func ExampleGlideClient_GetDel() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key", "my_value")
	result, err := client.GetDel(context.Background(), "my_key") // return value and delete key
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result.Value())
	value, _ := client.Get(context.Background(), "my_key") // key should be missing
	fmt.Println(value.IsNil())

	// Output:
	// my_value
	// true
}

func ExampleGlideClusterClient_GetDel() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "my_key", "my_value")
	result, err := client.GetDel(context.Background(), "my_key") // return value and delete key
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result.Value())
	value, _ := client.Get(context.Background(), "my_key") // key should be missing
	fmt.Println(value.IsNil())

	// Output:
	// my_value
	// true
}

func ExampleGlideClient_LCSLen() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key1", "ohmytext")
	client.Set(context.Background(), "my_key2", "mynewtext")

	result, err := client.LCSLen(context.Background(), "my_key1", "my_key2")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// LCS is only available in 7.0 and above. It will fail in any server < 7.0
	// Output: 3
}

func ExampleGlideClusterClient_LCSLen() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "{my_key}1", "ohmytext")
	client.Set(context.Background(), "{my_key}2", "mynewtext")

	result, err := client.LCSLen(context.Background(), "{my_key}1", "{my_key}2")
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println(result)

	// LCS is only available in 7.0 and above. It will fail in any release < 7.0
	// Output: 3
}

func ExampleGlideClient_LCSWithOptions() {
	var client *GlideClient = getExampleGlideClient() // example helper function

	client.Set(context.Background(), "my_key1", "ohmytext")
	client.Set(context.Background(), "my_key2", "mynewtext")

	// Basic LCS IDX without additional options
	opts := options.NewLCSIdxOptions()
	result1, err := client.LCSWithOptions(context.Background(), "my_key1", "my_key2", *opts)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println("Basic LCS result:", result1)

	// LCS IDX with MINMATCHLEN = 4
	optsWithMin := options.NewLCSIdxOptions()
	optsWithMin.SetMinMatchLen(4)
	result2, err := client.LCSWithOptions(context.Background(), "my_key1", "my_key2", *optsWithMin)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}
	fmt.Println("With MinMatchLen 4:", result2)

	// LCS is only available in 7.0 and above. It will fail in any server < 7.0
	// Output:
	// Basic LCS result: map[len:3 matches:[[0 1 0 1] [6 7 4 5]]]
	// With MinMatchLen 4: map[len:0 matches:[]]
}

func ExampleGlideClusterClient_LCSWithOptions() {
	var client *GlideClusterClient = getExampleGlideClusterClient() // example helper function

	client.Set(context.Background(), "{my_key}1", "ohmytext")
	client.Set(context.Background(), "{my_key}2", "mynewtext")

	// LCS IDX with MINMATCHLEN and WITHMATCHLEN
	opts := options.NewLCSIdxOptions()
	opts.SetMinMatchLen(2)
	opts.SetWithMatchLen(true)
	result, err := client.LCSWithOptions(context.Background(), "{my_key}1", "{my_key}2", *opts)
	if err != nil {
		fmt.Println("Glide example failed with an error: ", err)
	}

	fmt.Println("Full result with both options:", result)

	// LCS is only available in 7.0 and above. It will fail in any release < 7.0
	// Output:
	// Full result with both options: map[len:3 matches:[[0 1 0 1 2] [6 7 4 5 2]]]
}
