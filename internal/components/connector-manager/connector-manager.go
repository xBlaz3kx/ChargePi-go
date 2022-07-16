package connectorManager

import (
	"errors"
	"fmt"
	"github.com/lorenzodonini/ocpp-go/ocpp1.6/core"
	"github.com/reactivex/rxgo/v2"
	log "github.com/sirupsen/logrus"
	"github.com/xBlaz3kx/ChargePi-go/internal/components/connector"
	"github.com/xBlaz3kx/ChargePi-go/internal/components/hardware/evcc"
	powerMeter "github.com/xBlaz3kx/ChargePi-go/internal/components/hardware/power-meter"
	"github.com/xBlaz3kx/ChargePi-go/internal/models"
	"github.com/xBlaz3kx/ChargePi-go/internal/models/session"
	"github.com/xBlaz3kx/ChargePi-go/internal/models/settings"
	"github.com/xBlaz3kx/ChargePi-go/pkg/util"
	"sync"
)

var (
	ErrConnectorNotFound      = errors.New("connector not found")
	ErrConnectorAlreadyExists = errors.New("connector already exists")
	ErrConnectorStatusInvalid = errors.New("connector status invalid")
	ErrConnectorNil           = errors.New("connector is nil")

	manager Manager
	once    = sync.Once{}
)

type (
	Manager interface {
		GetConnectors() []connector.Connector
		FindConnector(evseId, connectorID int) connector.Connector
		FindAvailableConnector() connector.Connector
		FindConnectorWithTagId(tagId string) connector.Connector
		FindConnectorWithTransactionId(transactionId string) connector.Connector
		FindConnectorWithReservationId(reservationId int) connector.Connector
		StartChargingConnector(evseId, connectorID int, tagId, transactionId string) error
		StopChargingConnector(tagId, transactionId string, reason core.Reason) error
		StopAllConnectors(reason core.Reason) error
		AddConnector(c connector.Connector) error
		AddConnectorFromSettings(maxChargingTime int, c *settings.Connector) error
		AddConnectorsFromConfiguration(maxChargingTime int, c []*settings.Connector) error
		RestoreConnectorStatus(*settings.Connector) error
		SetNotificationChannel(notificationChannel chan rxgo.Item)
		SetMeterValuesChannel(notificationChannel chan models.MeterValueNotification)
	}

	managerImpl struct {
		connectors          sync.Map
		notificationChannel chan rxgo.Item
		meterValuesChannel  chan models.MeterValueNotification
	}
)

func init() {
	once.Do(func() {
		GetManager()
	})
}

func GetManager() Manager {
	if manager == nil {
		log.Debug("Creating connector manager")
		manager = NewManager(nil)
	}

	return manager
}

func NewManager(notificationChannel chan rxgo.Item) Manager {
	return &managerImpl{
		connectors:          sync.Map{},
		notificationChannel: notificationChannel,
	}
}

func (m *managerImpl) GetConnectors() []connector.Connector {
	var connectors []connector.Connector

	m.connectors.Range(func(key, value interface{}) bool {
		c, canCast := value.(connector.Connector)
		if canCast {
			connectors = append(connectors, c)
		}
		return true
	})

	return connectors
}

func (m *managerImpl) SetNotificationChannel(notificationChannel chan rxgo.Item) {
	if notificationChannel != nil {
		m.notificationChannel = notificationChannel
	}
}

func (m *managerImpl) SetMeterValuesChannel(notificationChannel chan models.MeterValueNotification) {
	if notificationChannel != nil {
		m.meterValuesChannel = notificationChannel
	}
}

func (m *managerImpl) FindConnector(evseId, connectorID int) connector.Connector {
	var (
		key        = fmt.Sprintf("Evse%dConnector%d", evseId, connectorID)
		c, isFound = m.connectors.Load(key)
	)

	if isFound {
		return c.(connector.Connector)
	}

	return nil
}

func (m *managerImpl) FindAvailableConnector() connector.Connector {
	var availableConnector connector.Connector

	m.connectors.Range(func(key, value interface{}) bool {
		c, canCast := value.(connector.Connector)
		// todo check if there is no connector with the same EVSE charging
		if canCast && c.IsAvailable() {
			availableConnector = c
			return false
		}

		return true
	})

	return availableConnector
}

func (m *managerImpl) FindConnectorWithTagId(tagId string) connector.Connector {
	var connectorWithTag connector.Connector

	m.connectors.Range(func(key, value interface{}) bool {
		c, _ := value.(connector.Connector)
		if c.GetTagId() == tagId {
			connectorWithTag = c
			return false
		}

		return true
	})

	return connectorWithTag
}

func (m *managerImpl) FindConnectorWithTransactionId(transactionId string) connector.Connector {
	var connectorWithTransaction connector.Connector

	m.connectors.Range(func(key, value interface{}) bool {
		c, _ := value.(connector.Connector)
		if c.GetTransactionId() == transactionId {
			connectorWithTransaction = c
			return false
		}

		return true
	})

	return connectorWithTransaction
}

func (m *managerImpl) FindConnectorWithReservationId(reservationId int) connector.Connector {
	var connectorWithReservation connector.Connector

	m.connectors.Range(func(key, value interface{}) bool {
		c, _ := value.(connector.Connector)
		if c.GetReservationId() == reservationId {
			connectorWithReservation = c
			return false
		}

		return true
	})

	return connectorWithReservation
}

func (m *managerImpl) StartChargingConnector(evseId, connectorID int, tagId, transactionId string) error {
	c := m.FindConnector(evseId, connectorID)

	if c != nil {
		return c.StartCharging(transactionId, tagId)
	}

	return ErrConnectorNotFound
}

func (m *managerImpl) StopChargingConnector(tagId, transactionId string, reason core.Reason) error {
	c := m.FindConnectorWithTransactionId(transactionId)

	if c != nil {
		return c.StopCharging(reason)
	}

	return ErrConnectorNotFound
}

func (m *managerImpl) StopAllConnectors(reason core.Reason) error {
	log.Debugf("Stopping all connectors: %s", reason)

	var err error

	for _, c := range m.GetConnectors() {
		stopErr := c.StopCharging(reason)
		if stopErr != nil {
			err = stopErr
		}
	}

	return err
}

func (m *managerImpl) AddConnector(c connector.Connector) error {
	if util.IsNilInterfaceOrPointer(c) {
		return ErrConnectorNil
	}

	var (
		logInfo = log.WithFields(log.Fields{
			"evseId":      c.GetEvseId(),
			"connectorId": c.GetConnectorId(),
		})
		key = fmt.Sprintf("Evse%dConnector%d", c.GetEvseId(), c.GetConnectorId())
	)

	logInfo.Debugf("Adding a connector to manager")
	c.SetNotificationChannel(m.notificationChannel)
	c.SetMeterValuesChannel(m.meterValuesChannel)

	// Add the connector
	_, isLoaded := m.connectors.LoadOrStore(key, c)
	if isLoaded {
		return ErrConnectorAlreadyExists
	}

	return nil
}

func (m *managerImpl) AddConnectorFromSettings(maxChargingTime int, c *settings.Connector) error {
	if util.IsNilInterfaceOrPointer(c) {
		return ErrConnectorNil
	}

	var (
		relay, err = evcc.NewEVCCFromType(c.EVCC)
		// Create a PowerMeter from connector settings
		meter, powerMeterErr = powerMeter.NewPowerMeter(c.PowerMeter)
	)

	if err != nil {
		log.Fatalf("Cannot create EVCC from settings")
	}

	if powerMeterErr != nil {
		log.Warnf("Cannot instantiate power meter: %s", powerMeterErr)
	}

	// Create a new connector
	connectorObj, err := connector.NewConnector(
		c.EvseId,
		c.ConnectorId,
		c.Type,
		relay,
		meter,
		c.PowerMeter.Enabled,
		maxChargingTime,
	)
	if err != nil {
		return err
	}

	return m.AddConnector(connectorObj)
}

func (m *managerImpl) AddConnectorsFromConfiguration(maxChargingTime int, connectors []*settings.Connector) error {

	for _, c := range connectors {
		addErr := m.AddConnectorFromSettings(maxChargingTime, c)
		if addErr != nil {
			return addErr
		}
	}

	return nil
}

func (m *managerImpl) RestoreConnectorStatus(c *settings.Connector) error {
	if util.IsNilInterfaceOrPointer(c) {
		return ErrConnectorNotFound
	}

	logInfo := log.WithFields(log.Fields{
		"evseId":         c.EvseId,
		"connectorId":    c.ConnectorId,
		"previousStatus": c.Status,
		"session":        c.Session,
	})
	logInfo.Debugf("Attempting to restore connector status")

	var (
		err  error
		conn = m.FindConnector(c.EvseId, c.ConnectorId)
	)

	if conn == nil {
		return ErrConnectorNotFound
	}

	// Set the previous status to determine what action to do
	connectorPreviousStatus := core.ChargePointStatus(c.Status)
	conn.SetStatus(connectorPreviousStatus, core.NoError)

	switch connectorPreviousStatus {
	case core.ChargePointStatusAvailable:
		return nil
	case core.ChargePointStatusReserved,
		core.ChargePointStatusFinishing,
		core.ChargePointStatusSuspendedEV,
		core.ChargePointStatusSuspendedEVSE:
		return nil
	case core.ChargePointStatusPreparing:
		err = conn.StartCharging(c.Session.TransactionId, c.Session.TagId)
		if err != nil {
			conn.SetStatus(core.ChargePointStatusAvailable, core.InternalError)
		}

		return err
	case core.ChargePointStatusCharging:
		err, timeLeft := conn.ResumeCharging(session.Session(c.Session))
		if err != nil {
			logInfo.Errorf("Resume charging failed, reason: %v", err)

			// Attempt to stop charging
			err = conn.StopCharging(core.ReasonDeAuthorized)
			if err != nil {
				logInfo.Errorf("Err stopping the charging: %v", err)
				conn.SetStatus(core.ChargePointStatusFaulted, core.InternalError)
			}
		}
		//todo figure out what if time left
		log.Println(timeLeft)

		logInfo.Debugf("Successfully resumed charging")
		return nil
	case core.ChargePointStatusFaulted:
		conn.SetStatus(core.ChargePointStatusFaulted, core.InternalError)
		return nil
	default:
		return ErrConnectorStatusInvalid
	}
}
