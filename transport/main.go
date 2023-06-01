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

var validate = validator.New()

// New initializes a new Transport instance. It connects to a NATS server using provided URL and nkey file,
// and creates error and OK response pools for efficient handling of responses.
// natsUrl: NATS server URL
// nkeyFile: path to the nkey file for authentication
// name: name of the NATS client
// Returns a pointer to a Transport instance or an error.
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

// Handle subscribes to a topic in the message broker, executes a function 'fn' for each received message,
// and sends a response back. It logs the processing time for each message and handles any errors that occur,
// using pools for error and OK responses to optimize resource usage.
// e: topic to subscribe to
// fn: function to process each message received
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

// MapperHandler creates a function that serves as a bridge between your database
// and a message broker (nats.Msg). It takes a function as a parameter that retrieves
// a value of type V from a pointer to a value of type R, and maps this to a nats.Msg.
//
// The returned function takes a pointer to nats.Msg as its argument. It attempts
// to unmarshal the data from the message into a value of type R. If the data
// is not nil and unmarshalling is successful, the function performs validation on the
// struct. If validation fails or if an error occurs during unmarshalling, it returns
// the error with a 400 status code.
//
// After successful validation, the dbFn function is called with a pointer to the
// value of type R, and if an error occurs during this database operation, it returns
// the error with a 500 status code.
//
// If all operations are successful, it returns the value retrieved from the database
// (of type V), a 200 status code, and nil for error.
//
// dbFn: function that takes a pointer to a value of type R and returns a value of type V and an error.
// It is used to perform the database operation.
//
// Returns: A function that takes a pointer to nats.Msg and returns a value of type any,
// a status code of type int, and an error. This returned function serves as the handler
// for processing the message broker data and mapping it to the database model.
func MapperHandler[R, V any](dbFn func(*R) (V, error)) func(*nats.Msg) (any, int, error) {
	return func(m *nats.Msg) (any, int, error) {
		var r R
		if m.Data != nil && len(m.Data) > 0 {
			if err := json.Unmarshal(m.Data, &r); err != nil {
				return nil, 400, err
			}
			if err := validate.Struct(r); err != nil {
				return nil, 400, err
			}
		}
		v, err := dbFn(&r)
		if err != nil {
			return nil, 500, err
		}
		return v, 200, nil
	}
}
