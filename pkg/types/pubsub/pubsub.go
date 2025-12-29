package pubsub

type Publisher interface {
	Publish(data []byte) error
}

type Subscriber interface {
	Subscribe() error
}

type PubSub interface {
	Publisher
	Subscriber
}
