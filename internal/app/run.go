package app

import (
	"errors"
	"fmt"
	"io"

	"github.com/jacoelho/gophershot/internal/config"
	"github.com/jacoelho/gophershot/internal/doc"
	"github.com/jacoelho/gophershot/internal/lang"
	"github.com/jacoelho/gophershot/internal/pipeline"
	"github.com/jacoelho/gophershot/internal/plan"
	"github.com/jacoelho/gophershot/internal/render"
	"github.com/jacoelho/gophershot/internal/transform"
)

type Config struct {
	Args   []string
	Input  io.Reader
	Output io.Writer
}

type renderImageFunc func(input doc.Document, w io.Writer, opts plan.RenderOptions) error

type service struct {
	renderImage renderImageFunc
	catalog     plan.TransformCatalog
}

func NewService() service {
	return service{
		renderImage: func(input doc.Document, w io.Writer, opts plan.RenderOptions) error {
			renderOpts := render.Options{
				FontSize: opts.FontSize,
				Language: opts.Language,
			}.WithLineNumbers(opts.ShowLineNumbers)
			return render.ToPNG(input, w, renderOpts)
		},
		catalog: transform.NewDefaultRegistry(),
	}
}

func Run(cfg Config) error {
	parsed, err := config.Parse(cfg.Args)
	if err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	return RunParsed(parsed, cfg.Input, cfg.Output)
}

func RunParsed(cfg config.Config, input io.Reader, output io.Writer) error {
	return NewService().runParsed(cfg, input, output)
}

func (s service) runParsed(cfg config.Config, input io.Reader, output io.Writer) error {
	if input == nil {
		return errors.New("input reader is nil")
	}
	if output == nil {
		return errors.New("output writer is nil")
	}

	compiled, err := plan.Compile(plan.Request{
		LineSelector: cfg.LineSelector,
		LineNumbers:  cfg.LineNumbers,
		FontSize:     cfg.FontSize,
		Transforms:   cfg.Transforms,
	}, s.catalog)
	if err != nil {
		return err
	}
	compiled.Render.Language = lang.FromInputPath(cfg.InputPath)

	src, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	d, err := pipeline.Execute(src, compiled)
	if err != nil {
		return err
	}

	if err := s.renderImage(d, output, compiled.Render); err != nil {
		return fmt.Errorf("render png: %w", err)
	}

	return nil
}
