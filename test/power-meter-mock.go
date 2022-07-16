package test

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type PowerMeterMock struct {
	mock.Mock
}

/*---------------------- Power Meter Mock ----------------------*/

func (p *PowerMeterMock) Init(ctx context.Context) error {
	return p.Called().Error(0)
}

func (p *PowerMeterMock) Reset() {
	p.Called()
}

func (p *PowerMeterMock) GetEnergy() float64 {
	args := p.Called()
	return args.Get(0).(float64)
}

func (p *PowerMeterMock) GetPower() float64 {
	args := p.Called()
	return args.Get(0).(float64)
}

func (p *PowerMeterMock) GetCurrent() float64 {
	args := p.Called()
	return args.Get(0).(float64)
}

func (p *PowerMeterMock) GetVoltage() float64 {
	args := p.Called()
	return args.Get(0).(float64)
}

func (p *PowerMeterMock) GetRMSCurrent() float64 {
	args := p.Called()
	return args.Get(0).(float64)
}

func (p *PowerMeterMock) GetRMSVoltage() float64 {
	args := p.Called()
	return args.Get(0).(float64)
}
