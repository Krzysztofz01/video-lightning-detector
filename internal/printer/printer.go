package printer

import (
	"fmt"
	"io"

	"github.com/pterm/pterm"
)

// TODO: Implement Printer and PrinterConfig unit test

type PrinerLogLevel int

const (
	Quiet PrinerLogLevel = iota
	Default
	Debug
)

type PrinterConfig struct {
	UseColor  bool
	LogLevel  PrinerLogLevel
	OutStream io.Writer
}

func (c *PrinterConfig) IsValid() (bool, string) {
	switch c.LogLevel {
	case Quiet, Default, Debug:
	default:
		return false, "invalid unsupported log level"
	}

	if c.OutStream == nil {
		return false, "invalid nil out writer reference"
	}

	return true, ""
}

type Printer interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	InfoA(format string, args ...any)
	Warning(format string, args ...any)
	Error(format string, args ...any)
	Progress(msg string) (finalize func())
	ProgressSteps(msg string, steps int) (step func(), finalize func())
	Table(data [][]string)
}

type printer struct {
	config PrinterConfig
}

func (p *printer) Debug(format string, args ...any) {
	if p.config.LogLevel >= Debug {
		return
	}

	pterm.DefaultBasicText.WithStyle(&pterm.ThemeDefault.DescriptionMessageStyle).Printf(format, args...)
}

func (p *printer) Error(format string, args ...any) {
	pterm.DefaultBasicText.WithStyle(&pterm.ThemeDefault.ErrorMessageStyle).Printfln(format, args...)
}

func (p *printer) Info(format string, args ...any) {
	if p.config.LogLevel >= Default {
		return
	}

	p.InfoA(format, args...)
}

func (p *printer) InfoA(format string, args ...any) {
	pterm.DefaultBasicText.WithStyle(&pterm.ThemeDefault.InfoMessageStyle).Printfln(format, args...)
}

func (p *printer) Progress(msg string) (finalize func()) {
	if p.config.LogLevel >= Default {
		return func() {}
	}

	spinner, err := pterm.DefaultSpinner.
		WithText(msg).
		WithShowTimer(true).
		WithRemoveWhenDone(true).Start()

	if err != nil {
		panic(fmt.Errorf("printer: failed to start the underlying progress instance: %w", err))
	}

	return func() {
		if err := spinner.Stop(); err != nil {
			panic(fmt.Errorf("printer: failed to stop the underlying progress instance: %w", err))
		}
	}
}

func (p *printer) ProgressSteps(msg string, steps int) (step func(), finalize func()) {
	if p.config.LogLevel >= Default {
		return func() {}, func() {}
	}

	progress, err := pterm.DefaultProgressbar.
		WithTotal(steps).
		WithTitle(msg).
		WithShowPercentage(true).
		WithRemoveWhenDone(true).
		WithShowCount(true).Start()

	if err != nil {
		panic(fmt.Errorf("printer: failed to start the underlying progress steps instance: %w", err))
	}

	stepFunc := func() {
		progress.Increment()
	}

	finalizeFunc := func() {
		if _, err := progress.Stop(); err != nil {
			panic(fmt.Errorf("printer: failed to stop the underlying progress steps instance: %w", err))
		}
	}

	return stepFunc, finalizeFunc
}

func (p *printer) Table(data [][]string) {
	if p.config.LogLevel >= Default {
		return
	}

	table := pterm.DefaultTable.
		WithBoxed(true).
		WithHasHeader(false).
		WithData(data)

	if err := table.Render(); err != nil {
		panic(fmt.Errorf("printer: failed to render the underlying table instance: %w", err))
	}
}

func (p *printer) Warning(format string, args ...any) {
	pterm.DefaultBasicText.WithStyle(&pterm.ThemeDefault.WarningMessageStyle).Printfln(format, args...)
}

func NewPrinter(config PrinterConfig) Printer {
	if ok, msg := config.IsValid(); !ok {
		panic(fmt.Errorf("printer: failed to create printer due to invalid config cause by %s", msg))
	}

	printer := &printer{
		config: config,
	}

	return printer
}
