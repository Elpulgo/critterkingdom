// Command art exports the Inkscape drawings in assets-src/ that changed
// since their last export: character parts to PNGs plus a part JSON, and
// props and tiles to one PNG each, at 8 PNG pixels per world pixel.
//
//	go run ./tools/art [-src assets-src] [-out assets] [-inkscape inkscape] [-force]
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Elpulgo/critterkingdom/tools/art/artexport"
)

func main() {
	src := flag.String("src", "assets-src", "source root holding sprites/ and tiles/")
	out := flag.String("out", "assets", "output root for the PNGs and part JSON")
	inkscape := flag.String("inkscape", "inkscape", "Inkscape 1.x executable, a name on PATH or a path")
	force := flag.Bool("force", false, "export every drawing, changed or not")
	flag.Parse()
	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "art: unexpected arguments: %v\n", flag.Args())
		flag.Usage()
		os.Exit(2)
	}

	err := artexport.Run(artexport.Options{
		SrcDir:   *src,
		OutDir:   *out,
		Force:    *force,
		Inkscape: artexport.NewInkscape(*inkscape),
		Log:      os.Stdout,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "art:", err)
		os.Exit(1)
	}
}
