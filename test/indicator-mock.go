package test

import "github.com/stretchr/testify/mock"

type IndicatorMock struct {
	mock.Mock
}

/*------------------ Indicator mock ------------------*/

func (i *IndicatorMock) DisplayColor(index int, colorHex uint32) error {
	args := i.Called(index, colorHex)
	return args.Error(0)
}

func (i *IndicatorMock) Blink(index int, times int, colorHex uint32) error {
	args := i.Called(index, times, colorHex)
	return args.Error(0)
}

func (i *IndicatorMock) Cleanup() {
	i.Called()
}
