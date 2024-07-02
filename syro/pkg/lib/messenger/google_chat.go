package messenger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func SendToGoogleChat(ckatKey, msg string) error {
	if ckatKey == "" {
		return fmt.Errorf("google chat key is empty")
	}

	if msg == "" {
		return fmt.Errorf("sendable message to google chat is empty")
	}

	appMessage := map[string]string{"text": msg}

	body, err := json.Marshal(appMessage)
	if err != nil {
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", ckatKey, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	_, err = client.Do(req)
	return err
}
