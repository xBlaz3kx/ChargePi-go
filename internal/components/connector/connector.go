package connector

import (
	"errors"
	"fmt"
	"github.com/lorenzodonini/ocpp-go/ocpp1.6/core"
	"github.com/lorenzodonini/ocpp-go/ocpp1.6/types"
	"github.com/reactivex/rxgo/v2"
	log "github.com/sirupsen/logrus"
	"github.com/xBlaz3kx/ChargePi-go/internal/components/hardware/evcc"
	"github.com/xBlaz3kx/ChargePi-go/internal/components/hardware/power-meter"
	"github.com/xBlaz3kx/ChargePi-go/internal/components/settings"
	"github.com/xBlaz3kx/ChargePi-go/internal/models"
	"github.com/xBlaz3kx/ChargePi-go/internal/models/session"
	settingsModel "github.com/xBlaz3kx/ChargePi-go/internal/models/settings"
	"github.com/xBlaz3kx/ChargePi-go/pkg/scheduler"
	"github.com/xBlaz3kx/ChargePi-go/pkg/util"
	ocppConfigManager "github.com/xBlaz3kx/ocppManager-go"
	v16 "github.com/xBlaz3kx/ocppManager-go/v16"
	"sync"
	"time"
)

var (
	ErrInvalidEvseId            = errors.New("invalid evse id")
	ErrInvalidConnectorId       = errors.New("invalid connector id")
	ErrInvalidReservationId     = errors.New("invalid reservation id")
	ErrInvalidConnectorStatus   = errors.New("invalid connector status")
	ErrSessionTimeLimitExceeded = errors.New("session time limit exceeded")
	ErrNotCharging              = errors.New("connector not charging")
)

type (
	connectorImpl struct {
		mu                           sync.Mutex
		EvseId                       int
		ConnectorId                  int
		ConnectorType                string
		ConnectorStatus              core.ChargePointStatus
		ErrorCode                    core.ChargePointErrorCode
		evcc                         evcc.EVCC
		powerMeter                   powerMeter.PowerMeter
		PowerMeterEnabled            bool
		MaxChargingTime              int
		reservationId                int
		session                      *session.Session
		ConnectorNotificationChannel chan<- rxgo.Item
		meterValuesChannel           chan<- models.MeterValueNotification
	}

	Connector interface {
		StartCharging(transactionId string, tagId string) error
		ResumeCharging(session session.Session) (error, int)
		StopCharging(reason core.Reason) error
		SetNotificationChannel(notificationChannel chan<- rxgo.Item)
		SetMeterValuesChannel(notificationChannel chan<- models.MeterValueNotification)
		ReserveConnector(reservationId int, tagId string) error
		RemoveReservation() error
		GetReservationId() int
		GetTagId() string
		GetTransactionId() string
		GetConnectorId() int
		GetEvseId() int
		CalculateSessionAvgEnergyConsumption() float64
		SamplePowerMeter(measurands []types.Measurand)
		SetStatus(status core.ChargePointStatus, errCode core.ChargePointErrorCode)
		GetStatus() (core.ChargePointStatus, core.ChargePointErrorCode)
		IsAvailable() bool
		IsPreparing() bool
		IsCharging() bool
		IsReserved() bool
		IsUnavailable() bool
		GetPowerMeter() powerMeter.PowerMeter
		GetMaxChargingTime() int
	}
)

// NewConnector Create a new connector object from the provided arguments. EvseId, connectorId and maxChargingTime must be greater than zero.
// When created, it makes an empty session, turns off the evcc and defaults the status to Available.
func NewConnector(
	evseId int,
	connectorId int,
	connectorType string,
	evcc evcc.EVCC,
	powerMeter powerMeter.PowerMeter,
	powerMeterEnabled bool,
	maxChargingTime int,
) (*connectorImpl, error) {
	log.WithFields(log.Fields{
		"evseId":          evseId,
		"connectorId":     connectorId,
		"type":            connectorType,
		"maxChargingTime": maxChargingTime,
		"hasPowerMeter":   powerMeterEnabled,
	}).Info("Creating a new connector")

	if maxChargingTime <= 0 {
		maxChargingTime = 180
	}

	if evseId <= 0 {
		return nil, ErrInvalidEvseId
	}

	if connectorId <= 0 {
		return nil, ErrInvalidConnectorId
	}

	evcc.DisableCharging()
	return &connectorImpl{
		mu:                sync.Mutex{},
		EvseId:            evseId,
		ConnectorId:       connectorId,
		ConnectorType:     connectorType,
		evcc:              evcc,
		powerMeter:        powerMeter,
		reservationId:     -1,
		PowerMeterEnabled: powerMeterEnabled,
		MaxChargingTime:   maxChargingTime,
		ConnectorStatus:   core.ChargePointStatusAvailable,
		session:           session.NewEmptySession(),
	}, nil
}

// StartCharging Start charging a connector if connector is available and session could be started.
// It turns on the evcc (even if negative logic applies).
func (connector *connectorImpl) StartCharging(transactionId string, tagId string) error {
	logInfo := log.WithFields(log.Fields{
		"evseId":        connector.EvseId,
		"connectorId":   connector.ConnectorId,
		"transactionId": transactionId,
		"tagId":         tagId,
	})
	logInfo.Debugf("Trying to start charging on connector")

	if !(connector.IsAvailable() || connector.IsPreparing()) {
		return ErrInvalidConnectorStatus
	}

	connector.SetStatus(core.ChargePointStatusPreparing, core.NoError)
	sessionErr := connector.session.StartSession(transactionId, tagId)
	if sessionErr != nil {
		return sessionErr
	}

	err := connector.evcc.EnableCharging()
	if err != nil {
		return err
	}

	connector.SetStatus(core.ChargePointStatusCharging, core.NoError)

	settings.UpdateConnectorSessionInfo(
		connector.EvseId,
		connector.ConnectorId,
		&settingsModel.Session{
			IsActive:      connector.session.IsActive,
			TagId:         connector.session.TagId,
			TransactionId: connector.session.TransactionId,
			Started:       connector.session.Started,
			Consumption:   connector.session.Consumption,
		})

	if connector.PowerMeterEnabled && connector.GetPowerMeter() != nil {
		sampleError := connector.preparePowerMeterAtConnector()
		if sampleError != nil {
			logInfo.Errorf("Cannot sample connector: %v", sampleError)
		}
	}

	return nil
}

// ResumeCharging Resumes or restores the charging state after boot if a charging session was active.
func (connector *connectorImpl) ResumeCharging(session session.Session) (err error, chargingTimeElapsed int) {
	// Set the transaction id so connector is able to stop the transaction if charging fails
	logInfo := log.WithFields(log.Fields{
		"evseId":      connector.EvseId,
		"connectorId": connector.ConnectorId,
		"session":     session,
	})
	logInfo.Debugf("Trying to resume charging on connector")

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
			return fmt.Errorf("cannot resume session: %v", sessionErr), connector.MaxChargingTime
		}

		err = connector.evcc.EnableCharging()
		if err != nil {
			return err, 0
		}

		connector.session.Started = session.Started
		connector.session.Consumption = append(connector.session.Consumption, session.Consumption...)
		return nil, chargingTimeElapsed
	}

	return ErrInvalidConnectorStatus, connector.MaxChargingTime
}

// StopCharging Stops charging the connector by turning the evcc off and ending the session.
func (connector *connectorImpl) StopCharging(reason core.Reason) error {
	logInfo := log.WithFields(log.Fields{
		"evseId":      connector.EvseId,
		"connectorId": connector.ConnectorId,
		"reason":      reason,
	})

	if connector.IsCharging() || connector.IsPreparing() {
		logInfo.Debugf("Stopping charging")
		connector.session.EndSession()
		connector.evcc.DisableCharging()

		settings.UpdateConnectorSessionInfo(
			connector.EvseId,
			connector.ConnectorId,
			&settingsModel.Session{
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
func (connector *connectorImpl) SamplePowerMeter(measurands []types.Measurand) {
	logInfo := log.WithFields(log.Fields{
		"evseId":      connector.EvseId,
		"connectorId": connector.ConnectorId,
	})

	if !connector.PowerMeterEnabled || util.IsNilInterfaceOrPointer(connector.powerMeter) {
		return
	}

	logInfo.Debugf("Sampling connector %v", measurands)
	var (
		meterValues []types.MeterValue
		samples     []types.SampledValue
		value       = 0.0
	)

	for _, measurand := range measurands {

		switch measurand {
		case types.MeasurandEnergyActiveImportInterval, types.MeasurandEnergyActiveImportRegister,
			types.MeasurandEnergyActiveExportInterval, types.MeasurandEnergyActiveExportRegister:
			value = connector.powerMeter.GetEnergy()
			break
		case types.MeasurandCurrentImport, types.MeasurandCurrentExport:
			value = connector.powerMeter.GetCurrent()
			break
		case types.MeasurandPowerActiveImport, types.MeasurandPowerActiveExport:
			value = connector.powerMeter.GetPower()
			break
		case types.MeasurandVoltage:
			value = connector.powerMeter.GetVoltage()
			break
		}

		if value != 0.0 {
			sample := types.SampledValue{
				Value:     fmt.Sprintf("%.3f", value),
				Measurand: measurand,
			}

			meterValues = append(meterValues, types.MeterValue{SampledValue: []types.SampledValue{sample}, Timestamp: types.NewDateTime(time.Now())})
		}
	}

	if connector.meterValuesChannel != nil {
		connector.meterValuesChannel <- models.NewMeterValueNotification(
			connector.EvseId,
			connector.ConnectorId,
			nil,
			meterValues...,
		)
	}

	connector.session.AddSampledValue(samples)
}

// preparePowerMeterAtConnector
func (connector *connectorImpl) preparePowerMeterAtConnector() error {
	var (
		measurands          = util.GetTypesToSample()
		sampleTime          = "10s"
		sampleInterval, err = ocppConfigManager.GetConfigurationValue(v16.MeterValueSampleInterval.String())
		jobTag              = fmt.Sprintf("Evse%dConnector%dSampling", connector.EvseId, connector.ConnectorId)
	)
	if err != nil {
		sampleInterval = "10"
	}

	sampleTime = fmt.Sprintf("%ss", sampleInterval)
	// Schedule the sampling
	_, err = scheduler.GetScheduler().Every(sampleTime).
		Tag(jobTag).
		Do(connector.SamplePowerMeter, measurands)

	return err
}

func (connector *connectorImpl) IsAvailable() bool {
	connector.mu.Lock()
	defer connector.mu.Unlock()
	return connector.ConnectorStatus == core.ChargePointStatusAvailable
}
func (connector *connectorImpl) IsCharging() bool {
	connector.mu.Lock()
	defer connector.mu.Unlock()
	return connector.ConnectorStatus == core.ChargePointStatusCharging
}
func (connector *connectorImpl) IsPreparing() bool {
	connector.mu.Lock()
	defer connector.mu.Unlock()
	return connector.ConnectorStatus == core.ChargePointStatusPreparing
}
func (connector *connectorImpl) IsReserved() bool {
	connector.mu.Lock()
	defer connector.mu.Unlock()
	return connector.ConnectorStatus == core.ChargePointStatusReserved
}
func (connector *connectorImpl) IsUnavailable() bool {
	connector.mu.Lock()
	defer connector.mu.Unlock()
	return connector.ConnectorStatus == core.ChargePointStatusUnavailable
}

func (connector *connectorImpl) SetStatus(status core.ChargePointStatus, errCode core.ChargePointErrorCode) {
	logInfo := log.WithFields(log.Fields{
		"evseId":      connector.EvseId,
		"connectorId": connector.ConnectorId,
	})
	logInfo.Debugf("Setting connector status %s with err %s", status, errCode)

	connector.mu.Lock()
	connector.ConnectorStatus = status
	connector.ErrorCode = errCode
	settings.UpdateConnectorStatus(connector.EvseId, connector.ConnectorId, status)

	if connector.ConnectorNotificationChannel != nil {
		connector.ConnectorNotificationChannel <- rxgo.Of(connector)
	}

	connector.mu.Unlock()
}

func (connector *connectorImpl) GetTransactionId() string {
	return connector.session.TransactionId
}
func (connector *connectorImpl) GetTagId() string {
	return connector.session.TagId
}

func (connector *connectorImpl) ReserveConnector(reservationId int, tagId string) error {
	logInfo := log.WithFields(log.Fields{
		"evseId":      connector.EvseId,
		"connectorId": connector.ConnectorId,
		"tagId":       tagId,
	})
	logInfo.Debugf("Reserving connector for id %d", reservationId)

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
func (connector *connectorImpl) RemoveReservation() error {
	if !connector.IsReserved() {
		return ErrInvalidConnectorStatus
	}

	logInfo := log.WithFields(log.Fields{
		"evseId":      connector.EvseId,
		"connectorId": connector.ConnectorId,
	})
	logInfo.Debugf("Removing reservation")

	connector.reservationId = -1
	connector.SetStatus(core.ChargePointStatusAvailable, core.NoError)
	return nil
}

func (connector *connectorImpl) GetReservationId() int {
	return connector.reservationId
}

func (connector *connectorImpl) GetConnectorId() int {
	return connector.ConnectorId
}

func (connector *connectorImpl) GetEvseId() int {
	return connector.EvseId
}

func (connector *connectorImpl) CalculateSessionAvgEnergyConsumption() float64 {
	return connector.session.CalculateEnergyConsumptionWithAvgPower()
}

func (connector *connectorImpl) GetPowerMeter() powerMeter.PowerMeter {
	return connector.powerMeter
}

func (connector *connectorImpl) GetMaxChargingTime() int {
	return connector.MaxChargingTime
}

func (connector *connectorImpl) GetStatus() (core.ChargePointStatus, core.ChargePointErrorCode) {
	return connector.ConnectorStatus, connector.ErrorCode
}

func (connector *connectorImpl) SetNotificationChannel(notificationChannel chan<- rxgo.Item) {
	connector.ConnectorNotificationChannel = notificationChannel
}

func (connector *connectorImpl) SetMeterValuesChannel(notificationChannel chan<- models.MeterValueNotification) {
	connector.meterValuesChannel = notificationChannel
}
