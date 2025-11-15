package printer

import (
	"os"

	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/pterm/pterm"
)

var instance Printer

var defaultPrinterConfig = PrinterConfig{
	UseColor:  true,
	LogLevel:  options.Info,
	OutStream: os.Stdout,
}

func init() {
	Configure(defaultPrinterConfig)
}

func Configure(config PrinterConfig) {
	if config.LogLevel < options.Info {
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
