package connector

import (
	"github.com/lorenzodonini/ocpp-go/ocpp1.6/core"
	"github.com/lorenzodonini/ocpp-go/ocpp1.6/types"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/xBlaz3kx/ChargePi-go/internal/models"
	"github.com/xBlaz3kx/ChargePi-go/internal/models/session"
	"github.com/xBlaz3kx/ChargePi-go/test"
	"golang.org/x/net/context"
	"os/exec"
	"testing"
	"time"
)

const (
	fileName = "./connector-1.json"
)

type (
	ConnectorTestSuite struct {
		suite.Suite
		connector      *connectorImpl
		relayPinNum    int
		relayMock      *test.EvccMock
		powerMeterMock *test.PowerMeterMock
	}
)

/*---------------------- Test suite ----------------------*/

func NewConnectorTestSuite() *ConnectorTestSuite {
	return &ConnectorTestSuite{
		relayPinNum: 15,
	}
}

func (s *ConnectorTestSuite) SetupTest() {
	cmd := exec.Command("touch", fileName)
	err := cmd.Run()
	s.Require().NoError(err)

	s.relayMock = new(test.EvccMock)
	s.powerMeterMock = new(test.PowerMeterMock)

	s.relayMock.On("Enable").Return()
	s.relayMock.On("Disable").Return()

	// Create a new connector
	connector, err := NewConnector(
		1,
		1,
		"Schuko",
		s.relayMock,
		s.powerMeterMock,
		false,
		15,
	)
	s.Require().NoError(err)

	s.connector = connector
}

func (s *ConnectorTestSuite) TestCreateNewConnector() {
	s.relayMock.On("Disable").Return()
	s.relayMock.On("Enable").Return()

	// Ok case
	connector1, err := NewConnector(
		1,
		1,
		"Schuko",
		s.relayMock,
		s.powerMeterMock,
		false,
		15,
	)
	s.Require().Equal(1, connector1.ConnectorId)
	s.Require().Equal(1, connector1.EvseId)
	s.Require().Equal(core.ChargePointStatusAvailable, connector1.ConnectorStatus)
	s.Require().Equal("Schuko", connector1.ConnectorType)
	s.Require().Equal(15, connector1.MaxChargingTime)
	s.Require().False(connector1.PowerMeterEnabled)

	// Invalid evseId
	_, err = NewConnector(0, 1, "Schuko",
		s.relayMock, new(test.PowerMeterMock), false, 15)
	s.Require().Error(err)

	// Invalid connectorId
	_, err = NewConnector(1, 0, "Schuko",
		s.relayMock, new(test.PowerMeterMock), false, 15)
	s.Require().Error(err)

	// Negative connector id
	_, err = NewConnector(1, -1, "Schuko",
		s.relayMock, new(test.PowerMeterMock), false, 15)
	s.Require().Error(err)
}

func (s *ConnectorTestSuite) TestStartCharging() {
	// Ok case
	err := s.connector.StartCharging("1234", "exampleTag")
	s.Require().NoError(err)
	s.relayMock.AssertCalled(s.T(), "Enable")

	// Cannot start new session on a connector that is already charging
	err = s.connector.StartCharging("1234a", "exampleTag1")
	s.Require().Error(err)
	//s.relayMock.AssertNotCalled(s.T(), "Enable")

	err = s.connector.StopCharging(core.ReasonLocal)
	s.Require().NoError(err)

	// Invalid transaction and tag id
	err = s.connector.StartCharging("@1234asd", "~ˇˇ3123")
	s.Require().Error(err)
	//s.relayMock.AssertNotCalled(s.T(), "Enable")

	// Invalid transaction id
	err = s.connector.StartCharging("", "1234")
	s.Require().Error(err)
	//s.relayMock.AssertNotCalled(s.T(), "Enable")

	// Invalid tag id
	err = s.connector.StartCharging("1234", "")
	s.Require().Error(err)
	//s.relayMock.AssertNotCalled(s.T(), "Enable")

	// Invalid connector status
	s.connector.SetStatus(core.ChargePointStatusUnavailable, core.InternalError)
	err = s.connector.StartCharging("1234a", "1234a")
	s.Require().Error(err)
}

func (s *ConnectorTestSuite) TestStopCharging() {
	// Start charging
	err := s.connector.StartCharging("1234", "1234")
	s.Require().NoError(err)
	s.relayMock.AssertCalled(s.T(), "Enable")

	// Stop charging normally
	err = s.connector.StopCharging(core.ReasonLocal)
	s.Require().NoError(err)
	s.relayMock.AssertCalled(s.T(), "Disable")

	// Cannot stop charging if the connector is available
	err = s.connector.StopCharging(core.ReasonLocal)
	s.Require().Error(err)
	//s.relayMock.AssertNotCalled(s.T(), "Disable")
}

func (s *ConnectorTestSuite) TestResumeCharging() {
	var (
		maxChargingTime = s.connector.GetMaxChargingTime()
		validSession    = session.Session{
			IsActive:      true,
			TransactionId: "1234",
			TagId:         "1234",
			Started:       time.Now().Format(time.RFC3339),
			Consumption:   nil,
		}

		expiredSession = session.Session{
			IsActive:      true,
			TransactionId: "1234",
			TagId:         "1234",
			Started:       time.Date(2021, 1, 1, 1, 1, 1, 1, time.Local).Format(time.RFC3339),
			Consumption:   nil,
		}

		invalidSession = session.Session{
			IsActive:      false,
			TransactionId: "",
			TagId:         "",
			Started:       "",
			Consumption:   nil,
		}
	)

	// Ok case
	s.connector.SetStatus(core.ChargePointStatusCharging, core.NoError)
	err, timeElapsed := s.connector.ResumeCharging(validSession)
	s.Require().NoError(err)
	s.Require().InDelta(0, timeElapsed, 1)

	err = s.connector.StopCharging(core.ReasonLocal)
	s.Require().NoError(err)

	// Invalid session
	s.connector.SetStatus(core.ChargePointStatusCharging, core.NoError)
	err, timeElapsed = s.connector.ResumeCharging(invalidSession)
	s.Require().Error(err)
	s.Require().InDelta(maxChargingTime, timeElapsed, 1)

	// Invalid connector status
	s.connector.SetStatus(core.ChargePointStatusAvailable, core.NoError)
	err, timeElapsed = s.connector.ResumeCharging(validSession)
	s.Require().Error(err)
	s.Require().InDelta(maxChargingTime, timeElapsed, 1)

	// Expired session
	s.connector.SetStatus(core.ChargePointStatusAvailable, core.NoError)
	err, timeElapsed = s.connector.ResumeCharging(expiredSession)
	s.Require().Error(err)
	s.Require().InDelta(maxChargingTime, timeElapsed, 1)

	// Invalid status
	s.connector.SetStatus(core.ChargePointStatusUnavailable, core.EVCommunicationError)
	err, timeElapsed = s.connector.ResumeCharging(invalidSession)
	s.Require().Error(err)
	s.Require().InDelta(maxChargingTime, timeElapsed, 1)
}

func (s *ConnectorTestSuite) TestReserveConnector() {
	// Ok case
	err := s.connector.ReserveConnector(1, "")
	s.Require().NoError(err)

	// Connector already reserved
	err = s.connector.ReserveConnector(2, "")
	s.Require().Error(err)

	err = s.connector.RemoveReservation()
	s.Require().NoError(err)

	// Invalid connector status
	s.connector.SetStatus(core.ChargePointStatusCharging, core.NoError)
	err = s.connector.ReserveConnector(2, "")
	s.Require().Error(err)
}

func (s *ConnectorTestSuite) TestRemoveReservation() {
	// Make a reservation
	err := s.connector.ReserveConnector(1, "")
	s.Require().NoError(err)

	// Ok case
	err = s.connector.RemoveReservation()
	s.Require().NoError(err)

	// Cannot remove reservation that is not there
	err = s.connector.RemoveReservation()
	s.Require().Error(err)
}

func (s *ConnectorTestSuite) TestSamplePowerMeter() {
	s.powerMeterMock = new(test.PowerMeterMock)

	s.powerMeterMock.On("GetEnergy").Return(1.0)
	s.powerMeterMock.On("GetCurrent").Return(1.0)
	s.powerMeterMock.On("GetVoltage").Return(1.0)

	var (
		ctx, cancel    = context.WithTimeout(context.Background(), time.Second*30)
		meterValueChan = make(chan models.MeterValueNotification)
	)

	defer cancel()
	go func() {
	Loop:
		for {
			select {
			case notif := <-meterValueChan:
				s.Assert().EqualValues(s.connector.GetConnectorId(), notif.ConnectorId)
				s.Assert().EqualValues(s.connector.GetEvseId(), notif.EvseId)

				s.Assert().Len(notif.MeterValues, 3)
				s.Assert().EqualValues("1.000", notif.MeterValues[0].SampledValue[0].Value)
				s.Assert().EqualValues(types.MeasurandVoltage, notif.MeterValues[0].SampledValue[0].Measurand)

				s.Assert().EqualValues("1.000", notif.MeterValues[1].SampledValue[0].Value)
				s.Assert().EqualValues(types.MeasurandCurrentImport, notif.MeterValues[1].SampledValue[0].Measurand)

				s.Assert().EqualValues("1.000", notif.MeterValues[2].SampledValue[0].Value)
				s.Assert().EqualValues(types.MeasurandEnergyActiveImportInterval, notif.MeterValues[2].SampledValue[0].Measurand)
			case <-ctx.Done():
				break Loop
			}
		}
	}()

	s.connector.SetMeterValuesChannel(meterValueChan)
	s.connector.PowerMeterEnabled = true
	s.connector.powerMeter = s.powerMeterMock
	s.connector.SamplePowerMeter([]types.Measurand{types.MeasurandVoltage, types.MeasurandCurrentImport, types.MeasurandEnergyActiveImportInterval})

	time.Sleep(time.Second)

	s.connector.SamplePowerMeter([]types.Measurand{types.MeasurandVoltage, types.MeasurandCurrentImport, types.MeasurandEnergyActiveImportInterval})
}

func TestConnector(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	suite.Run(t, NewConnectorTestSuite())
}
