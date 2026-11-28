// Command audio renders every sound in the tools/audio/synth catalog to a
// WAV file and encodes it to OGG Vorbis with ffmpeg: sound effects and
// voices at -q:a 4, music at -q:a 3. It always re-encodes everything and
// deletes OGG files under the output folder that the catalog no longer
// lists.
//
//	go run ./tools/audio [-out assets/audio] [-wav build/audio-src] [-ffmpeg ffmpeg]
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Elpulgo/critterkingdom/tools/audio/audioexport"
)

func main() {
	out := flag.String("out", "assets/audio", "output folder for the OGG files")
	wav := flag.String("wav", "build/audio-src", "folder for the rendered WAV files")
	ffmpeg := flag.String("ffmpeg", "ffmpeg", "ffmpeg executable with libvorbis, a name on PATH or a path")
	flag.Parse()
	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "audio: unexpected arguments: %v\n", flag.Args())
		flag.Usage()
		os.Exit(2)
	}

	err := audioexport.Run(audioexport.Options{
		OutDir:  *out,
		WAVDir:  *wav,
		Encoder: audioexport.NewFFmpeg(*ffmpeg),
		Log:     os.Stdout,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "audio:", err)
		os.Exit(1)
	}
}
