package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
)

const (
	kafkaBroker  = "localhost:9092"
	kafkaTopic   = "messages"
	pollInterval = 10 * time.Second
)

func fetchMessage() ([]byte, error) {
	msg := fmt.Sprintf("%v: %s", time.Now(), "msg")
	return []byte(msg), nil
}

func writeToKafka(ctx context.Context, data []byte) error {
	writer := &kafka.Writer{
		Addr:  kafka.TCP(kafkaBroker),
		Topic: kafkaTopic,
	}
	defer writer.Close()

	msg := kafka.Message{
		Value: data,
	}

	return writer.WriteMessages(ctx, msg)
}

func main() {
	ctx := context.Background()

	mainCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, gCtx := errgroup.WithContext(mainCtx)
	g.Go(func() error {
		<-gCtx.Done()
		fmt.Println("Shutting down gracefully")
		return nil
	})
	g.Go(func() error {
		timer := time.NewTimer(pollInterval)
		for {
			select {
			case <-gCtx.Done():
				fmt.Println("Exiting publisher")
				return nil
			case <-timer.C:
				func() {
					defer timer.Reset(pollInterval)
					messages, err := fetchMessage()
					if err != nil {
						fmt.Printf("Error fetching messages: %s\n", err)
						return
					}

					if err := writeToKafka(ctx, messages); err != nil {
						fmt.Printf("Error writing to Kafka: %s\n", err)
						return
					}

					fmt.Println("Message written to Kafka successfully")
				}()
			}
		}
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("exit: %s \n", err)
	}

}
