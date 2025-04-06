package printer

import (
	"os"

	"github.com/pterm/pterm"
)

var instance Printer

var defaultPrinterConfig = PrinterConfig{
	UseColor:  true,
	LogLevel:  Default,
	OutStream: os.Stdout,
}

func init() {
	Configure(defaultPrinterConfig)
}

func Configure(config PrinterConfig) {
	if config.LogLevel < Default {
		pterm.EnableDebugMessages()
	} else {
		pterm.DisableDebugMessages()
	}

	if config.UseColor {
		pterm.EnableColor()
	} else {
		pterm.DisableColor()
	}

	pterm.SetDefaultOutput(config.OutStream)

	instance = NewPrinter(config)
}

func Instance() Printer {
	return instance
}
