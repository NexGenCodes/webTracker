package payment

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type PaystackService struct {
	secretKey string
}

func NewPaystackService(secretKey string) *PaystackService {
	return &PaystackService{secretKey: secretKey}
}

func (s *PaystackService) InitializeTransaction(email string, amount int, callbackURL string, metadata map[string]interface{}) (string, error) {
	url := "https://api.paystack.co/transaction/initialize"

	payload := map[string]interface{}{
		"email":        email,
		"amount":       amount, // in kobo
		"callback_url": callbackURL,
		"metadata":     metadata,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.secretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("paystack returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var res struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AuthorizationURL string `json:"authorization_url"`
			AccessCode       string `json:"access_code"`
			Reference        string `json:"reference"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	if !res.Status {
		return "", fmt.Errorf("paystack error: %s", res.Message)
	}

	return res.Data.AuthorizationURL, nil
}

func (s *PaystackService) VerifySignature(payload []byte, signature string) bool {
	mac := hmac.New(sha512.New, []byte(s.secretKey))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	return expectedMAC == signature
}
