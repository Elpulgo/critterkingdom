package artexport

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

const pxPerWorldPixel = 8

var partIDPattern = regexp.MustCompile(`^(down|up|side)_[a-z0-9_]+$`)

// Sheet describes the exported parts of a character drawing across views.
type Sheet struct {
	PxPerWorldPixel int             `json:"pxPerWorldPixel"`
	Views           map[string]View `json:"views"`
}

// View holds the ordered parts of a single facing direction ("down", "up" or "side").
type View struct {
	Parts []Part `json:"parts"`
}

// Part describes one character part in feet-relative world pixels.
type Part struct {
	Name   string  `json:"name"`
	Image  string  `json:"image"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	W      float64 `json:"w"`
	H      float64 `json:"h"`
	PivotX float64 `json:"pivotX"`
	PivotY float64 `json:"pivotY"`
}

// BuildSheet derives the part sheet from parsed --query-all boxes.
func BuildSheet(boxes []Box) (Sheet, error) {
	type viewAccum struct {
		feet     *Box
		parts    []Box
		partSet  map[string]bool
		pivots   map[string]Box
		pivotIDs []string
	}

	viewOrder := []string{"down", "up", "side"}
	accum := make(map[string]*viewAccum, len(viewOrder))
	getView := func(v string) *viewAccum {
		a, ok := accum[v]
		if !ok {
			a = &viewAccum{
				partSet: make(map[string]bool),
				pivots:  make(map[string]Box),
			}
			accum[v] = a
		}
		return a
	}

	for _, b := range boxes {
		if !partIDPattern.MatchString(b.ID) {
			continue
		}
		view, rest, _ := strings.Cut(b.ID, "_")
		a := getView(view)
		switch {
		case rest == "feet":
			box := b
			a.feet = &box
		case strings.HasSuffix(rest, "_pivot"):
			target := strings.TrimSuffix(b.ID, "_pivot")
			a.pivots[target] = b
			a.pivotIDs = append(a.pivotIDs, b.ID)
		default:
			a.parts = append(a.parts, b)
			a.partSet[b.ID] = true
		}
	}

	views := make(map[string]View)
	totalParts := 0

	for _, v := range viewOrder {
		a, ok := accum[v]
		if !ok {
			continue
		}
		for _, pid := range a.pivotIDs {
			target := strings.TrimSuffix(pid, "_pivot")
			if !a.partSet[target] {
				return Sheet{}, fmt.Errorf("pivot %q has no matching part %q", pid, target)
			}
		}
		if len(a.parts) == 0 {
			continue
		}
		if a.feet == nil {
			return Sheet{}, fmt.Errorf("view %q has parts (e.g. %q) but no %q anchor", v, a.parts[0].ID, v+"_feet")
		}

		feetX := a.feet.X + a.feet.W/2
		feetY := a.feet.Y + a.feet.H/2

		parts := make([]Part, 0, len(a.parts))
		for _, pb := range a.parts {
			x0 := math.Floor(pb.X)
			y0 := math.Floor(pb.Y)
			x1 := math.Ceil(pb.X + pb.W)
			y1 := math.Ceil(pb.Y + pb.H)

			var px, py float64
			if piv, ok := a.pivots[pb.ID]; ok {
				px = piv.X + piv.W/2
				py = piv.Y + piv.H/2
			} else {
				px = (x0 + x1) / 2
				py = (y0 + y1) / 2
			}

			_, name, _ := strings.Cut(pb.ID, "_")
			parts = append(parts, Part{
				Name:   name,
				Image:  pb.ID + ".png",
				X:      round4((x0 - feetX) / pxPerWorldPixel),
				Y:      round4((y0 - feetY) / pxPerWorldPixel),
				W:      round4((x1 - x0) / pxPerWorldPixel),
				H:      round4((y1 - y0) / pxPerWorldPixel),
				PivotX: round4((px - feetX) / pxPerWorldPixel),
				PivotY: round4((py - feetY) / pxPerWorldPixel),
			})
		}

		views[v] = View{Parts: parts}
		totalParts += len(parts)
	}

	if totalParts == 0 {
		return Sheet{}, fmt.Errorf("drawing has no parts matching <view>_<part>")
	}

	return Sheet{
		PxPerWorldPixel: pxPerWorldPixel,
		Views:           views,
	}, nil
}

func round4(v float64) float64 {
	r := math.Round(v*1e4) / 1e4
	if r == 0 {
		return 0
	}
	return r
}
