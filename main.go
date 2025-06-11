package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// SendgridEvent represents the webhook event structure
type SendgridEvent struct {
	Email     string `json:"email"`
	Timestamp int64  `json:"timestamp"`
	Event     string `json:"event"`
	Category  string `json:"category"`
	// เพิ่ม fields อื่นๆตามที่ต้องการ
}

func main() {
	// สร้าง log file
	f, err := os.OpenFile("sendgrid_events.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// webhook handler
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var events []SendgridEvent
		if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// บันทึก events ลง log
		for _, event := range events {
			log.Printf("Event: %s, Email: %s, Timestamp: %d\n",
				event.Event, event.Email, event.Timestamp)
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
