package artexport

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"
)

// NewInkscape returns the Inkscape runner that calls the Inkscape 1.x
// command line. binary is a name looked up on PATH or a path to the
// executable; it is resolved on the first call, so a run with nothing to
// export never needs Inkscape installed.
func NewInkscape(binary string) Inkscape {
	return &inkscapeCLI{binary: binary}
}

// inkscapeCLI runs Inkscape as a separate process, one call per command.
type inkscapeCLI struct {
	binary string
	path   string
}

func (c *inkscapeCLI) QueryAll(svg string) ([]byte, error) {
	out, _, err := c.run("--query-all", svg)
	return out, err
}

func (c *inkscapeCLI) ExportID(svg, id, png string) error {
	return c.export(png, svg,
		"--export-type=png",
		"--export-id="+id,
		"--export-id-only",
		"--export-area-snap",
		"--export-dpi=96",
		"--export-filename="+png,
	)
}

func (c *inkscapeCLI) ExportPage(svg, png string) error {
	return c.export(png, svg,
		"--export-type=png",
		"--export-area-page",
		"--export-dpi=96",
		"--export-filename="+png,
	)
}

// export runs an export command and checks that it wrote png. Inkscape can
// exit successfully without writing anything (for example for an unknown
// id), so the old file is removed first and its absence afterwards is an
// error.
func (c *inkscapeCLI) export(png string, args ...string) error {
	if err := os.Remove(png); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	_, stderr, err := c.run(args...)
	if err != nil {
		return err
	}
	if _, err := os.Stat(png); err != nil {
		return fmt.Errorf("inkscape %s: wrote no %s: %s", strings.Join(args, " "), png, stderr)
	}
	return nil
}

// run runs Inkscape with args and returns its stdout and trimmed stderr.
// Inkscape prints warnings on stderr, so stderr is only part of the error
// when the command fails.
func (c *inkscapeCLI) run(args ...string) ([]byte, string, error) {
	if c.path == "" {
		p, err := exec.LookPath(c.binary)
		if err != nil {
			return nil, "", fmt.Errorf("cannot run Inkscape %q (%v): install Inkscape 1.x or pass its path with the -inkscape flag, for example make art INKSCAPE=/path/to/inkscape", c.binary, err)
		}
		c.path = p
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(c.path, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	msg := strings.TrimSpace(stderr.String())
	if err != nil {
		if msg == "" {
			return nil, msg, fmt.Errorf("inkscape %s: %w", strings.Join(args, " "), err)
		}
		return nil, msg, fmt.Errorf("inkscape %s: %w: %s", strings.Join(args, " "), err, msg)
	}
	return stdout.Bytes(), msg, nil
}
