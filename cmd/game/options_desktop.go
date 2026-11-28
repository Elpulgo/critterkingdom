//go:build !js

package main

import "flag"

func parseOptions() options {
	pad := flag.Bool("pad", false, "steer with a floating touch pad on the left half of a touch screen")
	startMap := flag.String("map", defaultStartMap, "start on the named map from assets/maps, such as meadow")
	lang := flag.String("lang", "", "language to start in: sv (Swedish, the default) or en (English)")
	preset := flag.String("preset", "explorer", "age preset: explorer, or little-explorer for gentler and fewer enemies")
	peaceful := flag.Bool("peaceful", false, "put every enemy to sleep, so none of them moves or attacks")
	flag.Parse()
	return options{
		touchPad: *pad,
		startMap: *startMap,
		lang:     *lang,
		preset:   *preset,
		peaceful: *peaceful,
	}
}
