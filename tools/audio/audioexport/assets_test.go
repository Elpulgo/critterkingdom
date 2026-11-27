package audioexport_test

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/Elpulgo/critterkingdom/tools/audio/synth"
)

// committedAudio is the folder `make audio` writes to, seen from this
// package's folder.
var committedAudio = filepath.Join("..", "..", "..", "assets", "audio")

// audioBudget is the most all of assets/audio/ may weigh, in bytes.
const audioBudget = 3_000_000

// lengthLimits are the shortest and longest a committed sound of each kind
// may be, in seconds.
var lengthLimits = map[synth.Kind][2]float64{
	synth.SFX:   {0, 2.5},
	synth.Music: {45, 60},
	synth.Voice: {0, 0.3},
}

// TestCommittedAudio_MatchesCatalogIsMonoVorbisAndKeepsToLimits checks the
// OGG files in assets/audio/ against the sound catalog: one file per entry
// and nothing else, each mono Ogg Vorbis at 44100 Hz, within the length
// limits of its kind, and all of them within the size budget.
func TestCommittedAudio_MatchesCatalogIsMonoVorbisAndKeepsToLimits(t *testing.T) {
	kinds := map[string]synth.Kind{}
	for _, s := range synth.Catalog() {
		kinds[s.Path] = s.Kind
	}

	files := map[string]int64{}
	var total int64
	err := filepath.WalkDir(committedAudio, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(committedAudio, p)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = info.Size()
		total += info.Size()
		return nil
	})
	if err != nil {
		t.Fatalf("reading %s: %v", committedAudio, err)
	}

	for _, p := range sortedKeys(kinds) {
		if _, ok := files[p]; !ok {
			t.Errorf("%s is in the catalog but not in assets/audio/; run make audio", p)
		}
	}
	for _, p := range sortedKeys(files) {
		if _, ok := kinds[p]; !ok {
			t.Errorf("assets/audio/%s is not in the catalog; run make audio or remove it", p)
		}
	}
	if total > audioBudget {
		t.Errorf("assets/audio/ is %d bytes, over the %d byte budget", total, audioBudget)
	}

	for _, p := range sortedKeys(files) {
		kind, ok := kinds[p]
		if !ok {
			continue
		}
		t.Run(p, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(committedAudio, filepath.FromSlash(p)))
			if err != nil {
				t.Fatal(err)
			}
			info, err := readOggVorbis(data)
			if err != nil {
				t.Fatalf("not Ogg Vorbis: %v", err)
			}
			if info.channels != 1 || info.rate != synth.SampleRate {
				t.Errorf("got %d channels at %d Hz, want mono at %d Hz", info.channels, info.rate, synth.SampleRate)
			}
			secs := float64(info.samples) / float64(info.rate)
			limits := lengthLimits[kind]
			if secs < limits[0] || secs > limits[1] {
				t.Errorf("%s lasts %.3f s, want %g to %g s", kind, secs, limits[0], limits[1])
			}
		})
	}
}

// oggVorbis is what the lint test reads from an Ogg Vorbis file's pages.
type oggVorbis struct {
	channels int
	rate     int
	// samples is the granule position of the last page: the number of
	// samples per channel in the stream.
	samples int64
}

// readOggVorbis walks every Ogg page in data. The first page must carry a
// Vorbis identification header, which gives the channels and sample rate;
// the last page's granule position gives the length.
func readOggVorbis(data []byte) (oggVorbis, error) {
	var v oggVorbis
	pages := 0
	for off := 0; off < len(data); pages++ {
		if len(data)-off < 27 || string(data[off:off+4]) != "OggS" {
			return v, fmt.Errorf("no Ogg page at byte %d", off)
		}
		granule := int64(binary.LittleEndian.Uint64(data[off+6:]))
		segments := int(data[off+26])
		start := off + 27 + segments
		if start > len(data) {
			return v, fmt.Errorf("page at byte %d is cut short", off)
		}
		end := start
		for _, n := range data[off+27 : start] {
			end += int(n)
		}
		if end > len(data) {
			return v, fmt.Errorf("page at byte %d is cut short", off)
		}
		if pages == 0 {
			id := data[start:end]
			if len(id) < 16 || id[0] != 1 || string(id[1:7]) != "vorbis" {
				return v, errors.New("first page has no Vorbis identification header")
			}
			v.channels = int(id[11])
			v.rate = int(binary.LittleEndian.Uint32(id[12:16]))
			if v.rate == 0 {
				return v, errors.New("sample rate is 0")
			}
		}
		if granule != -1 { // -1: no packet ends on this page
			v.samples = granule
		}
		off = end
	}
	if pages == 0 {
		return v, errors.New("file is empty")
	}
	return v, nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
