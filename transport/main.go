package transport

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-playground/validator"
	"github.com/nats-io/nats.go"
)

type Transport struct {
	nc      *nats.Conn
	errPool *sync.Pool
	okPool  *sync.Pool
}

var vdPool = &sync.Pool{
	New: func() interface{} {
		return validator.New()
	},
}

func New(natsUrl, nkeyFile, name string) (*Transport, error) {
	var errPool = &sync.Pool{
		New: func() interface{} {
			return &NATSError{}
		},
	}
	var okPool = &sync.Pool{
		New: func() interface{} {
			return &NATSOk{}
		},
	}

	optNKey, err := nats.NkeyOptionFromSeed(nkeyFile)
	if err != nil {
		return nil, fmt.Errorf("can't read nkey file")
	}

	nc, err := nats.Connect(natsUrl, optNKey, nats.Name(name))
	if err != nil {
		return nil, err
	}

	return &Transport{
		nc:      nc,
		errPool: errPool,
		okPool:  okPool,
	}, nil
}

func (t *Transport) Handle(e string, fn func(*nats.Msg) (any, int, error)) {
	log.Printf("Subsribing to %s...\n", e)
	t.nc.Subscribe(e, func(msg *nats.Msg) {
		defer func(t time.Time) {
			log.Printf("%s spend %v", e, time.Since(t))
		}(time.Now())

		v, c, err := fn(msg)
		if err != nil {
			errMsg := t.getErrorFromPool(c, err.Error())
			defer t.returnErrorToPool(errMsg)

			resp, err := json.Marshal(errMsg)
			if err != nil {
				log.Println(err)
				return
			}
			if err := msg.Respond(resp); err != nil {
				log.Println(err)
			}
			return
		}

		o := t.getOkFromPool(c, v)
		defer t.returnOkToPool(o)

		resp, err := json.Marshal(o)
		if err != nil {
			log.Println(err)
			return
		}
		if err := msg.Respond(resp); err != nil {
			log.Println(err)
		}
	})
	log.Printf("Subsribed to %s\n", e)
}

func MapperHandler[R, V any](dbFn func(*R) (V, error)) func(*nats.Msg) (any, int, error) {
	validate := getValidatorFromPool()

	return func(m *nats.Msg) (any, int, error) {
		var r R
		if err := json.Unmarshal(m.Data, &r); err != nil {
			return nil, 400, err
		}
		if err := validate.Struct(r); err != nil {
			return nil, 400, err
		}
		v, err := dbFn(&r)
		if err != nil {
			return nil, 500, err
		}
		return v, 200, nil
	}
}
