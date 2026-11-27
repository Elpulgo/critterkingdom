//go:build !js

package main

import "flag"

func parseOptions() options {
	pad := flag.Bool("pad", false, "steer with a floating touch pad on the left half of a touch screen")
	flag.Parse()
	return options{touchPad: *pad}
}
