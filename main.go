package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type SendgridEvent struct {
	Email     string `json:"email"`
	Timestamp int64  `json:"timestamp"`
	Event     string `json:"event"`
	Category  string `json:"category"`
	// เพิ่ม fields อื่นๆตามที่ต้องการ
}

// ฟังก์ชันส่งข้อความไปยัง Lark Webhook
func sendLarkNotification(event SendgridEvent) {
	webhookURL := "https://open.larksuite.com/open-apis/bot/v2/hook/88fccfea-8fad-47d9-99a9-44d214785fff" // <-- เปลี่ยนเป็น URL ของคุณ

	message := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": "📧 SendGrid Event แจ้งเตือน:\n" +
				"Event: " + event.Event + "\n" +
				"Email: " + event.Email + "\n" +
				"Category: " + event.Category,
			//Spam Reports
			//Dropped
			//Bounced
		},
	}

	body, _ := json.Marshal(message)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Lark notify failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Lark response status: %s", resp.Status)
	}
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

		for _, event := range events {
			// log ลงไฟล์
			log.Printf("Event: %s, Email: %s, Timestamp: %d\n",
				event.Event, event.Email, event.Timestamp)

			// ส่งแจ้งเตือน Lark
			sendLarkNotification(event)
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Println("🚀 Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
