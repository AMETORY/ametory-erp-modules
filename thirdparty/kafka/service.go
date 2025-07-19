package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaService struct {
	ctx       context.Context
	topic     *string
	partition int
	server    *string
}

func NewKafkaService(ctx context.Context, server *string) *KafkaService {
	return &KafkaService{
		ctx:    ctx,
		server: server,
	}
}

// SetTopic sets the topic name for the Kafka topic.
//
// This is a required configuration.
func (s *KafkaService) SetTopic(topic string) {
	s.topic = &topic
}

// SetPartition sets the partition number for the Kafka topic.
//
// This is a required configuration.
func (s *KafkaService) SetPartition(partition int) {
	s.partition = partition
}

func (s *KafkaService) connect(ctx context.Context, server *string) (*kafka.Conn, error) {

	conn, err := kafka.DialLeader(ctx,
		"tcp",
		*server,
		*s.topic,
		s.partition,
	)
	if err != nil {
		fmt.Println("failed to dial leader")
	}
	return conn, err
}

// WriteMessage writes a message to the given topic.
//
// It will return an error if the topic and server are not set.
// It will also return an error if there is a problem connecting to kafka.
// It will also return an error if there is a problem writing the message.
func (s *KafkaService) WriteMessage(topic string, key string, msg []byte) error {
	if s.topic == nil {
		return fmt.Errorf("topic is not set")
	}
	if s.server == nil {
		return fmt.Errorf("server is not set")
	}
	conn, err := s.connect(s.ctx, s.server)
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = conn.WriteMessages(
		kafka.Message{
			Key:   []byte(key),
			Value: msg,
		})
	if err != nil {
		fmt.Println("failed to write messages:", err)
	}
	return err
}

// Read from the topic using kafka.Reader
//
// Example:
//
//	readDeadline, _ := context.WithDeadline(context.Background(),
//		time.Now().Add(5*time.Second))
//
//	for {
//		m, err := r.ReadMessage(readDeadline)
//		if err != nil {
//			break
//		}
//		fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
//	}
//
//	if err := r.Close(); err != nil {
//		log.Fatal("failed to close reader:", err)
//	}
func (s *KafkaService) ReadWithReader(topic string, server string, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{server},
		GroupID:  groupID,
		Topic:    topic,
		MaxBytes: 100, //per message
	})
}
