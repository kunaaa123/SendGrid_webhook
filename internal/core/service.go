package core

import (
	"sendgridtest/internal/domain"
	"sendgridtest/internal/ports"
	"sendgridtest/pkg/logger"
	"time"
)

type EventService struct {
	repository ports.EventRepository
	notifier   ports.Notifier
	logger     *logger.Logger
}

func NewEventService(repo ports.EventRepository, notifier ports.Notifier, logger *logger.Logger) *EventService {
	return &EventService{
		repository: repo,
		notifier:   notifier,
		logger:     logger,
	}
}

func isMainEvent(eventType string) bool {
	// รองรับทุก events จาก SendGrid
	supportedEvents := map[string]bool{
		"processed":         true,
		"deferred":          true,
		"delivered":         true,
		"open":              true,
		"click":             true,
		"bounce":            true,
		"dropped":           true,
		"spamreport":        true,
		"unsubscribe":       true,
		"group_unsubscribe": true,
		"group_resubscribe": true,
		"blocked":           true,
		"invalid_email":     true,
		"test":              true,
	}

	return supportedEvents[eventType]
}

func (s *EventService) HandleEvent(event domain.SendgridEvent) error {
	if !isMainEvent(event.Event) {
		return nil
	}

	s.logger.Info("SendGrid Event",
		"event", event.Event,
		"email", event.Email,
		"timestamp", time.Unix(event.Timestamp, 0).Format("2006-01-02 15:04:05"),
		"sg_event_id", event.SGEventID,
		"status", event.Status)

	if err := s.repository.SaveEvent(event); err != nil {
		s.logger.Error("Failed to save event", "error", err)
		return err
	}

	// แจ้งเตือนสำหรับ events ที่สำคัญ
	criticalEvents := map[string]bool{
		"bounce":        true,
		"spamreport":    true,
		"dropped":       true,
		"blocked":       true,
		"invalid_email": true,
	}

	if criticalEvents[event.Event] {
		if err := s.notifier.Notify(event); err != nil {
			s.logger.Error("Failed to send notification",
				"error", err,
				"event", event.Event,
				"email", event.Email)
			return err
		}
		s.logger.Info("Notification sent",
			"event", event.Event,
			"email", event.Email)
	}

	return nil
}
