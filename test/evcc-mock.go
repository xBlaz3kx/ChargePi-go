package test

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type EvccMock struct {
	mock.Mock
}

func (e *EvccMock) Init(ctx context.Context) error {
	return e.Called().Error(0)
}

func (e *EvccMock) EnableCharging() error {
	return e.Called().Error(0)
}

func (e *EvccMock) DisableCharging() {
	e.Called()
}

func (e *EvccMock) SetMaxChargingCurrent(value float64) error {
	return e.Called(value).Error(0)
}

func (e *EvccMock) GetMaxChargingCurrent() float64 {
	return e.Called().Get(0).(float64)
}

func (e *EvccMock) Lock() {
	e.Called()
}

func (e *EvccMock) Unlock() {
	e.Called()
}

func (e *EvccMock) GetState() string {
	return e.Called().String(0)
}

func (e *EvccMock) Cleanup() error {
	return e.Called().Error(0)
}
