package render

import (
	"bytes"
	"image/png"
	"strings"
	"testing"

	"github.com/jacoelho/gophershot/internal/doc"
)

func TestRenderPNGProducesImage(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	input := doc.FromSource([]byte("package main\nfunc main() { println(\"hi\") }\n"))
	err := ToPNG(input, &out, Options{})
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}
	if img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		t.Fatalf("invalid image bounds: %v", img.Bounds())
	}
}

func TestRenderPNGDeterministicBounds(t *testing.T) {
	t.Parallel()

	input := doc.FromSource([]byte("package main\n\nfunc main() {\n\tprintln(42)\n}\n"))
	opts := Options{FontSize: 14, LineHeight: 1.3, Padding: 16}.WithLineNumbers(false)

	var one bytes.Buffer
	if err := ToPNG(input, &one, opts); err != nil {
		t.Fatalf("render first: %v", err)
	}
	imgOne, err := png.Decode(bytes.NewReader(one.Bytes()))
	if err != nil {
		t.Fatalf("decode first: %v", err)
	}

	var two bytes.Buffer
	if err := ToPNG(input, &two, opts); err != nil {
		t.Fatalf("render second: %v", err)
	}
	imgTwo, err := png.Decode(bytes.NewReader(two.Bytes()))
	if err != nil {
		t.Fatalf("decode second: %v", err)
	}

	if imgOne.Bounds() != imgTwo.Bounds() {
		t.Fatalf("bounds mismatch: %v vs %v", imgOne.Bounds(), imgTwo.Bounds())
	}
}

func TestRenderLineNumbersIncreaseWidth(t *testing.T) {
	t.Parallel()

	input := doc.FromSource([]byte("package main\nfunc main() {}\n"))

	var withoutNumbers bytes.Buffer
	if err := ToPNG(input, &withoutNumbers, Options{}.WithLineNumbers(false)); err != nil {
		t.Fatalf("render without numbers: %v", err)
	}
	imgWithout, err := png.Decode(bytes.NewReader(withoutNumbers.Bytes()))
	if err != nil {
		t.Fatalf("decode without numbers: %v", err)
	}

	var withNumbers bytes.Buffer
	if err := ToPNG(input, &withNumbers, Options{}.WithLineNumbers(true)); err != nil {
		t.Fatalf("render with numbers: %v", err)
	}
	imgWith, err := png.Decode(bytes.NewReader(withNumbers.Bytes()))
	if err != nil {
		t.Fatalf("decode with numbers: %v", err)
	}

	if imgWith.Bounds().Dx() <= imgWithout.Bounds().Dx() {
		t.Fatalf("expected line numbers image wider, with=%d without=%d", imgWith.Bounds().Dx(), imgWithout.Bounds().Dx())
	}
}

func TestRenderFontSizeIncreaseHeight(t *testing.T) {
	t.Parallel()

	input := doc.FromSource([]byte("package main\nfunc main() {}\n"))

	var small bytes.Buffer
	if err := ToPNG(input, &small, Options{FontSize: 12}.WithLineNumbers(false)); err != nil {
		t.Fatalf("render small: %v", err)
	}
	imgSmall, err := png.Decode(bytes.NewReader(small.Bytes()))
	if err != nil {
		t.Fatalf("decode small: %v", err)
	}

	var large bytes.Buffer
	if err := ToPNG(input, &large, Options{FontSize: 24}.WithLineNumbers(false)); err != nil {
		t.Fatalf("render large: %v", err)
	}
	imgLarge, err := png.Decode(bytes.NewReader(large.Bytes()))
	if err != nil {
		t.Fatalf("decode large: %v", err)
	}

	if imgLarge.Bounds().Dy() <= imgSmall.Bounds().Dy() {
		t.Fatalf("expected larger font image taller, large=%d small=%d", imgLarge.Bounds().Dy(), imgSmall.Bounds().Dy())
	}
}

func TestFormatLineLabel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		origins []int
		want    string
	}{
		{origins: []int{7}, want: "7"},
		{origins: []int{10, 11, 12}, want: "10-12"},
		{origins: []int{5, 7, 9}, want: "5,7,9"},
		{origins: []int{3, 3, 2, 1}, want: "1-3"},
	}

	for _, tc := range cases {
		if got := formatLineLabel(tc.origins); got != tc.want {
			t.Fatalf("formatLineLabel(%v) = %q, want %q", tc.origins, got, tc.want)
		}
	}
}

func TestNormalizeRenderableTextReplacesTabAndOddSpaces(t *testing.T) {
	t.Parallel()

	in := "\tfoo\u00A0bar\v"
	got := normalizeRenderableText(in)

	if strings.ContainsRune(got, '\t') {
		t.Fatalf("expected tab to be replaced, got %q", got)
	}
	if strings.ContainsRune(got, '\u00A0') {
		t.Fatalf("expected non-breaking space to be replaced, got %q", got)
	}
	if strings.ContainsRune(got, '\v') {
		t.Fatalf("expected control char to be replaced, got %q", got)
	}
	if !strings.Contains(got, "    foo bar ") {
		t.Fatalf("unexpected normalization output %q", got)
	}
}

func TestTokenizeRemovesTabsFromSegments(t *testing.T) {
	t.Parallel()

	lines, err := tokenize("package main\nfunc main() {\n\tprintln(42)\n}\n")
	if err != nil {
		t.Fatalf("tokenize: %v", err)
	}
	for _, line := range lines {
		for _, seg := range line {
			if strings.ContainsRune(seg.text, '\t') {
				t.Fatalf("segment still has tab: %q", seg.text)
			}
		}
	}
}
