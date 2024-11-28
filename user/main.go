package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog"
	"github.com/google/uuid"
)

const (
	auditService = "127.0.0.1:9092"
	auditTopic   = "audit"
)

func main() {
	logger := httplog.NewLogger("user", httplog.Options{
		JSON: true,
	})
	ctx := context.Background()

	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_0_0_0

	sender, err := kafka_sarama.NewSender([]string{auditService}, saramaConfig, auditTopic)
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	defer sender.Close(context.Background())

	ceClient, err := cloudevents.NewClient(sender, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(logger))
	r.Post("/v1/user", storeUser(ctx, ceClient))

	http.Handle("/", r)
	srv := &http.Server{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Addr:         ":3000",
		Handler:      http.DefaultServeMux,
	}
	err = srv.ListenAndServe()
	if err != nil {
		logger.Panic().Msg(err.Error())
	}
}

type userRequest struct {
	ID       uuid.UUID
	Name     string `json:"name"`
	Password string `json:"password"`
}

func storeUser(ctx context.Context, ceClient cloudevents.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		oplog := httplog.LogEntry(r.Context())

		var ur userRequest
		err := json.NewDecoder(r.Body).Decode(&ur)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			oplog.Error().Msg(err.Error())
			return
		}
		ur.ID = uuid.New()
		//TODO: store user in a database

		// Create an Event.
		event := cloudevents.NewEvent()
		event.SetID(uuid.New().String())
		event.SetSource("github.com/eminetto/post-cloudevents")
		event.SetType("user.storeUser")
		event.SetData(cloudevents.ApplicationJSON, map[string]string{"id": ur.ID.String()})

		// Send that Event.
		if result := ceClient.Send(
			// Set the producer message key
			kafka_sarama.WithMessageKey(context.Background(), sarama.StringEncoder(event.ID())),
			event,
		); cloudevents.IsUndelivered(result) {
			oplog.Error().Msgf("failed to send, %v", result)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}
}
