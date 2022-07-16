package test

import (
	"github.com/lorenzodonini/ocpp-go/ocpp1.6/core"
	"github.com/lorenzodonini/ocpp-go/ocpp1.6/types"
	"github.com/reactivex/rxgo/v2"
	"github.com/stretchr/testify/mock"
	"github.com/xBlaz3kx/ChargePi-go/internal/components/connector"
	powerMeter "github.com/xBlaz3kx/ChargePi-go/internal/components/hardware/power-meter"
	"github.com/xBlaz3kx/ChargePi-go/internal/models"
	"github.com/xBlaz3kx/ChargePi-go/internal/models/session"
	"github.com/xBlaz3kx/ChargePi-go/internal/models/settings"
)

type (
	ConnectorMock struct {
		mock.Mock
	}

	ManagerMock struct {
		mock.Mock
	}
)

/*------------------ Manager mock ------------------*/

func (o *ManagerMock) GetConnectors() []connector.Connector {
	return o.Called().Get(0).([]connector.Connector)
}

func (o *ManagerMock) FindConnector(evseId, connectorID int) connector.Connector {
	args := o.Called(evseId, connectorID)
	if args.Get(0) != nil {
		return args.Get(0).(connector.Connector)
	}

	return nil
}

func (o *ManagerMock) FindAvailableConnector() connector.Connector {
	args := o.Called()
	if args.Get(0) != nil {
		return args.Get(0).(connector.Connector)
	}

	return nil
}

func (o *ManagerMock) FindConnectorWithTagId(tagId string) connector.Connector {
	args := o.Called(tagId)
	if args.Get(0) != nil {
		return args.Get(0).(connector.Connector)
	}

	return nil
}

func (o *ManagerMock) FindConnectorWithTransactionId(transactionId string) connector.Connector {
	args := o.Called(transactionId)
	if args.Get(0) != nil {
		return args.Get(0).(connector.Connector)
	}

	return nil
}

func (o *ManagerMock) FindConnectorWithReservationId(reservationId int) connector.Connector {
	args := o.Called(reservationId)
	if args.Get(0) != nil {
		return args.Get(0).(connector.Connector)
	}

	return nil
}

func (o *ManagerMock) StartChargingConnector(evseId, connectorID int, tagId, transactionId string) error {
	return o.Called(evseId, connectorID, tagId, transactionId).Error(0)
}

func (o *ManagerMock) StopChargingConnector(tagId, transactionId string, reason core.Reason) error {
	return o.Called(tagId, transactionId).Error(0)
}

func (o *ManagerMock) StopAllConnectors(reason core.Reason) error {
	return o.Called().Error(0)
}

func (o *ManagerMock) AddConnector(c connector.Connector) error {
	return o.Called(c).Error(0)
}

func (o *ManagerMock) AddConnectorFromSettings(maxChargingTime int, c *settings.Connector) error {
	return o.Called(c).Error(0)
}

func (o *ManagerMock) AddConnectorsFromConfiguration(maxChargingTime int, c []*settings.Connector) error {
	return o.Called(c).Error(0)
}

func (o *ManagerMock) RestoreConnectorStatus(s *settings.Connector) error {
	return o.Called(s).Error(0)
}

func (o *ManagerMock) SetNotificationChannel(notificationChannel chan rxgo.Item) {
	o.Called()
}

func (o *ManagerMock) SetMeterValuesChannel(notificationChannel chan models.MeterValueNotification) {
	o.Called()
}

/*------------------ Connector mock ------------------*/

func (m *ConnectorMock) StartCharging(transactionId string, tagId string) error {
	args := m.Called(transactionId, tagId)
	return args.Error(0)
}

func (m *ConnectorMock) ResumeCharging(session session.Session) (error, int) {
	args := m.Called(session)
	return args.Error(0), args.Int(1)
}

func (m *ConnectorMock) StopCharging(reason core.Reason) error {
	args := m.Called(reason)
	return args.Error(0)
}

func (m *ConnectorMock) SetNotificationChannel(notificationChannel chan<- rxgo.Item) {
	m.Called(notificationChannel)
}

func (m *ConnectorMock) SetMeterValuesChannel(notificationChannel chan<- models.MeterValueNotification) {
	m.Called(notificationChannel)
}

func (m *ConnectorMock) ReserveConnector(reservationId int, tagId string) error {
	args := m.Called(reservationId, tagId)
	return args.Error(0)
}

func (m *ConnectorMock) RemoveReservation() error {
	args := m.Called()
	return args.Error(0)
}

func (m *ConnectorMock) GetReservationId() int {
	args := m.Called()
	return args.Int(0)
}

func (m *ConnectorMock) GetTagId() string {
	args := m.Called()
	return args.String(0)
}

func (m *ConnectorMock) GetTransactionId() string {
	args := m.Called()
	return args.String(0)
}

func (m *ConnectorMock) GetConnectorId() int {
	args := m.Called()
	return args.Int(0)
}

func (m *ConnectorMock) GetEvseId() int {
	args := m.Called()
	return args.Int(0)
}

func (m *ConnectorMock) CalculateSessionAvgEnergyConsumption() float64 {
	args := m.Called()
	return args.Get(0).(float64)
}

func (m *ConnectorMock) SamplePowerMeter(measurands []types.Measurand) {
	m.Called(measurands)
}

func (m *ConnectorMock) SetStatus(status core.ChargePointStatus, errCode core.ChargePointErrorCode) {
	m.Called(status, errCode)
}

func (m *ConnectorMock) GetStatus() (core.ChargePointStatus, core.ChargePointErrorCode) {
	args := m.Called()
	return core.ChargePointStatus(args.String(0)), core.ChargePointErrorCode(args.String(1))
}

func (m *ConnectorMock) IsAvailable() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *ConnectorMock) IsPreparing() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *ConnectorMock) IsCharging() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *ConnectorMock) IsReserved() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *ConnectorMock) IsUnavailable() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *ConnectorMock) GetPowerMeter() powerMeter.PowerMeter {
	args := m.Called()
	return args.Get(0).(powerMeter.PowerMeter)
}

func (m *ConnectorMock) GetMaxChargingTime() int {
	args := m.Called()
	return args.Int(0)
}
