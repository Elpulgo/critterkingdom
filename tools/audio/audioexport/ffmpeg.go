package audioexport

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// NewFFmpeg returns the Encoder that runs the ffmpeg command line. binary
// is a name looked up on PATH or a path to the executable; it is resolved
// on the first Encode call.
func NewFFmpeg(binary string) Encoder {
	return &ffmpegCLI{binary: binary}
}

// ffmpegCLI runs ffmpeg as a separate process, one call per file.
type ffmpegCLI struct {
	binary string
	path   string
}

// Encode runs
//
//	ffmpeg -y -i wav -map_metadata -1 -bitexact -ac 1 -c:a libvorbis -q:a N ogg
//
// with the banner and progress output turned off. -map_metadata -1 and
// -bitexact keep encoder names and dates out of the file, so the same WAV
// gives the same OGG.
func (c *ffmpegCLI) Encode(wav, ogg string, quality int) error {
	if c.path == "" {
		p, err := exec.LookPath(c.binary)
		if err != nil {
			return fmt.Errorf("cannot run ffmpeg %q (%v): install ffmpeg with libvorbis or pass its path with the -ffmpeg flag, for example make audio FFMPEG=/path/to/ffmpeg", c.binary, err)
		}
		c.path = p
	}
	args := []string{
		"-hide_banner", "-nostdin", "-loglevel", "error",
		"-y", "-i", wav,
		"-map_metadata", "-1", "-bitexact",
		"-ac", "1",
		"-c:a", "libvorbis", "-q:a", strconv.Itoa(quality),
		ogg,
	}
	var stderr bytes.Buffer
	cmd := exec.Command(c.path, args...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return fmt.Errorf("ffmpeg %s: %w", strings.Join(args, " "), err)
		}
		return fmt.Errorf("ffmpeg %s: %w: %s", strings.Join(args, " "), err, msg)
	}
	return nil
}
