package kafka

type Message struct {
	Key   []byte
	Value []byte
}

type MockProducer struct {
	Messages []*Message
}

func NewMockProducer() *MockProducer {
	return &MockProducer{}
}

func (p *MockProducer) SendMessage(key, value []byte) error {
	p.Messages = append(p.Messages, &Message{Key: key, Value: value})
	return nil
}

func (p *MockProducer) Close() error {
	return nil
}
