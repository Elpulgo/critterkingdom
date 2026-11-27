//go:build js

package main

import "syscall/js"

// parseOptions reads the page's query string, such as ?pad=1&lang=en.
func parseOptions() options {
	params := js.Global().Get("URLSearchParams").New(js.Global().Get("location").Get("search"))
	opts := options{touchPad: params.Call("get", "pad").String() == "1"}
	if lang := params.Call("get", "lang"); !lang.IsNull() {
		opts.lang = lang.String()
	}
	return opts
}
