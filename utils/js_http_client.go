package utils

import (
	"fmt"
	"syscall/js"
)

func Get(url string, header map[string]any) ([]byte, error) {
	// 创建 Promise 完成通道
	done := make(chan []byte)
	fail := make(chan error)

	// 创建 JavaScript 的 Headers 对象
	jsHeaders := js.Global().Get("Object").New()
	for k, v := range header {
		jsHeaders.Set(k, v)
	}

	// 创建请求配置对象
	opts := js.Global().Get("Object").New()
	opts.Set("method", "GET")
	opts.Set("headers", jsHeaders)
	opts.Set("credentials", "include")

	// 调用 fetch API
	promise := js.Global().Get("fetch").Invoke(url, opts)

	// 处理响应
	promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		response := args[0]
		if !response.Get("ok").Bool() {
			fail <- fmt.Errorf("HTTP error: %s", response.Get("statusText").String())
			return nil
		}

		// 获取响应体的 arrayBuffer
		response.Call("arrayBuffer").Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			// 将 ArrayBuffer 转换为 Uint8Array
			uint8Array := js.Global().Get("Uint8Array").New(args[0])
			data := make([]byte, uint8Array.Length())
			js.CopyBytesToGo(data, uint8Array)
			done <- data
			return nil
		}))
		return nil
	})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fail <- fmt.Errorf("fetch error: %s", args[0].String())
		return nil
	}))

	// 等待请求完成或失败
	select {
	case data := <-done:
		return data, nil
	case err := <-fail:
		return nil, err
	}
}

// 添加新的函数支持带请求体的请求
func Post(url string, header map[string]any, body []byte) ([]byte, error) {
	done := make(chan []byte)
	fail := make(chan error)

	jsHeaders := js.Global().Get("Object").New()
	for k, v := range header {
		jsHeaders.Set(k, v)
	}

	opts := js.Global().Get("Object").New()
	opts.Set("method", "POST")
	opts.Set("headers", jsHeaders)
	opts.Set("credentials", "include")

	// 如果有请求体，创建 Uint8Array 并设置
	if len(body) > 0 {
		uint8Array := js.Global().Get("Uint8Array").New(len(body))
		js.CopyBytesToJS(uint8Array, body)
		opts.Set("body", uint8Array)
	}

	// 原有的 fetch 和响应处理逻辑保持不变
	promise := js.Global().Get("fetch").Invoke(url, opts)

	// 处理响应
	promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		response := args[0]
		if !response.Get("ok").Bool() {
			fail <- fmt.Errorf("HTTP error: %s", response.Get("statusText").String())
			return nil
		}

		// 获取响应体的 arrayBuffer
		response.Call("arrayBuffer").Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			// 将 ArrayBuffer 转换为 Uint8Array
			uint8Array := js.Global().Get("Uint8Array").New(args[0])
			data := make([]byte, uint8Array.Length())
			js.CopyBytesToGo(data, uint8Array)
			done <- data
			return nil
		}))
		return nil
	})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fail <- fmt.Errorf("fetch error: %s", args[0].String())
		return nil
	}))

	// 等待请求完成或失败
	select {
	case data := <-done:
		return data, nil
	case err := <-fail:
		return nil, err
	}
}

// RequestHead 发送 HEAD 请求并返回响应头
func Head(url string, header map[string]any) (map[string]string, error) {
	done := make(chan map[string]string)
	fail := make(chan error)

	jsHeaders := js.Global().Get("Object").New()
	for k, v := range header {
		jsHeaders.Set(k, v)
	}

	opts := js.Global().Get("Object").New()
	opts.Set("method", "HEAD")
	opts.Set("headers", jsHeaders)
	opts.Set("credentials", "include")

	promise := js.Global().Get("fetch").Invoke(url, opts)

	promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		response := args[0]
		if !response.Get("ok").Bool() {
			fail <- fmt.Errorf("HTTP error: %s", response.Get("statusText").String())
			return nil
		}

		// 获取响应头
		headers := make(map[string]string)
		responseHeaders := response.Get("headers")
		responseHeaders.Call("forEach", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			value := args[0].String()
			key := args[1].String()
			headers[key] = value
			return nil
		}))

		done <- headers
		return nil
	})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fail <- fmt.Errorf("fetch error: %s", args[0].String())
		return nil
	}))

	select {
	case headers := <-done:
		return headers, nil
	case err := <-fail:
		return nil, err
	}
}
