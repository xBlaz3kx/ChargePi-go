package evcc

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/warthog618/gpiod"
)

var (
	ErrInvalidPinNumber = errors.New("pin number must be greater than 0")
	ErrInitFailed       = errors.New("init failed")
)

type (
	relayAsEvcc struct {
		RelayPin     int
		InverseLogic bool
		currentState bool
		pin          *gpiod.Line
	}
)

// NewRelay creates a new RelayImpl struct that will communicate with the GPIO pin specified.
func NewRelay(relayPin int, inverseLogic bool) (*relayAsEvcc, error) {
	if relayPin <= 0 {
		return nil, ErrInvalidPinNumber
	}

	log.Debugf("Creating new relay at pin %d", relayPin)
	relay := relayAsEvcc{
		RelayPin:     relayPin,
		InverseLogic: inverseLogic,
		currentState: inverseLogic,
	}

	err := relay.Init(nil)
	if err != nil {
		return nil, ErrInitFailed
	}

	return &relay, nil
}

func (r *relayAsEvcc) Init(ctx context.Context) error {
	// Refer to gpiod docs
	c, err := gpiod.NewChip("gpiochip0")
	if err != nil {
		return err
	}

	r.pin, err = c.RequestLine(r.RelayPin, gpiod.AsOutput(0))
	return err
}

func (r *relayAsEvcc) Lock() {
}

func (r *relayAsEvcc) Unlock() {
}

func (r *relayAsEvcc) GetState() string {
	return ""
}

func (r *relayAsEvcc) EnableCharging() error {
	if r.InverseLogic {
		_ = r.pin.SetValue(0)
	} else {
		_ = r.pin.SetValue(1)
	}

	// Always consider positive logic for status determination
	r.currentState = true
	return nil
}

func (r *relayAsEvcc) DisableCharging() {
	if r.InverseLogic {
		_ = r.pin.SetValue(1)
	} else {
		_ = r.pin.SetValue(0)
	}

	// Always consider positive logic for status determination
	r.currentState = false
}

func (r *relayAsEvcc) SetMaxChargingCurrent(value float64) error {
	return nil
}

func (r *relayAsEvcc) GetMaxChargingCurrent() float64 {
	return 0.0
}

func (r *relayAsEvcc) Cleanup() error {
	return r.pin.Close()
}
