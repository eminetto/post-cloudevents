package main

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM/sarama"

	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

const (
	auditService = "127.0.0.1:9092"
	auditTopic   = "audit"
	auditGroupID = "audit-group-id"
)

func receive(event cloudevents.Event) {
	// do something with event.
	fmt.Printf("%s", event)
}

func main() {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_0_0_0

	receiver, err := kafka_sarama.NewConsumer([]string{auditService}, saramaConfig, auditGroupID, auditTopic)
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	defer receiver.Close(context.Background())

	c, err := cloudevents.NewClient(receiver)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	if err = c.StartReceiver(context.Background(), receive); err != nil {
		log.Fatalf("failed to start receiver: %v", err)
	}
}
