package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
)

const (
	kafkaBrokerAddress = "localhost:9092"
	topic              = "messages"
)

func serverPage(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "cmd/client/index.html")
}

func main() {

	ctx := context.Background()

	mainCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, gCtx := errgroup.WithContext(mainCtx)
	g.Go(func() error {
		<-gCtx.Done()
		fmt.Println("Shutting down gracefully ...")
		return nil
	})

	hub := newHub()
	g.Go(func() error {
		hub.run(gCtx)
		return nil
	})
	g.Go(func() error {
		log.Println("Starting the Kafka consumer...")
		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{kafkaBrokerAddress},
			Topic:     topic,
			Partition: 0,
			MinBytes:  10e3,
			MaxBytes:  10e6,
		})
		for {
			m, err := r.ReadMessage(gCtx)
			if err != nil {
				log.Fatalf("Error reading message: %v", err)
			}
			fmt.Printf("Received message: %s from topic: %s\n", m.Value, m.Topic)
			hub.publish(string(m.Value))
		}
	})

	http.HandleFunc("/", serverPage)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 3 * time.Second,
	}
	g.Go(func() error {
		fmt.Println("Starting Http server")
		return server.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		fmt.Println("Shutting down ")
		return server.Shutdown(ctx)
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("exit: %s \n", err)
	}

}
