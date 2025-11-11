package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
)

func (app *Application) isHelpMessage(msg, atMsg string) bool {
	return strings.Contains(msg, atMsg) && !strings.Contains(msg, "vv ")
}

func (app *Application) Post(URL string, payload []byte, token string) (err error) {
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Client.Do(req)
	if err != nil {
		log.Printf("Error sending message to napcat: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to send message to napcat, status: %s, response: %s", resp.Status, string(body))
		return
	}
	return nil
}
