package artexport

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Inkscape is the part of Inkscape's command line the exporter needs.
type Inkscape interface {
	// QueryAll returns the stdout of `inkscape --query-all <svg>`.
	QueryAll(svg string) ([]byte, error)
	// ExportID exports the object with the given id to a PNG, snapped to
	// whole pixels at 1 drawing px per PNG pixel.
	ExportID(svg, id, png string) error
	// ExportPage exports the whole page of the drawing to a PNG at 1 drawing
	// px per PNG pixel.
	ExportPage(svg, png string) error
}

// Options configures Run.
type Options struct {
	// SrcDir is the source root holding sprites/ and tiles/; defaults to "assets-src".
	SrcDir string
	// OutDir is the output root; defaults to "assets".
	OutDir string
	// Force exports every drawing, changed or not.
	Force bool
	// Inkscape runs the exports.
	Inkscape Inkscape
	// Log receives one line per exported drawing; nil discards it.
	Log io.Writer
}

// sourceFolders are the folders under SrcDir that hold drawings to export.
var sourceFolders = []string{"sprites", "tiles"}

// partCategories are the sprites/ categories whose drawings are split into parts.
var partCategories = map[string]bool{"critters": true, "npcs": true, "enemies": true}

// Run exports every SVG under SrcDir/sprites and SrcDir/tiles that changed
// since its last export (or all of them with Force). Drawings in
// sprites/critters, sprites/npcs and sprites/enemies become one PNG per part
// plus a part JSON in a folder named after the drawing; every other drawing
// becomes one PNG of its page at the mirrored path. It stops at the first
// drawing that fails and returns an error naming its SVG.
func Run(o Options) error {
	if o.SrcDir == "" {
		o.SrcDir = "assets-src"
	}
	if o.OutDir == "" {
		o.OutDir = "assets"
	}
	if o.Log == nil {
		o.Log = io.Discard
	}
	if o.Inkscape == nil {
		return errors.New("artexport: no Inkscape runner given")
	}

	sources, err := findSources(o.SrcDir)
	if err != nil {
		return err
	}

	exported := 0
	for _, rel := range sources {
		svg := filepath.Join(o.SrcDir, filepath.FromSlash(rel))
		did, err := exportDrawing(o, svg, rel)
		if err != nil {
			return fmt.Errorf("%s: %w", svg, err)
		}
		if did {
			exported++
		}
	}
	if exported == 0 {
		fmt.Fprintln(o.Log, "art: nothing to export")
	}
	return nil
}

// findSources returns the slash-separated paths, relative to srcDir, of every
// SVG under srcDir/sprites and srcDir/tiles, in sorted order. A missing
// folder has no sources.
func findSources(srcDir string) ([]string, error) {
	var sources []string
	for _, folder := range sourceFolders {
		root := filepath.Join(srcDir, folder)
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				if p == root && errors.Is(err, fs.ErrNotExist) {
					return filepath.SkipDir
				}
				return err
			}
			if d.IsDir() || !strings.EqualFold(filepath.Ext(p), ".svg") {
				return nil
			}
			rel, err := filepath.Rel(srcDir, p)
			if err != nil {
				return err
			}
			sources = append(sources, filepath.ToSlash(rel))
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return sources, nil
}

// exportDrawing exports one source drawing at slash path rel under SrcDir and
// reports whether it did any work (false when the outputs were up to date).
func exportDrawing(o Options, svg, rel string) (bool, error) {
	base := strings.TrimSuffix(rel, path.Ext(rel))
	segments := strings.Split(rel, "/")
	if segments[0] == "sprites" && len(segments) >= 3 && partCategories[segments[1]] {
		dir := filepath.Join(o.OutDir, filepath.FromSlash(base))
		name := path.Base(base)
		jsonPath := filepath.Join(dir, name+".json")
		if !o.Force && upToDate(svg, jsonPath) {
			return false, nil
		}
		return true, exportParts(o, svg, dir, jsonPath)
	}

	png := filepath.Join(o.OutDir, filepath.FromSlash(base)+".png")
	if !o.Force && upToDate(svg, png) {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(png), 0o755); err != nil {
		return true, err
	}
	if err := o.Inkscape.ExportPage(svg, png); err != nil {
		return true, err
	}
	fmt.Fprintf(o.Log, "exported %s\n", svg)
	return true, nil
}

// exportParts exports each part of a character drawing into dir, deletes
// PNGs in dir that are no longer parts, and writes the part JSON last so a
// failed export is retried on the next run.
func exportParts(o Options, svg, dir, jsonPath string) error {
	out, err := o.Inkscape.QueryAll(svg)
	if err != nil {
		return err
	}
	boxes, err := ParseQueryAll(out)
	if err != nil {
		return err
	}
	sheet, err := BuildSheet(boxes)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	keep := make(map[string]bool)
	for _, v := range viewOrder {
		for _, p := range sheet.Views[v].Parts {
			keep[p.Image] = true
			id := strings.TrimSuffix(p.Image, ".png")
			if err := o.Inkscape.ExportID(svg, id, filepath.Join(dir, p.Image)); err != nil {
				return err
			}
		}
	}

	if err := removeStalePNGs(dir, keep); err != nil {
		return err
	}

	data, err := json.MarshalIndent(sheet, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, append(data, '\n'), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(o.Log, "exported %s: %d parts\n", svg, len(keep))
	return nil
}

// removeStalePNGs deletes the *.png files directly in dir whose names are not in keep.
func removeStalePNGs(dir string, keep map[string]bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".png" || keep[e.Name()] {
			continue
		}
		if err := os.Remove(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

// upToDate reports whether stamp exists and is not older than src.
func upToDate(src, stamp string) bool {
	s, err := os.Stat(stamp)
	if err != nil {
		return false
	}
	in, err := os.Stat(src)
	if err != nil {
		return false
	}
	return !s.ModTime().Before(in.ModTime())
}
