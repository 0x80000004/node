package communication

import dto_discovery "github.com/mysterium/node/service_discovery/dto"

type Server interface {
	Start() error
	Stop() error
	ServeDialogs(callback DialogHandler) error
}

type DialogHandler func(Sender, Receiver)

type Client interface {
	Start() error
	Stop() error
	CreateDialog(contact dto_discovery.ContactDefinition) (Sender, Receiver, error)
}

type Receiver interface {
	Receive(handler MessageHandler) error
	Respond(handler RequestHandler) error
}

type Sender interface {
	Send(producer MessageProducer) error
	Request(producer RequestProducer) (responsePtr interface{}, err error)
}
