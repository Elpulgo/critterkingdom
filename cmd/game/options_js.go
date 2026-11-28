//go:build js

package main

import "syscall/js"

// parseOptions reads the page's query string, such as ?pad=1&lang=en or
// ?preset=little-explorer&peaceful=1. The browser build always starts on the
// default map.
func parseOptions() options {
	params := js.Global().Get("URLSearchParams").New(js.Global().Get("location").Get("search"))
	opts := options{
		touchPad: params.Call("get", "pad").String() == "1",
		peaceful: params.Call("get", "peaceful").String() == "1",
		startMap: defaultStartMap,
	}
	if lang := params.Call("get", "lang"); !lang.IsNull() {
		opts.lang = lang.String()
	}
	if preset := params.Call("get", "preset"); !preset.IsNull() {
		opts.preset = preset.String()
	}
	return opts
}
