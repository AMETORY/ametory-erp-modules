package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaService struct {
	ctx         context.Context
	connections map[string]*kafka.Conn
}

func NewKafkaService(ctx context.Context) *KafkaService {

	return &KafkaService{
		ctx:         ctx,
		connections: make(map[string]*kafka.Conn), // = value
	}
}

func (s *KafkaService) AddNewConnection(server string, topic string, partition int) error {
	conn, err := connect(s.ctx, server, topic, partition)
	if err != nil {
		return err
	}
	s.connections[server] = conn
	return nil
}

func connect(ctx context.Context, server string, topic string, partition int) (*kafka.Conn, error) {
	conn, err := kafka.DialLeader(ctx, "tcp",
		server, topic, partition)
	if err != nil {
		fmt.Println("failed to dial leader")
	}
	return conn, err
}

func (s *KafkaService) WriteMessages(topic string, msgs []string) error {
	conn, ok := s.connections[topic]
	if !ok {
		return fmt.Errorf("connection not found")
	}
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	var err error
	for _, msg := range msgs {
		_, err = conn.WriteMessages(
			kafka.Message{Value: []byte(msg)})

	}
	if err != nil {
		fmt.Println("failed to write messages:", err)
		return err
	}

	return nil
} //end writeMessages

// Reads all messages in the partition from the start
// Specify a minimum and maximum size in bytes to read (1 char = 1 byte)
func (s *KafkaService) ReadMessages(topic string, minSize int, maxSize int, callback func(output []byte)) error {
	conn, ok := s.connections[topic]
	if !ok {
		return fmt.Errorf("connection not found")
	}
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	batch := conn.ReadBatch(minSize, maxSize) //in bytes

	msg := make([]byte, 10e3) //set the max length of each message
	for {
		msgSize, err := batch.Read(msg)
		if err != nil {
			break
		}
		// fmt.Println(string(msg[:msgSize]))
		callback(msg[:msgSize])
	}

	if err := batch.Close(); err != nil { //make sure to close the batch
		fmt.Println("failed to close batch:", err)
	}
	return nil
} //end readMessages

// Read from the topic using kafka.Reader
// Readers can use consumer groups (but are not required to)
func (s *KafkaService) ReadWithReader(topic string, server string, groupID string, callback func(output kafka.Message)) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{server},
		GroupID:  groupID,
		Topic:    topic,
		MaxBytes: 100, //per message
		// more options are available
	})

	//Create a deadline
	readDeadline, _ := context.WithDeadline(context.Background(),
		time.Now().Add(5*time.Second))
	for {
		msg, err := r.ReadMessage(readDeadline)
		if err != nil {
			break
		}
		fmt.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n",
			msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
		callback(msg)
	}

	if err := r.Close(); err != nil {
		fmt.Println("failed to close reader:", err)
	}
}
