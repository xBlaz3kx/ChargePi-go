# ➡️Adding hardware support

There are four hardware component groups that are included in the project:

1. NFC/RFID tag reader,
2. LCD (display),
3. (Led) Indicator,
4. Power meter,
5. EVCC.

These hardware components have corresponding interfaces that are included in the `ChargePointHandler` struct. This
allows adding support for other models of hardware with similar functionalities.

You're welcome to submit a Pull Request with any additional hardware model implementations! Be sure to test and document
your changes, update the [supported hardware](../hardware/hardware.md) table(s) with the new hardware model(s). It would
be nice to have a wiring sketch or a connection table included for the new model(s).

## 💳 Reader hardware

All readers must implement the `Reader` interface. It is recommended that you implement the interface in a new file
named after the model of the reader in the `hardware/reader` package. Then you should add a **constant** named after
the **model** of the reader in the `reader` file in the package and add a switch case with the implementation and the
necessary logic that returns a pointer to the struct.

The settings of the reader are read from the `settings.json` file, which is stored in the cache and are available in the
NewTagReader method.

```golang
package reader

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/xBlaz3kx/ChargePi-go/internal/models/settings"
)

// Supported readers - by libnfc
const (
	PN532 = "PN532"
)

var (
	ErrReaderUnsupported = errors.New("reader type unsupported")
	ErrReaderDisabled    = errors.New("reader disabled")
)

// Reader is an abstraction for an RFID/NFC tag reader.
type Reader interface {
	ListenForTags(ctx context.Context)
	Cleanup()
	Reset()
	GetTagChannel() <-chan string
}

// NewTagReader creates an instance of the Reader interface based on the provided configuration.
func NewTagReader(reader settings.TagReader) (Reader, error) {
	if reader.IsEnabled {
		log.Infof("Preparing tag reader from config: %s", reader.ReaderModel)

		switch reader.ReaderModel {
		case PN532:
			return nil, nil
			// Your custom implmentation
		default:
			return nil, ErrReaderUnsupported
		}
	}

	return nil, ErrReaderDisabled
}

```

## 🖥️ Display hardware

All displays must implement the `LCD` interface. It is recommended that you implement the interface in a new file named
after the model of the display/LCD in the `hardware/display` package. Then you should add a **constant** named after
the **model** of the display in the `display` file in the package and add a switch case with the implementation and the
necessary logic that returns a pointer to the struct.

```golang
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
func NewDisplay(lcdSettings settings.Display) (LCD, error) {
	if lcdSettings.IsEnabled {
		log.Info("Preparing LCD from config")

		switch lcdSettings.Driver {
		case DriverHD44780:
		// custom implementation
		default:
			return nil, ErrDisplayUnsupported
		}
	}

	return nil, ErrDisplayDisabled
}

```

## Indicator hardware

The process is the same as the previous description.

```golang
package indicator

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// color constants
const (
	Off    = 0x0
	White  = 0xFFFFFF
	Red    = 0xff0000
	Green  = 0x00ff00
	Blue   = 0x000ff
	Yellow = 0xeeff00
	Orange = 0xff7b00
)

// Supported types
const (
	TypeWS281x = "WS281x"
)

var (
	ErrInvalidIndex        = errors.New("invalid index")
	ErrInvalidPin          = errors.New("invalid data pin #")
	ErrInvalidNumberOfLeds = errors.New("number of leds must be greater than zero")
)

type (
	// Indicator is an abstraction layer for connector status indication, usually an RGB LED strip.
	Indicator interface {
		DisplayColor(index int, colorHex uint32) error
		Blink(index int, times int, colorHex uint32) error
		Cleanup()
	}
)

// NewIndicator constructs the Indicator based on the type provided by the settings file.
func NewIndicator(stripLength int) Indicator {
	var (
		indicatorEnabled = viper.GetBool("chargepoint.hardware.ledIndicator.enabled")
		indicatorType    = viper.GetString("chargepoint.hardware.ledIndicator.type")
		indicateCardRead = viper.GetBool("chargepoint.hardware.ledIndicator.indicateCardRead")
	)

	if indicatorEnabled {
		if indicateCardRead {
			stripLength++
		}

		log.Infof("Preparing Indicator from config: %s", indicatorType)
		switch indicatorType {
		case TypeWS281x:
			// Your custom implementation
		default:
			return nil
		}
	}

	return nil
}
```

## ⚡ Power meters

The process is the same as the previous description.

```golang
package powerMeter

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/xBlaz3kx/ChargePi-go/internal/models/settings"
)

// Supported power meters
const (
	TypeC5460A = "cs5460a"
)

var (
	ErrPowerMeterUnsupported = errors.New("power meter type not supported")
	ErrPowerMeterDisabled    = errors.New("power meter not enabled")
)

// PowerMeter is an abstraction for measurement hardware.
type (
	PowerMeter interface {
		Init(ctx context.Context) error
		Reset()
		GetEnergy() float64
		GetPower() float64
		GetCurrent() float64
		GetVoltage() float64
		GetRMSCurrent() float64
		GetRMSVoltage() float64
	}
)

// NewPowerMeter creates a new power meter based on the connector settings.
func NewPowerMeter(meterSettings settings.PowerMeter) (PowerMeter, error) {
	if meterSettings.Enabled {
		log.Infof("Creating a new power meter: %s", meterSettings.Type)

		switch meterSettings.Type {
		case TypeC5460A:
			// Your custom implementation
			return nil, nil
		default:
			return nil, ErrPowerMeterUnsupported
		}
	}

	return nil, ErrPowerMeterDisabled
}
```

## ⚡ EVCC

The process is similar to other components. Note: Init method will be called whenever the charge point boots.
The init method should perform any necessary setup steps, such as opening a communication path. Any two-way
communication should be initiated in another thread and should communicate through channels.

```golang
package evcc

import (
	"context"
	"github.com/xBlaz3kx/ChargePi-go/internal/models/settings"
)

const (
	PhoenixEMCPPPETH = "EM-CP-PP-ETH"
	Relay            = "Relay"
)

type (
	EVCC interface {
		Init(ctx context.Context) error
		EnableCharging() error
		DisableCharging()
		SetMaxChargingCurrent(value float64) error
		GetMaxChargingCurrent() float64
		Lock()
		Unlock()
		GetState() string
		Cleanup() error
	}
)

// NewPowerMeter creates a new power meter based on the connector settings.
func NewEVCCFromType(evccSettings settings.EVCC) (EVCC, error) {
	switch evccSettings.Type {
	case Relay:
		return nil, nil
	case PhoenixEMCPPPETH:
		return nil, nil
		// Your custom implementation
	default:
		return nil, nil
	}
}

}
```
