package connector

import (
	"errors"
	"fmt"
	"github.com/lorenzodonini/ocpp-go/ocpp1.6/core"
	"github.com/lorenzodonini/ocpp-go/ocpp1.6/types"
	"github.com/reactivex/rxgo/v2"
	"github.com/xBlaz3kx/ChargePi-go/components/cache"
	"github.com/xBlaz3kx/ChargePi-go/components/hardware"
	"github.com/xBlaz3kx/ChargePi-go/components/hardware/power-meter"
	"github.com/xBlaz3kx/ChargePi-go/components/scheduler"
	"github.com/xBlaz3kx/ChargePi-go/components/settings/conf-manager"
	settings2 "github.com/xBlaz3kx/ChargePi-go/components/settings/settings-manager"
	"github.com/xBlaz3kx/ChargePi-go/data/session"
	"github.com/xBlaz3kx/ChargePi-go/data/settings"
	"log"
	"sync"
	"time"
)

var (
	ErrInvalidEvseId            = errors.New("invalid evse id")
	ErrInvalidConnectorId       = errors.New("invalid connector id")
	ErrInvalidReservationId     = errors.New("invalid reservation id")
	ErrInvalidConnectorStatus   = errors.New("invalid connector status")
	ErrRelayPointerNil          = errors.New("relay pointer cannot be nil")
	ErrSessionTimeLimitExceeded = errors.New("session time limit exceeded")
	ErrNotCharging              = errors.New("connector not charging")
)

type (
	ConnectorImpl struct {
		mu                           sync.Mutex
		EvseId                       int
		ConnectorId                  int
		ConnectorType                string
		ConnectorStatus              core.ChargePointStatus
		ErrorCode                    core.ChargePointErrorCode
		relay                        hardware.Relay
		powerMeter                   power_meter.PowerMeter
		PowerMeterEnabled            bool
		MaxChargingTime              int
		reservationId                int
		session                      *session.Session
		ConnectorNotificationChannel chan<- rxgo.Item
	}

	Connector interface {
		StartCharging(transactionId string, tagId string) error
		ResumeCharging(session session.Session) (error, int)
		StopCharging(reason core.Reason) error
		SetNotificationChannel(notificationChannel chan<- rxgo.Item)
		ReserveConnector(reservationId int) error
		RemoveReservation() error
		GetReservationId() int
		GetTagId() string
		GetTransactionId() string
		GetConnectorId() int
		GetEvseId() int
		CalculateSessionAvgEnergyConsumption() float32
		SamplePowerMeter(measurands []types.Measurand)
		SetStatus(status core.ChargePointStatus, errCode core.ChargePointErrorCode)
		GetStatus() (core.ChargePointStatus, core.ChargePointErrorCode)
		IsAvailable() bool
		IsPreparing() bool
		IsCharging() bool
		IsReserved() bool
		IsUnavailable() bool
		GetPowerMeter() power_meter.PowerMeter
		GetMaxChargingTime() int
		preparePowerMeterAtConnector() error
	}
)

// NewConnector Create a new connector object from the provided arguments. EvseId, connectorId and maxChargingTime must be greater than zero.
// When created, it makes an empty session, turns off the relay and defaults the status to Available.
func NewConnector(evseId int, connectorId int, connectorType string, relay hardware.Relay,
	powerMeter power_meter.PowerMeter, powerMeterEnabled bool, maxChargingTime int) (*ConnectorImpl, error) {
	if maxChargingTime <= 0 {
		maxChargingTime = 180
	}

	if evseId <= 0 {
		return nil, ErrInvalidEvseId
	}

	if connectorId <= 0 {
		return nil, ErrInvalidConnectorId
	}

	if relay == nil {
		return nil, ErrRelayPointerNil
	}

	relay.Disable()
	return &ConnectorImpl{
		mu:                sync.Mutex{},
		EvseId:            evseId,
		ConnectorId:       connectorId,
		ConnectorType:     connectorType,
		relay:             relay,
		powerMeter:        powerMeter,
		reservationId:     -1,
		PowerMeterEnabled: powerMeterEnabled,
		MaxChargingTime:   maxChargingTime,
		ConnectorStatus:   core.ChargePointStatusAvailable,
		session:           session.NewEmptySession(),
	}, nil
}

// StartCharging Start charging a connector if connector is available and session could be started.
// It turns on the relay (even if negative logic applies).
func (connector *ConnectorImpl) StartCharging(transactionId string, tagId string) error {
	if !(connector.IsAvailable() || connector.IsPreparing()) {
		return ErrInvalidConnectorStatus
	}

	connector.SetStatus(core.ChargePointStatusPreparing, core.NoError)
	sessionErr := connector.session.StartSession(transactionId, tagId)
	if sessionErr != nil {
		return sessionErr
	}

	connector.relay.Enable()
	connector.SetStatus(core.ChargePointStatusCharging, core.NoError)

	settings2.UpdateConnectorSessionInfo(
		connector.EvseId,
		connector.ConnectorId,
		&settings.Session{
			IsActive:      connector.session.IsActive,
			TagId:         connector.session.TagId,
			TransactionId: connector.session.TransactionId,
			Started:       connector.session.Started,
			Consumption:   connector.session.Consumption,
		})

	if connector.GetPowerMeter() != nil {
		sampleError := connector.preparePowerMeterAtConnector()
		if sampleError != nil {
			log.Printf("Cannot sample connector %d; %v \n", connector.GetConnectorId(), sampleError)
		}
	}

	return nil
}

// ResumeCharging Resumes or restores the charging state after boot if a charging session was active.
func (connector *ConnectorImpl) ResumeCharging(session session.Session) (err error, chargingTimeElapsed int) {
	// Set the transaction id so connector is able to stop the transaction if charging fails
	chargingTimeElapsed = connector.MaxChargingTime
	connector.session.TransactionId = session.TransactionId

	startedChargingTime, err := time.Parse(time.RFC3339, session.Started)
	if err != nil {
		return
	}

	chargingTimeElapsed = int(time.Now().Sub(startedChargingTime).Minutes())
	if connector.MaxChargingTime <= chargingTimeElapsed {
		chargingTimeElapsed = connector.MaxChargingTime
		err = ErrSessionTimeLimitExceeded
		return
	}

	if connector.IsCharging() || connector.IsPreparing() {
		sessionErr := connector.session.StartSession(session.TransactionId, session.TagId)
		if sessionErr != nil {
			return fmt.Errorf("cannot resume session: %v", sessionErr), 0
		}

		connector.relay.Enable()
		connector.session.Started = session.Started
		connector.session.Consumption = append(connector.session.Consumption, session.Consumption...)
		return nil, chargingTimeElapsed
	}

	return ErrInvalidConnectorStatus, connector.MaxChargingTime
}

// StopCharging Stops charging the connector by turning the relay off and ending the session.
func (connector *ConnectorImpl) StopCharging(reason core.Reason) error {
	if connector.IsCharging() || connector.IsPreparing() {
		connector.session.EndSession()
		connector.relay.Disable()

		settings2.UpdateConnectorSessionInfo(
			connector.EvseId,
			connector.ConnectorId,
			&settings.Session{
				IsActive:      connector.session.IsActive,
				TagId:         connector.session.TagId,
				TransactionId: connector.session.TransactionId,
				Started:       connector.session.Started,
				Consumption:   connector.session.Consumption,
			})

		switch reason {
		case core.ReasonEVDisconnected:
			connector.SetStatus(core.ChargePointStatusSuspendedEVSE, core.NoError)
			break
		case core.ReasonUnlockCommand:
			connector.SetStatus(core.ChargePointStatusUnavailable, core.NoError)
			break
		default:
			connector.SetStatus(core.ChargePointStatusFinishing, core.NoError)
			connector.SetStatus(core.ChargePointStatusAvailable, core.NoError)
		}
		return nil
	}

	return ErrNotCharging
}

// SamplePowerMeter Get a sample from the power meter. The measurands argument takes the list of all the types of the measurands to sample.
// It will add all the samples to the connector's Session if it is active.
func (connector *ConnectorImpl) SamplePowerMeter(measurands []types.Measurand) {
	if !connector.PowerMeterEnabled || connector.powerMeter == nil {
		return
	}

	log.Println("Sampling connector", connector.ConnectorId)
	var (
		samples []types.SampledValue
		value   = 0.0
	)

	for _, measurand := range measurands {
		value = 0.0
		switch measurand {
		case types.MeasurandEnergyActiveExportInterval:
			value = connector.powerMeter.GetEnergy()
			break
		case types.MeasurandCurrentExport:
			value = connector.powerMeter.GetCurrent()
			break
		case types.MeasurandPowerActiveExport:
			value = connector.powerMeter.GetPower()
			break
		case types.MeasurandVoltage:
			value = connector.powerMeter.GetVoltage()
			break
		}

		if value != 0.0 {
			samples = append(samples, types.SampledValue{
				Value:     fmt.Sprintf("%.3f", value),
				Measurand: measurand,
			})
		}
	}

	connector.session.AddSampledValue(samples)
}

// preparePowerMeterAtConnector
func (connector *ConnectorImpl) preparePowerMeterAtConnector() error {
	var (
		measurands                 = conf_manager.GetTypesToSample()
		sampleTime                 = "10s"
		sampleInterval, err        = conf_manager.GetConfigurationValue("MeterValueSampleInterval")
		meterValueLastSentIndexKey = fmt.Sprintf("MeterValueLastIndex%d%d", connector.EvseId, connector.ConnectorId)
		jobTag                     = fmt.Sprintf("Evse%dConnector%dSampling", connector.EvseId, connector.ConnectorId)
	)
	if err != nil {
		sampleInterval = "10"
	}

	cache.GetCache().Set(meterValueLastSentIndexKey, 0, time.Duration(connector.MaxChargingTime)*time.Minute)

	sampleTime = fmt.Sprintf("%ss", sampleInterval)
	// Schedule the sampling
	_, err = scheduler.GetScheduler().Every(sampleTime).
		Tag(jobTag).
		Do(connector.SamplePowerMeter, measurands)

	return err
}

func (connector *ConnectorImpl) IsAvailable() bool {
	connector.mu.Lock()
	defer connector.mu.Unlock()
	return connector.ConnectorStatus == core.ChargePointStatusAvailable
}
func (connector *ConnectorImpl) IsCharging() bool {
	connector.mu.Lock()
	defer connector.mu.Unlock()
	return connector.ConnectorStatus == core.ChargePointStatusCharging
}
func (connector *ConnectorImpl) IsPreparing() bool {
	connector.mu.Lock()
	defer connector.mu.Unlock()
	return connector.ConnectorStatus == core.ChargePointStatusPreparing
}
func (connector *ConnectorImpl) IsReserved() bool {
	connector.mu.Lock()
	defer connector.mu.Unlock()
	return connector.ConnectorStatus == core.ChargePointStatusReserved
}
func (connector *ConnectorImpl) IsUnavailable() bool {
	connector.mu.Lock()
	defer connector.mu.Unlock()
	return connector.ConnectorStatus == core.ChargePointStatusUnavailable
}

func (connector *ConnectorImpl) SetStatus(status core.ChargePointStatus, errCode core.ChargePointErrorCode) {
	connector.mu.Lock()
	connector.ConnectorStatus = status
	connector.ErrorCode = errCode
	settings2.UpdateConnectorStatus(connector.EvseId, connector.ConnectorId, status)

	if connector.ConnectorNotificationChannel != nil {
		connector.ConnectorNotificationChannel <- rxgo.Of(connector)
	}

	connector.mu.Unlock()
}

func (connector *ConnectorImpl) GetTransactionId() string {
	return connector.session.TransactionId
}
func (connector *ConnectorImpl) GetTagId() string {
	return connector.session.TagId
}

func (connector *ConnectorImpl) ReserveConnector(reservationId int) error {
	if reservationId <= 0 {
		return ErrInvalidReservationId
	}

	if !connector.IsAvailable() {
		return ErrInvalidConnectorStatus
	}

	connector.reservationId = reservationId
	connector.SetStatus(core.ChargePointStatusReserved, core.NoError)
	return nil
}
func (connector *ConnectorImpl) RemoveReservation() error {
	if !connector.IsReserved() {
		return ErrInvalidConnectorStatus
	}

	connector.reservationId = -1
	connector.SetStatus(core.ChargePointStatusAvailable, core.NoError)
	return nil
}

func (connector *ConnectorImpl) GetReservationId() int {
	return connector.reservationId
}

func (connector *ConnectorImpl) GetConnectorId() int {
	return connector.ConnectorId
}

func (connector *ConnectorImpl) GetEvseId() int {
	return connector.EvseId
}

func (connector *ConnectorImpl) CalculateSessionAvgEnergyConsumption() float32 {
	return connector.session.CalculateEnergyConsumptionWithAvgPower()
}

func (connector *ConnectorImpl) GetPowerMeter() power_meter.PowerMeter {
	return connector.powerMeter
}

func (connector *ConnectorImpl) GetMaxChargingTime() int {
	return connector.MaxChargingTime
}

func (connector *ConnectorImpl) GetStatus() (core.ChargePointStatus, core.ChargePointErrorCode) {
	return connector.ConnectorStatus, connector.ErrorCode
}

func (connector *ConnectorImpl) SetNotificationChannel(notificationChannel chan<- rxgo.Item) {
	connector.ConnectorNotificationChannel = notificationChannel
}
