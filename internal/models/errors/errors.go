package errors

import "errors"

var (
	ErrConnectorNil               = errors.New("connector pointer is nil")
	ErrConnectorNotCharging       = errors.New("connector not charging")
	ErrNoConnectorWithTag         = errors.New("no connector with tag id")
	ErrNoConnectorWithTransaction = errors.New("no connector with transaction id")
	ErrNoAvailableConnectors      = errors.New("no available connectors")
	ErrConnectorUnavailable       = errors.New("connector unavailable")
	ErrChargePointUnavailable     = errors.New("charge point unavailable")
	ErrTagUnauthorized            = errors.New("tag unauthorized")
)
