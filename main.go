package main

import (
	"fmt"
	"sync"
	"time"
)

type Message struct {
	Topic   string
	Payload interface{}
}

type Subscriber struct {
	Channel     chan interface{}
	Unsubscribe chan bool
}

type Broker struct {
	subscribers map[string][]*Subscriber
	mutex       sync.Mutex
}

func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[string][]*Subscriber),
	}
}

func (b *Broker) Subscribe(topic string) *Subscriber {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	subscriber := &Subscriber{
		Channel:     make(chan interface{}, 1),
		Unsubscribe: make(chan bool),
	}

	b.subscribers[topic] = append(b.subscribers[topic], subscriber)
	return subscriber
}

func (b *Broker) Unsubscribe(topic string, subscriber *Subscriber) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if subscribers, found := b.subscribers[topic]; found {
		for i, sub := range subscribers {
			if sub == subscriber {
				close(sub.Channel)
				b.subscribers[topic] = append(subscribers[:i], subscribers[i+1:]...)
				return
			}
		}
	}
}

func (b *Broker) Publish(topic string, payload interface{}) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if subscribers, found := b.subscribers[topic]; found {
		for _, sub := range subscribers {
			select {
			case sub.Channel <- payload:
			case <-time.After(time.Second):
				fmt.Printf("slow susbscriber, unsubscribing from topic %s/n", topic)
				b.Unsubscribe(topic, sub)
			}

		}
	}
}

func main() {

	broker := NewBroker()
	subscriber := broker.Subscribe("sample_topic")

	go func() {
		for {
			select {
			case msg, ok := <-subscriber.Channel:
				if !ok {
					return
				}
				fmt.Printf("received %v\n", msg)
			case <-subscriber.Unsubscribe:
				return
			}
		}
	}()

	broker.Publish("sample_topic", "Hello World!")
	broker.Publish("sample_topic", "This is testing!")

	time.Sleep(2 * time.Second)

	broker.Unsubscribe("sample_topic", subscriber)

}
