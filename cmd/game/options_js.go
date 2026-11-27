//go:build js

package main

import "syscall/js"

// parseOptions reads the page's query string, such as ?pad=1.
func parseOptions() options {
	params := js.Global().Get("URLSearchParams").New(js.Global().Get("location").Get("search"))
	return options{touchPad: params.Call("get", "pad").String() == "1"}
}
