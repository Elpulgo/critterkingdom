//go:build !js

package main

import "flag"

func parseOptions() options {
	pad := flag.Bool("pad", false, "steer with a floating touch pad on the left half of a touch screen")
	lang := flag.String("lang", "", "language to start in: sv (Swedish, the default) or en (English)")
	flag.Parse()
	return options{touchPad: *pad, lang: *lang}
}
