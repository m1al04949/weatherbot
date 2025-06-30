package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/IBM/sarama"
)

type Consumer struct {
	consumer sarama.Consumer
	log      *slog.Logger
	wg       sync.WaitGroup
}

func NewConsumer(address []string, log *slog.Logger) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_5_0_0
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	c, err := sarama.NewConsumer(address, config)
	if err != nil {
		return nil, fmt.Errorf("error creating consumer: %w", err)
	}

	return &Consumer{
		consumer: c,
		log:      log,
	}, nil
}

func (c *Consumer) Consume(ctx context.Context, topics []string, handler func(message *sarama.ConsumerMessage)) error {
	partitionConsumers := make([]sarama.PartitionConsumer, 0, len(topics))
	for _, topic := range topics {
		partitions, err := c.consumer.Partitions(topic)
		if err != nil {
			return fmt.Errorf("failed to get partitions for topic %s: %w", topic, err)
		}

		for _, partition := range partitions {
			pc, err := c.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
			if err != nil {
				return fmt.Errorf("failed to consume partition %d of topic %s: %w", partition, topic, err)
			}
			defer pc.AsyncClose()
			partitionConsumers = append(partitionConsumers, pc)
		}
	}
	//Preparing message
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-ctx.Done():
				c.log.Info("stopping consumer due to context cancellation")
				return
			default:
				for _, pc := range partitionConsumers {
					select {
					case err := <-pc.Errors():
						c.log.Error("partition consumer error", slog.String("error", err.Error()))
					case msg := <-pc.Messages():
						c.log.Debug("message received",
							slog.String("topic", msg.Topic),
							slog.Int("partition", int(msg.Partition)),
							slog.Int64("offset", msg.Offset))
						handler(msg)
					default:
						continue
					}
				}
			}
		}
	}()

	return nil
}

func (c *Consumer) Close() error {
	c.wg.Wait()
	if err := c.consumer.Close(); err != nil {
		return fmt.Errorf("error closing consumer: %w", err)
	}
	c.log.Info("consumer closed successfully")
	return nil
}
