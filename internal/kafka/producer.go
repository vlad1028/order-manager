package kafka

import "github.com/IBM/sarama"

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewSyncProducer(brokers []string, topic string) (*Producer, error) {
	c := sarama.NewConfig()
	c.Producer.Partitioner = sarama.NewHashPartitioner
	c.Producer.RequiredAcks = sarama.WaitForAll
	c.Producer.Return.Errors = true
	c.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, c)
	if err != nil {
		return nil, err
	}

	return &Producer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *Producer) SendMessage(key, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	_, _, err := p.producer.SendMessage(msg)
	return err
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
