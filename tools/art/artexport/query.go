package artexport

import (
	"fmt"
	"strconv"
	"strings"
)

// Box is one line from Inkscape's --query-all output: an element id and its
// bounding box in drawing pixels (origin at top-left, y down).
type Box struct {
	ID         string
	X, Y, W, H float64
}

// ParseQueryAll parses the stdout of `inkscape --query-all`. Blank lines are
// ignored. Any non-blank line without five comma-separated fields or with a
// non-numeric coordinate returns an error quoting the line.
func ParseQueryAll(out []byte) ([]Box, error) {
	var boxes []Box
	for _, raw := range strings.Split(string(out), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) != 5 || fields[0] == "" {
			return nil, fmt.Errorf("malformed query line %q: expected id,x,y,width,height", line)
		}
		vals := make([]float64, 4)
		for i := range 4 {
			v, err := strconv.ParseFloat(fields[i+1], 64)
			if err != nil {
				return nil, fmt.Errorf("malformed query line %q: %w", line, err)
			}
			vals[i] = v
		}
		boxes = append(boxes, Box{
			ID: fields[0],
			X:  vals[0],
			Y:  vals[1],
			W:  vals[2],
			H:  vals[3],
		})
	}
	return boxes, nil
}
