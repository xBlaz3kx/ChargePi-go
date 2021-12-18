package power_meter

import (
	"errors"
	"github.com/xBlaz3kx/ChargePi-go/data/settings"
	"log"
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
		log.Println("Creating a new power meter:", meterSettings.Type)
		switch meterSettings.Type {
		case TypeC5460A:
			return NewCS5460PowerMeter(
				meterSettings.PowerMeterPin,
				meterSettings.SpiBus,
				meterSettings.ShuntOffset,
				meterSettings.VoltageDividerOffset,
			)
		default:
			return nil, ErrPowerMeterUnsupported
		}
	}

	return nil, ErrPowerMeterDisabled
}
