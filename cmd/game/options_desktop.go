//go:build !js

package main

import "flag"

func parseOptions() options {
	pad := flag.Bool("pad", false, "steer with a floating touch pad on the left half of a touch screen")
	startMap := flag.String("map", defaultStartMap, "start on the named map from assets/maps, such as meadow")
	flag.Parse()
	return options{touchPad: *pad, startMap: *startMap}
}
