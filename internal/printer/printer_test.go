package printer

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
)

func TestPrinterConfigShouldTellIfValid(t *testing.T) {
	pc := PrinterConfig{
		UseColor:     false,
		LogLevel:     options.Verbose,
		OutStream:    nil,
		ParsableMode: false,
	}

	ok, msg := pc.IsValid()
	assert.False(t, ok)
	assert.NotEmpty(t, msg)

	pc = PrinterConfig{
		UseColor:     false,
		LogLevel:     -999,
		OutStream:    &bytes.Buffer{},
		ParsableMode: false,
	}

	ok, msg = pc.IsValid()
	assert.False(t, ok)
	assert.NotEmpty(t, msg)

	pc = PrinterConfig{
		UseColor:     false,
		LogLevel:     options.Verbose,
		OutStream:    &bytes.Buffer{},
		ParsableMode: false,
	}

	ok, msg = pc.IsValid()
	assert.True(t, ok)
	assert.Empty(t, msg)
}

func TestPrinterShouldCreateAndNotFailAtExposedApiCalls(t *testing.T) {
	cases := []PrinterConfig{
		{
			UseColor:     false,
			LogLevel:     options.Verbose,
			ParsableMode: false,
		},
		{
			UseColor:     true,
			LogLevel:     options.Verbose,
			ParsableMode: false,
		},
		{
			UseColor:     false,
			LogLevel:     options.Verbose,
			ParsableMode: true,
		},
		{
			UseColor:     false,
			LogLevel:     options.Info,
			ParsableMode: true,
		},
		{
			UseColor:     false,
			LogLevel:     options.Quiet,
			ParsableMode: true,
		},
	}

	for _, c := range cases {
		stream := &bytes.Buffer{}
		c.OutStream = stream

		// NOTE: Configuring and accessing the singleton printer instead of creating one
		Configure(c)
		p := Instance()

		assert.NotNil(t, p)

		p.Debug("Hello world %s", "test")
		p.Info("Hello world %s", "test")
		p.InfoA("Hello world %s", "test")
		p.Warning("Hello world %s", "test")
		p.Error("Hello world %s", "test")

		finalizeProgress := p.Progress("Hello world")
		finalizeProgress()

		stepCount := 4
		finalizeProgressStep, step := p.ProgressSteps("Hello world", stepCount)
		for i := 0; i < stepCount; i += 1 {
			step()
		}
		finalizeProgressStep()

		p.Table([][]string{
			{"Hello", "World"},
			{"Hello", "World"},
		})

		p.WriteParsable(struct {
			Message string `json:"message"`
		}{
			Message: "Hello world",
		})

		assert.True(t, p.IsLogLevel(c.LogLevel))
	}
}
