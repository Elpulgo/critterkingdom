package ui

import "strings"

// LinesPerPage is the most lines the dialogue box shows at once.
const LinesPerPage = 2

// Paginate wraps text into lines no wider than width, as measured by
// measure, and groups them into pages of at most LinesPerPage lines. Lines
// break only at spaces; runs of spaces collapse to one. A word wider than
// width on its own is kept whole on a line of its own. A text with no words
// gives no pages.
func Paginate(text string, width float64, measure func(string) float64) [][]string {
	var lines []string
	line := ""
	for _, word := range strings.Split(text, " ") {
		if word == "" {
			continue
		}
		if line == "" {
			line = word
			continue
		}
		if next := line + " " + word; measure(next) <= width {
			line = next
			continue
		}
		lines = append(lines, line)
		line = word
	}
	if line != "" {
		lines = append(lines, line)
	}

	var pages [][]string
	for start := 0; start < len(lines); start += LinesPerPage {
		pages = append(pages, lines[start:min(start+LinesPerPage, len(lines))])
	}
	return pages
}
