package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.AsyncProducer
	log      *slog.Logger
}

func NewProducer(address []string, log *slog.Logger) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Timeout = 10 * time.Second

	p, err := sarama.NewAsyncProducer(address, config)
	if err != nil {
		return nil, fmt.Errorf("error creating new producer: %w", err)
	}

	return &Producer{
		producer: p,
		log:      log,
	}, nil
}

func (p *Producer) Produce(ctx context.Context, message, topic string) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}
	//Sending message
	select {
	case p.producer.Input() <- msg:
	case <-ctx.Done():
		return ctx.Err()
	}
	// Feedback
	select {
	case err := <-p.producer.Errors():
		return fmt.Errorf("delivery failed: %w", err)
	case <-p.producer.Successes():
		p.log.Info("success sending message")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(15 * time.Second):
		return errors.New("delivery timeout exceeded")
	}
}

func (p *Producer) Close() error {
	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("error closing producer: %w", err)
	}
	p.log.Info("producer closed successfully")
	return nil
}
