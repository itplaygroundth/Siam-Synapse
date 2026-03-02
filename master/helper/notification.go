package helper

import (
	"bytes"
	"encoding/json"
	//"fmt"
	"log"
	"net/http"
	"os"
	"github.com/joho/godotenv"
)

// PushMessage โครงสร้างข้อมูลสำหรับ LINE Push Message
type PushMessage struct {
	To       string            `json:"to"`
	Messages []LineTextMessage `json:"messages"`
}

type LineTextMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// SendLineAdminPush ทำหน้าที่ส่งแจ้งเตือนหา Admin โดยตรง
func SendLineAdminPush(message string) {

	err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
	
	token := os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")
	adminID := os.Getenv("LINE_ADMIN_USER_ID")
	
	if token == "" || adminID == "" {
		log.Println("⚠️ LINE_CHANNEL_ACCESS_TOKEN or LINE_ADMIN_USER_ID not set")
		return
	}

	url := "https://api.line.me/v2/bot/message/push"
	
	// เตรียมข้อมูล JSON
	payload := PushMessage{
		To: adminID,
		Messages: []LineTextMessage{
			{
				Type: "text",
				Text: message,
			},
		},
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
        log.Println("Error sending LINE notify:", err)
        return
    }
    defer resp.Body.Close()

    log.Println("LINE Notify sent, status:", resp.Status)
}