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
		return NewRelay(evccSettings.RelayPin, evccSettings.InverseLogic)
	case PhoenixEMCPPPETH:
		return nil, nil
	default:
		return nil, nil
	}
}
