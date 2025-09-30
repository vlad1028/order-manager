package service

import (
	"encoding/json"
	"github.com/vlad1028/order-manager/internal/models/order"
	"log"
)

func (s *Service) sendEvent(event order.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return
	}
	err = s.kafkaProducer.SendMessage([]byte(event.OrderID.String()), data)
	if err != nil {
		log.Printf("Failed to send event to Kafka: %v", err)
	}
}
