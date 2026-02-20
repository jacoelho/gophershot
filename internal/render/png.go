package render

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/jacoelho/gophershot/internal/doc"
	"github.com/jacoelho/gophershot/internal/lang"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed JetBrainsMono-Regular.ttf
var jetBrainsMonoTTF []byte

var (
	backgroundColor = color.RGBA{R: 0xF6, G: 0xF8, B: 0xFA, A: 0xFF}
	defaultColor    = color.RGBA{R: 0x24, G: 0x2A, B: 0x2F, A: 0xFF}
	lineNumberColor = color.RGBA{R: 0x7A, G: 0x84, B: 0x8C, A: 0xFF}
	tabReplacement  = "    "
)

type Options struct {
	FontSize        float64
	LineHeight      float64
	Padding         int
	ShowLineNumbers bool
	Language        lang.Language

	lineNumbersSet bool
}

func (o Options) WithLineNumbers(enabled bool) Options {
	o.ShowLineNumbers = enabled
	o.lineNumbersSet = true
	return o
}

type segment struct {
	text  string
	color color.RGBA
}

func ToPNG(input doc.Document, w io.Writer, opts Options) error {
	if w == nil {
		return fmt.Errorf("output writer is nil")
	}

	opts = opts.withDefaults()
	renderDoc := input.Clone()
	if len(renderDoc.Lines) == 0 {
		renderDoc.Lines = []doc.Line{{Text: "", Origins: nil}}
	}

	fontData, err := opentype.Parse(jetBrainsMonoTTF)
	if err != nil {
		return fmt.Errorf("parse embedded font: %w", err)
	}

	face, err := opentype.NewFace(fontData, &opentype.FaceOptions{
		Size:    opts.FontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return fmt.Errorf("create font face: %w", err)
	}
	if c, ok := face.(io.Closer); ok {
		defer c.Close()
	}

	tokenLines := tokenize(string(renderDoc.Bytes()), opts.Language)
	normalizedSegments := normalizeSegmentLines(tokenLines, len(renderDoc.Lines))

	lineHeightPx := max(int(math.Ceil(opts.FontSize*opts.LineHeight)), 1)
	ascent := face.Metrics().Ascent.Ceil()
	if ascent < 1 {
		ascent = lineHeightPx
	}

	maxLineWidth := 1
	for _, line := range normalizedSegments {
		width := 0
		for _, seg := range line {
			width += font.MeasureString(face, seg.text).Ceil()
		}
		if width > maxLineWidth {
			maxLineWidth = width
		}
	}

	lineLabels := make([]string, len(renderDoc.Lines))
	maxLabelWidth := 0
	for i, line := range renderDoc.Lines {
		label := formatLineLabel(line.Origins)
		lineLabels[i] = label
		if !opts.ShowLineNumbers || label == "" {
			continue
		}
		width := font.MeasureString(face, label).Ceil()
		if width > maxLabelWidth {
			maxLabelWidth = width
		}
	}

	gutterWidth := 0
	if opts.ShowLineNumbers && maxLabelWidth > 0 {
		gutterWidth = maxLabelWidth + 16
	}

	imgWidth := opts.Padding*2 + gutterWidth + maxLineWidth
	imgHeight := opts.Padding*2 + lineHeightPx*len(renderDoc.Lines)
	if imgWidth < 1 {
		imgWidth = 1
	}
	if imgHeight < 1 {
		imgHeight = 1
	}

	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: backgroundColor}, image.Point{}, draw.Src)

	d := font.Drawer{Dst: img, Face: face}
	y := opts.Padding + ascent
	for i, line := range normalizedSegments {
		if opts.ShowLineNumbers {
			label := lineLabels[i]
			if label != "" {
				labelWidth := font.MeasureString(face, label).Ceil()
				x := opts.Padding + maxLabelWidth - labelWidth
				d.Src = image.NewUniform(lineNumberColor)
				d.Dot = fixed.P(x, y)
				d.DrawString(label)
			}
		}

		x := opts.Padding + gutterWidth
		for _, seg := range line {
			d.Src = image.NewUniform(seg.color)
			d.Dot = fixed.P(x, y)
			d.DrawString(seg.text)
			x += font.MeasureString(face, seg.text).Ceil()
		}
		y += lineHeightPx
	}

	if err := png.Encode(w, img); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}
	return nil
}

func (o Options) withDefaults() Options {
	if o.FontSize <= 0 {
		o.FontSize = 16
	}
	if o.LineHeight <= 0 {
		o.LineHeight = 1.35
	}
	if o.Padding <= 0 {
		o.Padding = 24
	}
	if !o.lineNumbersSet {
		o.ShowLineNumbers = true
	}
	if o.Language == "" {
		o.Language = lang.LanguageGo
	}
	return o
}

func normalizeSegmentLines(lines [][]segment, expected int) [][]segment {
	if expected <= 0 {
		return [][]segment{{}}
	}
	out := make([][]segment, expected)
	for i := range expected {
		if i < len(lines) {
			out[i] = lines[i]
		} else {
			out[i] = []segment{}
		}
	}
	return out
}

func normalizeRenderableText(text string) string {
	var b strings.Builder
	b.Grow(len(text))

	for _, r := range text {
		switch {
		case r == '\t':
			b.WriteString(tabReplacement)
		case r == '\n':
			b.WriteRune(r)
		case unicode.IsControl(r):
			b.WriteRune(' ')
		case unicode.IsSpace(r) && r != ' ':
			b.WriteRune(' ')
		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}

func formatLineLabel(origins []int) string {
	normalized := normalizeOrigins(origins)
	if len(normalized) == 0 {
		return ""
	}
	if len(normalized) == 1 {
		return strconv.Itoa(normalized[0])
	}
	if isContiguous(normalized) {
		return strconv.Itoa(normalized[0]) + "-" + strconv.Itoa(normalized[len(normalized)-1])
	}

	parts := make([]string, len(normalized))
	for i, v := range normalized {
		parts[i] = strconv.Itoa(v)
	}
	return strings.Join(parts, ",")
}

func normalizeOrigins(origins []int) []int {
	if len(origins) == 0 {
		return nil
	}
	set := make(map[int]struct{}, len(origins))
	for _, v := range origins {
		set[v] = struct{}{}
	}
	out := slices.Collect(maps.Keys(set))
	slices.Sort(out)
	return out
}

func isContiguous(values []int) bool {
	if len(values) <= 1 {
		return true
	}
	for i := 1; i < len(values); i++ {
		if values[i] != values[i-1]+1 {
			return false
		}
	}
	return true
}
