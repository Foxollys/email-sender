package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

// Config хранит настройки SMTP
type Config struct {
	SMTPServer string
	SMTPPort   string
	Username   string
	Password   string
	FromEmail  string
}

// Загружает конфиг из .env
func loadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return Config{
		SMTPServer: os.Getenv("SMTP_SERVER"),
		SMTPPort:   os.Getenv("SMTP_PORT"),
		Username:   os.Getenv("SMTP_USERNAME"),
		Password:   os.Getenv("SMTP_PASSWORD"),
		FromEmail:  os.Getenv("FROM_EMAIL"),
	}
}

// Отправляет письмо через SMTP
func SendEmail(config Config, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", config.SMTPServer, config.SMTPPort)
	auth := smtp.PlainAuth("", config.Username, config.Password, config.SMTPServer)
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)

	err := smtp.SendMail(addr, auth, config.FromEmail, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("SendMail error: %v", err)
	}
	log.Printf("Email sent to %s", to)
	return nil
}

// EmailRequest - структура для JSON
type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// Обработчик HTTP-запросов
func sendEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EmailRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	config := loadConfig()
	err = SendEmail(config, req.To, req.Subject, req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "ok"}`)
}

func main() {

	// Проверка загрузки .env
	//config := loadConfig()
	//fmt.Printf("Loaded config: %+v\n", config) // Выведет данные из .env

	config := loadConfig()
	log.Printf("Server started with SMTP: %s", config.SMTPServer)

	http.HandleFunc("/send", sendEmailHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
