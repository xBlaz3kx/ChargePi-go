package settings

import (
	"github.com/lorenzodonini/ocpp-go/ocpp1.6/types"
)

const (
	OCPP16  = ProtocolVersion("1.6")
	OCPP201 = ProtocolVersion("2.0.1")
)

type (
	ProtocolVersion string

	OCPPInfo struct {
		Vendor                  string `fig:"Vendor" default:"UL FE" json:"vendor,omitempty" yaml:"vendor" mapstructure:"vendor"`
		Model                   string `fig:"Model" default:"ChargePi" json:"model,omitempty" yaml:"model" mapstructure:"model"`
		ChargeBoxSerialNumber   string `fig:"ChargeBoxSerialNumber" default:"" json:"charge_box_serial_number,omitempty" yaml:"charge_box_serial_number" mapstructure:"charge_box_serial_number"`
		ChargePointSerialNumber string `fig:"ChargePointSerialNumber" default:"" json:"charge_point_serial_number,omitempty" yaml:"charge_point_serial_number" mapstructure:"charge_point_serial_number"`
		Iccid                   string `fig:"Iccid" default:"" json:"iccid,omitempty" yaml:"iccid" mapstructure:"iccid"`
		Imsi                    string `fig:"Imsi" default:"" json:"imsi,omitempty" yaml:"imsi" mapstructure:"imsi"`
	}

	Connector struct {
		EvseId      int        `fig:"EvseId" validate:"required" json:"EvseId,omitempty" yaml:"EvseId" mapstructure:"EvseId"`
		ConnectorId int        `fig:"ConnectorId" validate:"required" json:"ConnectorId,omitempty" yaml:"ConnectorId" mapstructure:"ConnectorId"`
		Type        string     `fig:"Type" validate:"required" json:"type,omitempty" yaml:"type" mapstructure:"type"`
		Status      string     `fig:"Status" validation:"required" json:"status,omitempty" yaml:"status" mapstructure:"status"`
		Session     Session    `fig:"Session" json:"session" yaml:"session" mapstructure:"session"`
		EVCC        EVCC       `fig:"evcc" json:"evcc" yaml:"evcc" mapstructure:"evcc"`
		PowerMeter  PowerMeter `fig:"PowerMeter" json:"PowerMeter" yaml:"PowerMeter" mapstructure:"PowerMeter"`
	}

	Session struct {
		IsActive      bool               `fig:"IsActive" json:"IsActive,omitempty" yaml:"IsActive" mapstructure:"IsActive"`
		TransactionId string             `fig:"TransactionId" default:"" json:"TransactionId,omitempty" yaml:"TransactionId" mapstructure:"TransactionId"`
		TagId         string             `fig:"TagId" default:"" json:"TagId,omitempty" yaml:"TagId" mapstructure:"TagId"`
		Started       string             `fig:"Started" default:"" json:"started,omitempty" yaml:"started" mapstructure:"started"`
		Consumption   []types.MeterValue `fig:"Consumption" json:"consumption,omitempty" yaml:"consumption" mapstructure:"consumption"`
	}
)
