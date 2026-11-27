package synth

import (
	"encoding/binary"
	"io"
	"math"
)

// WriteWAV writes samples as a mono, SampleRate Hz, 16-bit PCM WAV file.
// Samples outside -1..1 are clamped.
func WriteWAV(w io.Writer, samples []float64) error {
	const (
		channels      = 1
		bitsPerSample = 16
		blockAlign    = channels * bitsPerSample / 8
	)
	dataLen := len(samples) * blockAlign

	buf := make([]byte, 44+dataLen)
	copy(buf[0:], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:], uint32(36+dataLen))
	copy(buf[8:], "WAVE")
	copy(buf[12:], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:], 16) // fmt chunk size
	binary.LittleEndian.PutUint16(buf[20:], 1)  // PCM
	binary.LittleEndian.PutUint16(buf[22:], channels)
	binary.LittleEndian.PutUint32(buf[24:], SampleRate)
	binary.LittleEndian.PutUint32(buf[28:], SampleRate*blockAlign)
	binary.LittleEndian.PutUint16(buf[32:], blockAlign)
	binary.LittleEndian.PutUint16(buf[34:], bitsPerSample)
	copy(buf[36:], "data")
	binary.LittleEndian.PutUint32(buf[40:], uint32(dataLen))

	for i, s := range samples {
		s = math.Max(-1, math.Min(1, s))
		v := int16(math.Round(s * math.MaxInt16))
		binary.LittleEndian.PutUint16(buf[44+i*blockAlign:], uint16(v))
	}
	_, err := w.Write(buf)
	return err
}
