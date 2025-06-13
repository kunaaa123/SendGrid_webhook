package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sendgridtest/config"
	"sendgridtest/internal/adapters/lark"
	"sendgridtest/internal/adapters/mysql"
	"sendgridtest/internal/core"
	"sendgridtest/internal/domain"
	"sendgridtest/pkg/logger"
	"sendgridtest/pkg/verify"
)

func main() {

	cfg := config.NewConfig()

	logger, err := logger.NewLogger(cfg.LogFile)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logger.Info("Config loaded",
		"public_key_configured", cfg.SendgridPublicKey != "")

	repo, err := mysql.NewRepository(cfg.DatabaseDSN)
	if err != nil {
		logger.Error("Failed to initialize repository", "error", err)
		log.Fatal(err)
	}

	notifier := lark.NewNotifier(cfg.LarkWebhookURL)

	service := core.NewEventService(repo, notifier, logger)

	http.HandleFunc("/webhook", makeWebhookHandler(service, logger, cfg))

	logger.Info("Server starting", "port", cfg.ServerPort)
	if err := http.ListenAndServe(cfg.ServerPort, nil); err != nil {
		logger.Error("Server failed to start", "error", err)
		log.Fatal(err)
	}
}

func readBody(r *http.Request) ([]byte, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	return bodyBytes, err
}

func makeWebhookHandler(service *core.EventService, logger *logger.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		signature := r.Header.Get("X-Twilio-Email-Event-Webhook-Signature")
		timestamp := r.Header.Get("X-Twilio-Email-Event-Webhook-Timestamp")

		if cfg.SendgridPublicKey != "" {
			if signature == "" || timestamp == "" {
				logger.Error("Missing signature headers")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		bodyBytes, err := readBody(r)
		if err != nil {
			http.Error(w, "Cannot read body", http.StatusInternalServerError)
			return
		}

		logger.Info("Signature Verification Headers",
			"signature", signature,
			"timestamp", timestamp)

		if cfg.SendgridPublicKey != "" {
			logger.Info("Starting signature verification",
				"public_key_configured", true)

			valid, err := verify.VerifySignature(bodyBytes, signature, timestamp, cfg.SendgridPublicKey)
			if err != nil {
				logger.Error("Signature verification failed",
					"error", err,
					"signature", signature,
					"timestamp", timestamp)
				http.Error(w, "Invalid signature", http.StatusUnauthorized)
				return
			}
			if !valid {
				logger.Error("Invalid signature detected",
					"signature", signature,
					"timestamp", timestamp)
				http.Error(w, "Invalid signature", http.StatusUnauthorized)
				return
			}

			logger.Info("Signature verification successful",
				"signature_valid", true)
		} else {
			logger.Warn("Signature verification skipped - no public key configured")
		}

		var events []domain.SendgridEvent
		if err := json.Unmarshal(bodyBytes, &events); err != nil {
			logger.Error("Invalid JSON", "error", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		for _, event := range events {
			if err := service.HandleEvent(event); err != nil {
				logger.Error("Event handling failed",
					"event", event.Event,
					"email", event.Email,
					"error", err)
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
