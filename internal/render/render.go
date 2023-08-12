package render

import (
	"fmt"

	"github.com/pterm/pterm"
)

// TODO: Add unit tests

type Renderer interface {
	// Render a formated text message with the Debug log level.
	LogDebug(format string, a ...any)

	// Render a formated text message with the Info log level.
	LogInfo(format string, a ...any)

	// Render a formated text message with the warning log level.
	LogWarning(format string, a ...any)

	// Render a formated text message with the error log level.
	LogError(format string, a ...any)

	// Render a progress bar with a given title and amount of steps. The functions returns a function
	// to increment the progress bar and a function to close the progress bar at the end.
	Progress(title string, steps int) (func(), func())

	// Render a spinner with a given title. The function return a function to close the spinner at the end.
	Spinner(title string) func()

	// Render a table filled with the values from the privided two dimentional slice.
	Table(data [][]string)
}

// Create a new renderer instance and specify if verbose debug logging should be enabled
func CreateRenderer(verbose bool) Renderer {
	pterm.EnableColor()
	if verbose {
		pterm.EnableDebugMessages()
	}

	return &ptermRenderer{
		verbose: verbose,
	}
}

type ptermRenderer struct {
	verbose bool
}

func (r *ptermRenderer) LogDebug(format string, a ...any) {
	if !r.verbose {
		return
	}

	pterm.DefaultBasicText.WithStyle(&pterm.ThemeDefault.DescriptionMessageStyle).Printfln(format, a...)
}

func (r *ptermRenderer) LogInfo(format string, a ...any) {
	pterm.DefaultBasicText.WithStyle(&pterm.ThemeDefault.InfoMessageStyle).Printfln(format, a...)
}

func (r *ptermRenderer) LogWarning(format string, a ...any) {
	pterm.DefaultBasicText.WithStyle(&pterm.ThemeDefault.WarningMessageStyle).Printfln(format, a...)
}

func (r *ptermRenderer) LogError(format string, a ...any) {
	pterm.DefaultBasicText.WithStyle(&pterm.ThemeDefault.ErrorMessageStyle).Printfln(format, a...)
}

func (r *ptermRenderer) Progress(title string, steps int) (func(), func()) {
	progress, err := pterm.DefaultProgressbar.
		WithTotal(steps).
		WithTitle(title).
		WithShowPercentage(true).
		WithRemoveWhenDone(true).
		WithShowCount(true).Start()

	if err != nil {
		panic(fmt.Errorf("render: failed to start the underlying progress bar instance: %w", err))
	}

	stepFunc := func() {
		progress.Increment()
	}

	stopFunc := func() {
		if _, err := progress.Stop(); err != nil {
			panic(fmt.Errorf("render: failed to stop the underlying progress bar instance: %w", err))
		}
	}

	return stepFunc, stopFunc
}

func (r *ptermRenderer) Spinner(title string) func() {
	spinner, err := pterm.DefaultSpinner.
		WithText(title).
		WithShowTimer(true).
		WithRemoveWhenDone(true).Start()

	if err != nil {
		panic(fmt.Errorf("render: failed to start the underlying spinner instance: %w", err))
	}

	return func() {
		if err := spinner.Stop(); err != nil {
			panic(fmt.Errorf("render: failed to stop the underlying spinner instance: %w", err))
		}
	}
}

func (r *ptermRenderer) Table(data [][]string) {
	table := pterm.DefaultTable.
		WithBoxed(true).
		WithHasHeader(false).
		WithData(data)

	if err := table.Render(); err != nil {
		panic(fmt.Errorf("render: failed to render the underlying table instance: %w", err))
	}
}
