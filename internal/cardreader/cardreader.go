package cardreader

import (
	"context"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/ebfe/scard"
)

type CardReader struct {
	device  string
	polling bool
	ctx     *scard.Context
}

type Scan struct {
	UID      string
	ScanTime time.Time
	Error    error
}

func NewCardReader() (*CardReader, error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return nil, err
	}

	rls, err := ctx.ListReaders()
	if err != nil {
		return nil, err
	}

	device := rls[0]

	return &CardReader{
		device: device,
		ctx:    ctx,
	}, nil
}

func (r *CardReader) ReadCard(ct context.Context) (Scan, error) {
	r.polling = true

	var card Scan

	for r.polling {
		if ct.Err() == context.DeadlineExceeded {
			return card, errors.New("Deadline Exceeded")
		}
		ctx := r.ctx
		if ctx == nil {
			continue
		}

		rs := []scard.ReaderState{{
			Reader:       r.device,
			CurrentState: scard.StateUnaware,
		}}

		err := ctx.GetStatusChange(rs, 250*time.Millisecond)
		if err != nil {
			log.Printf("error getting status change: %s", err)
			continue
		}

		if rs[0].EventState&scard.StatePresent == 0 {
			continue
		}

		tag, err := ctx.Connect(r.device, scard.ShareShared, scard.ProtocolAny)
		if err != nil {
			log.Printf("error connecting to reader: %s", err)
			continue
		}

		res, err := tag.Transmit([]byte{0xFF, 0xCA, 0x00, 0x00, 0x00})
		if err != nil {
			log.Printf("error transmitting: %s", err)
			continue
		}

		if len(res) < 2 {
			log.Printf("invalid response")
			_ = tag.Disconnect(scard.ResetCard)
			continue
		}

		uid := res[:len(res)-2]

		card.UID = hex.EncodeToString(uid)
		card.ScanTime = time.Now()

		_ = tag.Disconnect(scard.ResetCard)

		r.Close()
	}

	return card, nil
}

func (r *CardReader) Close() error {
	r.polling = false
	if r.ctx != nil {
		err := r.ctx.Release()
		if err != nil {
			return err
		}
	}
	return nil
}
