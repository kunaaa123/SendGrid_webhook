package ports

import "sendgridtest/domain"

type EventNotifier interface {
	Notify(event domain.SendgridEvent) error
}
