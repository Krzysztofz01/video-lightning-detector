package printer

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/pterm/pterm"

	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
)

// TODO: Implement Printer and PrinterConfig unit test

type PrinterConfig struct {
	UseColor     bool
	LogLevel     options.LogLevel
	OutStream    io.Writer
	ParsableMode bool
}

func (c *PrinterConfig) IsValid() (bool, string) {
	if !options.IsValidLogLevel(c.LogLevel) {
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
	WriteParsable(data any)
	WriteRaw(format string, args ...any)
	IsLogLevel(l options.LogLevel) bool
}

type printer struct {
	Config PrinterConfig
}

func (p *printer) Debug(format string, args ...any) {
	if p.Config.LogLevel < options.Verbose {
		return
	}

	if p.Config.ParsableMode {
		p.LogParsable("verbose", format, args...)
	} else {
		p.LogTerminal(&pterm.ThemeDefault.DescriptionMessageStyle, format, args...)
	}

}

func (p *printer) Error(format string, args ...any) {
	if p.Config.ParsableMode {
		p.LogParsable("error", format, args...)
	} else {
		p.LogTerminal(&pterm.ThemeDefault.ErrorMessageStyle, format, args...)
	}
}

func (p *printer) Info(format string, args ...any) {
	if p.Config.LogLevel < options.Info {
		return
	}

	p.InfoA(format, args...)
}

func (p *printer) InfoA(format string, args ...any) {
	if p.Config.ParsableMode {
		p.LogParsable("info", format, args...)
	} else {
		p.LogTerminal(&pterm.ThemeDefault.InfoMessageStyle, format, args...)
	}
}

func (p *printer) Progress(msg string) (finalize func()) {
	if p.Config.LogLevel < options.Info {
		return func() {}
	}

	if p.Config.ParsableMode {
		p.Warning("Progress printing not supported in parsable mode")
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
	if p.Config.LogLevel < options.Info {
		return func() {}, func() {}
	}

	if p.Config.ParsableMode {
		p.Warning("Step progress printing not supported in parsable mode")
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
	if p.Config.LogLevel < options.Info {
		return
	}

	if p.Config.ParsableMode {
		p.Warning("Table printing not supported in parsable mode")
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
	if p.Config.ParsableMode {
		p.LogParsable("warn", format, args...)
	} else {
		p.LogTerminal(&pterm.ThemeDefault.WarningMessageStyle, format, args...)
	}
}

func (p *printer) IsLogLevel(l options.LogLevel) bool {
	return p.Config.LogLevel >= l
}

func (p *printer) WriteParsable(data any) {
	dataJson, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Errorf("printer: failed to marshall parsable message: %w", err))
	}

	if _, err := fmt.Fprintln(p.Config.OutStream, string(dataJson)); err != nil {
		panic(fmt.Errorf("printer: failed to write parsable to out stream"))
	}
}

func (p *printer) WriteRaw(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if _, err := fmt.Fprint(p.Config.OutStream, message); err != nil {
		panic(fmt.Errorf("printer: failed to write raw to out stream: %w", err))
	}
}

func (p *printer) LogTerminal(style *pterm.Style, format string, args ...any) {
	pterm.DefaultBasicText.WithStyle(style).Printfln(format, args...)
}

func (p *printer) LogParsable(level string, format string, args ...any) {
	log := struct {
		LogLevel string `json:"level"`
		Message  string `json:"message"`
	}{
		LogLevel: string(level),
		Message:  fmt.Sprintf(format, args...),
	}

	p.WriteParsable(log)
}

func NewPrinter(config PrinterConfig) Printer {
	if ok, msg := config.IsValid(); !ok {
		panic(fmt.Errorf("printer: failed to create printer due to invalid config cause by %s", msg))
	}

	printer := &printer{
		Config: config,
	}

	return printer
}
