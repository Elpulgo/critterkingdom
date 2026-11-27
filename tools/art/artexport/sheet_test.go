package artexport_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/tools/art/artexport"
)

func TestBuildSheet_WorkedExample(t *testing.T) {
	query := strings.Join([]string{
		"svg1,0,0,256,256",
		"layer1,0,0,256,256",
		"down_feet,124,236,8,8",
		"down_body,80.5,150.2,95,90.3",
		"down_head,64,40,128,120",
		"down_head_pivot,126,146,4,4",
		"",
	}, "\n")

	boxes, err := artexport.ParseQueryAll([]byte(query))
	if err != nil {
		t.Fatalf("ParseQueryAll: %v", err)
	}

	got, err := artexport.BuildSheet(boxes)
	if err != nil {
		t.Fatalf("BuildSheet: %v", err)
	}

	want := artexport.Sheet{
		PxPerWorldPixel: 8,
		Views: map[string]artexport.View{
			"down": {
				Parts: []artexport.Part{
					{
						Name:   "body",
						Image:  "down_body.png",
						X:      -6,
						Y:      -11.25,
						W:      12,
						H:      11.375,
						PivotX: 0,
						PivotY: -5.5625,
					},
					{
						Name:   "head",
						Image:  "down_head.png",
						X:      -8,
						Y:      -25,
						W:      16,
						H:      15,
						PivotX: 0,
						PivotY: -11.5,
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildSheet =\n%+v\nwant\n%+v", got, want)
	}
}

func TestBuildSheet_SnapsFractionalBoxesOutwardToWholePixelsBeforeDividingByEight(t *testing.T) {
	// Feet at (100, 100). Part box x=92.9, y=84.1, w=14.2 (right 107.1), h=15.8 (bottom 99.9).
	// Snapped outward: x0=92, y0=84, x1=108, y1=100 -> w=16px (2 world px), h=16px (2 world px).
	// Unmarked pivot is centre of snapped box: (100, 92) -> pivotX=0, pivotY=-1.
	query := strings.Join([]string{
		"up_feet,96,96,8,8",
		"up_tail,92.9,84.1,14.2,15.8",
	}, "\n")

	boxes, err := artexport.ParseQueryAll([]byte(query))
	if err != nil {
		t.Fatalf("ParseQueryAll: %v", err)
	}

	got, err := artexport.BuildSheet(boxes)
	if err != nil {
		t.Fatalf("BuildSheet: %v", err)
	}

	want := artexport.Part{
		Name:   "tail",
		Image:  "up_tail.png",
		X:      -1,
		Y:      -2,
		W:      2,
		H:      2,
		PivotX: 0,
		PivotY: -1,
	}

	parts := got.Views["up"].Parts
	if len(parts) != 1 || parts[0] != want {
		t.Errorf("up parts = %+v, want [%+v]", parts, want)
	}
}

func TestBuildSheet_PairsPartNameWithUnderscoresWithItsPivotAndIgnoresOtherIDs(t *testing.T) {
	query := strings.Join([]string{
		"svg5,0,0,256,256",
		"path12,10,10,20,20",
		"front_body,10,10,20,20",
		"down_Body,10,10,20,20",
		"down_feet,124,236,8,8",
		"down_ear_left,88,24,16,32",
		"down_ear_left_pivot,94,52,4,4",
	}, "\n")

	boxes, err := artexport.ParseQueryAll([]byte(query))
	if err != nil {
		t.Fatalf("ParseQueryAll: %v", err)
	}

	got, err := artexport.BuildSheet(boxes)
	if err != nil {
		t.Fatalf("BuildSheet: %v", err)
	}

	want := artexport.Part{
		Name:   "ear_left",
		Image:  "down_ear_left.png",
		X:      -5,
		Y:      -27,
		W:      2,
		H:      4,
		PivotX: -4,
		PivotY: -23.25,
	}

	if len(got.Views) != 1 {
		t.Fatalf("views = %v, want only \"down\"", got.Views)
	}
	parts := got.Views["down"].Parts
	if len(parts) != 1 || parts[0] != want {
		t.Errorf("down parts = %+v, want [%+v]", parts, want)
	}
}

func TestParseAndBuildSheet_RejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name      string
		query     string
		wantInErr string
	}{
		{
			name:      "malformed line with too few fields",
			query:     "down_body,10,20,30",
			wantInErr: "down_body,10,20,30",
		},
		{
			name:      "malformed line with non-numeric coordinate",
			query:     "down_body,10,abc,30,40",
			wantInErr: "down_body,10,abc,30,40",
		},
		{
			name: "view has parts but no feet anchor",
			query: strings.Join([]string{
				"side_body,80,150,96,88",
			}, "\n"),
			wantInErr: "side_feet",
		},
		{
			name: "pivot with no matching part",
			query: strings.Join([]string{
				"down_feet,124,236,8,8",
				"down_body,80,150,96,88",
				"down_wing_pivot,100,100,4,4",
			}, "\n"),
			wantInErr: "down_wing_pivot",
		},
		{
			name: "pivot in otherwise empty view",
			query: strings.Join([]string{
				"down_feet,124,236,8,8",
				"down_body,80,150,96,88",
				"up_head_pivot,100,100,4,4",
			}, "\n"),
			wantInErr: "up_head_pivot",
		},
		{
			name: "no matching parts at all",
			query: strings.Join([]string{
				"svg1,0,0,256,256",
				"layer1,0,0,256,256",
				"down_feet,124,236,8,8",
			}, "\n"),
			wantInErr: "no parts",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			boxes, err := artexport.ParseQueryAll([]byte(tc.query))
			if err == nil {
				_, err = artexport.BuildSheet(boxes)
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantInErr)
			}
			if !strings.Contains(err.Error(), tc.wantInErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tc.wantInErr)
			}
		})
	}
}
