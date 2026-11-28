// Package anim builds characters from parts and moves them with shared
// motions that are pure functions of time. It does not draw: the world scene
// asks it where each part sits and draws the part there. It must not import
// Ebitengine, so everything in it is unit tested.
package anim

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// The view names a part sheet uses. Down is the only view every sheet must
// have; a missing up or side view shows the down view instead.
const (
	ViewDown = "down"
	ViewUp   = "up"
	ViewSide = "side"
)

// Part is one piece of a character in one view, in world pixels measured from
// the character's feet (Y grows downwards, so parts above the feet have
// negative Y).
type Part struct {
	// Name is the part's name without the view, for example "head".
	Name string
	// Image is the part's PNG file name in the sheet's folder.
	Image string
	// Box is where the part's image sits.
	Box geom.Rect
	// Pivot is the point the part turns and squashes about.
	Pivot geom.Vec
}

// View is one facing of a character: its parts in drawing order, back to
// front.
type View struct {
	Name  string
	Parts []Part
}

// Layout is a character's parts across its views, as read from the part JSON
// that `make art` writes.
type Layout struct {
	// PxPerWorldPixel is how many drawing pixels make one world pixel.
	PxPerWorldPixel int
	views           map[string]View
}

// sheetJSON, viewJSON and partJSON decode the part JSON. They mirror
// tools/art/artexport's Sheet, which the game must not import.
type sheetJSON struct {
	PxPerWorldPixel int                 `json:"pxPerWorldPixel"`
	Views           map[string]viewJSON `json:"views"`
}

type viewJSON struct {
	Parts []partJSON `json:"parts"`
}

type partJSON struct {
	Name   string  `json:"name"`
	Image  string  `json:"image"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	W      float64 `json:"w"`
	H      float64 `json:"h"`
	PivotX float64 `json:"pivotX"`
	PivotY float64 `json:"pivotY"`
}

// ParseLayout reads a part sheet in the format `make art` writes. It returns
// an error for malformed JSON, for a sheet without a down view and for a view
// that names the same part twice.
func ParseLayout(data []byte) (Layout, error) {
	var s sheetJSON
	if err := json.Unmarshal(data, &s); err != nil {
		return Layout{}, fmt.Errorf("anim: part sheet: %w", err)
	}
	if _, ok := s.Views[ViewDown]; !ok {
		return Layout{}, errors.New("anim: part sheet has no down view")
	}

	names := make([]string, 0, len(s.Views))
	for name := range s.Views {
		names = append(names, name)
	}
	sort.Strings(names) // report the same error every time

	views := make(map[string]View, len(s.Views))
	for _, name := range names {
		v := View{Name: name, Parts: make([]Part, 0, len(s.Views[name].Parts))}
		seen := make(map[string]bool, len(s.Views[name].Parts))
		for _, p := range s.Views[name].Parts {
			if seen[p.Name] {
				return Layout{}, fmt.Errorf("anim: part sheet view %q has two parts named %q", name, p.Name)
			}
			seen[p.Name] = true
			v.Parts = append(v.Parts, Part{
				Name:  p.Name,
				Image: p.Image,
				Box: geom.Rect{
					Min: geom.Vec{X: p.X, Y: p.Y},
					Max: geom.Vec{X: p.X + p.W, Y: p.Y + p.H},
				},
				Pivot: geom.Vec{X: p.PivotX, Y: p.PivotY},
			})
		}
		views[name] = v
	}
	return Layout{PxPerWorldPixel: s.PxPerWorldPixel, views: views}, nil
}

// View returns the named view. A view the sheet lacks, such as the up or side
// view of a character drawn only from the front, falls back to the down view;
// the returned View's Name is then "down".
func (l Layout) View(name string) View {
	if v, ok := l.views[name]; ok {
		return v
	}
	return l.views[ViewDown]
}
