package ui

import (
	"reflect"
	"testing"
	"unicode/utf8"
)

func TestPaginate_WrapsAtSpacesIntoPagesOfAtMostTwoLines(t *testing.T) {
	// Every character is 10 px wide and the box is 100 px, so a line holds
	// 10 characters.
	measure := func(s string) float64 { return 10 * float64(utf8.RuneCountInString(s)) }
	const width = 100

	cases := []struct {
		name string
		text string
		want [][]string
	}{
		{
			name: "three lines become a two-line page and a one-line page",
			text: "one two three four five six",
			want: [][]string{{"one two", "three four"}, {"five six"}},
		},
		{
			name: "a short text is one page of one line",
			text: "hi there",
			want: [][]string{{"hi there"}},
		},
		{
			name: "a word wider than the box gets its own line",
			text: "a grandmotherly hug",
			want: [][]string{{"a", "grandmotherly"}, {"hug"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Paginate(tc.text, width, measure); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Paginate(%q) = %q, want %q", tc.text, got, tc.want)
			}
		})
	}
}
