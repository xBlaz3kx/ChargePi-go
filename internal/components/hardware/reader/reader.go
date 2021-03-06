//go:build rpi || dev
// +build rpi dev

package reader

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/xBlaz3kx/ChargePi-go/internal/models/settings"
)

// Supported readers
const (
	PN532 = "PN532"
)

var (
	ErrReaderUnsupported = errors.New("reader type unsupported")
	ErrReaderDisabled    = errors.New("reader disabled")
)

// Reader is an abstraction for an RFID/NFC tag reader.
type Reader interface {
	ListenForTags(ctx context.Context)
	Cleanup()
	Reset()
	GetTagChannel() <-chan string
}

// NewTagReader creates an instance of the Reader interface based on the provided configuration.
func NewTagReader(reader settings.TagReader) (Reader, error) {
	if reader.IsEnabled {
		log.Infof("Preparing tag reader from config: %s", reader.ReaderModel)
		tagChannel := make(chan string, 5)

		switch reader.ReaderModel {
		case PN532:
			return &TagReader{
				TagChannel:       tagChannel,
				DeviceConnection: reader.Device,
				ResetPin:         reader.ResetPin,
			}, nil
		default:
			return nil, ErrReaderUnsupported
		}
	}

	return nil, ErrReaderDisabled
}
