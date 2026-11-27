package web_test

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"net/url"
	"os"
	"strings"
	"testing"
)

// manifest holds the fields of manifest.webmanifest the tests check.
type manifest struct {
	StartURL string         `json:"start_url"`
	Scope    string         `json:"scope"`
	Icons    []manifestIcon `json:"icons"`
}

type manifestIcon struct {
	Src     string `json:"src"`
	Sizes   string `json:"sizes"`
	Type    string `json:"type"`
	Purpose string `json:"purpose"`
}

func loadManifest(t *testing.T) manifest {
	t.Helper()
	data, err := os.ReadFile("manifest.webmanifest")
	if err != nil {
		t.Fatalf("read manifest.webmanifest: %v", err)
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse manifest.webmanifest: %v", err)
	}
	return m
}

func TestManifest_IconsExistAtDeclaredSizesAndMeetInstallCriteria(t *testing.T) {
	m := loadManifest(t)
	if len(m.Icons) == 0 {
		t.Fatal("manifest.webmanifest lists no icons")
	}

	var has192, has512, hasMaskable bool
	for _, icon := range m.Icons {
		t.Run(icon.Src, func(t *testing.T) {
			if icon.Type != "image/png" {
				t.Errorf("icon %q type = %q, want \"image/png\"", icon.Src, icon.Type)
			}
			f, err := os.Open(icon.Src)
			if err != nil {
				t.Fatalf("open icon %q: %v", icon.Src, err)
			}
			defer f.Close()

			cfg, format, err := image.DecodeConfig(f)
			if err != nil {
				t.Fatalf("decode icon %q: %v", icon.Src, err)
			}
			if format != "png" {
				t.Errorf("icon %q format = %q, want \"png\"", icon.Src, format)
			}
			gotSizes := fmt.Sprintf("%dx%d", cfg.Width, cfg.Height)
			if gotSizes != icon.Sizes {
				t.Errorf("icon %q pixel size = %s, want declared sizes %q", icon.Src, gotSizes, icon.Sizes)
			}
		})

		purpose := icon.Purpose
		if purpose == "" {
			purpose = "any"
		}
		if icon.Sizes == "192x192" && strings.Contains(purpose, "any") {
			has192 = true
		}
		if icon.Sizes == "512x512" && strings.Contains(purpose, "any") {
			has512 = true
		}
		if strings.Contains(purpose, "maskable") {
			hasMaskable = true
		}
	}

	if !has192 {
		t.Error("manifest.webmanifest has no 192x192 icon with purpose \"any\"")
	}
	if !has512 {
		t.Error("manifest.webmanifest has no 512x512 icon with purpose \"any\"")
	}
	if !hasMaskable {
		t.Error("manifest.webmanifest has no icon with purpose \"maskable\"")
	}

	t.Run("icons/apple-touch-icon.png", func(t *testing.T) {
		f, err := os.Open("icons/apple-touch-icon.png")
		if err != nil {
			t.Fatalf("open icons/apple-touch-icon.png: %v", err)
		}
		defer f.Close()

		img, format, err := image.Decode(f)
		if err != nil {
			t.Fatalf("decode icons/apple-touch-icon.png: %v", err)
		}
		if format != "png" {
			t.Errorf("apple-touch-icon format = %q, want \"png\"", format)
		}
		b := img.Bounds()
		if b.Dx() != 180 || b.Dy() != 180 {
			t.Errorf("apple-touch-icon size = %dx%d, want 180x180", b.Dx(), b.Dy())
		}
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				_, _, _, a := img.At(x, y).RGBA()
				if a != 0xffff {
					t.Fatalf("apple-touch-icon pixel (%d,%d) alpha = %#x, want 0xffff (fully opaque)", x, y, a)
				}
			}
		}
	})
}

func TestManifest_URLsAreRelativeForSubPathDeployment(t *testing.T) {
	m := loadManifest(t)

	type urlCase struct {
		name  string
		value string
	}
	cases := []urlCase{
		{"start_url", m.StartURL},
		{"scope", m.Scope},
	}
	for _, icon := range m.Icons {
		cases = append(cases, urlCase{"icon " + icon.Src, icon.Src})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.value == "" {
				t.Fatalf("%s is empty", tc.name)
			}
			if strings.HasPrefix(tc.value, "/") {
				t.Errorf("%s = %q starts with '/'; want relative URL", tc.name, tc.value)
			}
			u, err := url.Parse(tc.value)
			if err != nil {
				t.Fatalf("parse %s %q: %v", tc.name, tc.value, err)
			}
			if u.Scheme != "" || u.Host != "" {
				t.Errorf("%s = %q has scheme/host; want relative URL", tc.name, tc.value)
			}
		})
	}
}
