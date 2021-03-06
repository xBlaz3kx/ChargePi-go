package display

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/xBlaz3kx/ChargePi-go/internal/models/settings"
	"time"
)

const (
	DriverHD44780 = "hd44780"
)

var (
	ErrDisplayUnsupported = errors.New("display type unsupported")
	ErrDisplayDisabled    = errors.New("display disabled")
)

type (
	// LCDMessage Object representing the message that will be displayed on the LCD.
	// Each array element in Messages represents a line being displayed on the 16x2 screen.
	LCDMessage struct {
		Messages        []string
		MessageDuration time.Duration
	}

	// LCD is an abstraction layer for concrete implementation of a display.
	LCD interface {
		DisplayMessage(message LCDMessage)
		ListenForMessages(ctx context.Context)
		Cleanup()
		Clear()
		GetLcdChannel() chan<- LCDMessage
	}
)

// NewMessage creates a new message for the LCD.
func NewMessage(duration time.Duration, messages []string) LCDMessage {
	return LCDMessage{
		Messages:        messages,
		MessageDuration: duration,
	}
}

// NewDisplay returns a concrete implementation of an LCD based on the drivers that are supported.
// The LCD is built with the settings from the settings file.
func NewDisplay(lcdSettings settings.Lcd) (LCD, error) {
	if lcdSettings.IsEnabled {
		log.Info("Preparing LCD from config")

		lcdChannel := make(chan LCDMessage, 5)

		switch lcdSettings.Driver {
		case DriverHD44780:
			lcd, err := NewHD44780(lcdChannel, lcdSettings.I2CAddress, lcdSettings.I2CBus)
			if err != nil {
				return nil, err
			}

			return lcd, nil
		default:
			return nil, ErrDisplayUnsupported
		}
	}

	return nil, ErrDisplayDisabled
}
