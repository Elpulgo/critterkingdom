package synth

import (
	"bytes"
	"encoding/binary"
	"slices"
	"testing"
)

func TestWriteWAV_WritesMono44100Hz16BitPCM(t *testing.T) {
	// 0.5 * 32767 = 16383.5 rounds away from zero; 2 and -2 are clamped.
	samples := []float64{0, 0.5, -0.5, 1, -1, 2, -2}
	wantPCM := []int16{0, 16384, -16384, 32767, -32767, 32767, -32767}

	var buf bytes.Buffer
	if err := WriteWAV(&buf, samples); err != nil {
		t.Fatal(err)
	}
	b := buf.Bytes()
	if len(b) != 44+14 {
		t.Fatalf("file is %d bytes, want 58 (44 header + 7 samples * 2)", len(b))
	}

	le := binary.LittleEndian
	checks := []struct {
		field     string
		got, want any
	}{
		{"RIFF tag", string(b[0:4]), "RIFF"},
		{"RIFF size", le.Uint32(b[4:]), uint32(50)},
		{"WAVE tag", string(b[8:12]), "WAVE"},
		{"fmt tag", string(b[12:16]), "fmt "},
		{"fmt size", le.Uint32(b[16:]), uint32(16)},
		{"format (1 = PCM)", le.Uint16(b[20:]), uint16(1)},
		{"channels", le.Uint16(b[22:]), uint16(1)},
		{"sample rate", le.Uint32(b[24:]), uint32(44100)},
		{"byte rate", le.Uint32(b[28:]), uint32(88200)},
		{"block align", le.Uint16(b[32:]), uint16(2)},
		{"bits per sample", le.Uint16(b[34:]), uint16(16)},
		{"data tag", string(b[36:40]), "data"},
		{"data size", le.Uint32(b[40:]), uint32(14)},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.field, c.got, c.want)
		}
	}

	gotPCM := make([]int16, len(samples))
	for i := range gotPCM {
		gotPCM[i] = int16(le.Uint16(b[44+2*i:]))
	}
	if !slices.Equal(gotPCM, wantPCM) {
		t.Errorf("PCM samples = %v, want %v", gotPCM, wantPCM)
	}
}
