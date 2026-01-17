package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"webtracker-bot/internal/models"
)

var (
	rxReceiverName    = regexp.MustCompile(`(?i)(?:(?:Receivers?|Receive) Name|Name):\s*([^\n\r]+)`)
	rxReceiverPhone   = regexp.MustCompile(`(?i)(?:(?:Receivers?|Receive) Phone|Phone|Mobile|Tel):\s*([^\n\r]+)`)
	rxReceiverAddress = regexp.MustCompile(`(?i)(?:(?:Receivers?|Receive) Address|Address|Addr):\s*([^\n\r]+)`)
	rxReceiverCountry = regexp.MustCompile(`(?i)(?:(?:Receivers?|Receive) Country|Destination|To|Country):\s*([^\n\r]+)`)
	rxReceiverEmail   = regexp.MustCompile(`(?i)(?:(?:Receivers?|Receive) Email|Email|Mail):\s*([^\n\r]+)`)
	rxReceiverID      = regexp.MustCompile(`(?i)(?:(?:Receivers?|Receive) ID|ID|Passport|NIN):\s*([^\n\r]+)`)
	rxSenderName      = regexp.MustCompile(`(?i)(?:(?:Senders?|Send) Name|Sender):\s*([^\n\r]+)`)
	rxSenderCountry   = regexp.MustCompile(`(?i)(?:(?:Senders?|Send) Country|Origin|From):\s*([^\n\r]+)`)
)

func extract(re *regexp.Regexp, text string) string {
	m := re.FindStringSubmatch(text)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func ParseRegex(text string) models.Manifest {
	m := models.Manifest{
		ReceiverName:    extract(rxReceiverName, text),
		ReceiverAddress: extract(rxReceiverAddress, text),
		ReceiverPhone:   extract(rxReceiverPhone, text),
		ReceiverCountry: extract(rxReceiverCountry, text),
		ReceiverEmail:   extract(rxReceiverEmail, text),
		ReceiverID:      extract(rxReceiverID, text),
		SenderName:      extract(rxSenderName, text),
		SenderCountry:   extract(rxSenderCountry, text),
	}

	m.Validate()
	return m
}

func ParseAI(text, apiKey string) (models.Manifest, error) {
	prompt := fmt.Sprintf(`Extract shipment details from this text and return ONLY JSON.
REQUIRED: receiverName, receiverPhone, receiverCountry, senderName, senderCountry.
OPTIONAL: receiverAddress, receiverEmail, receiverID.

Text: "%s"

JSON Schema:
{
  "receiverName": "string",
  "receiverAddress": "string",
  "receiverPhone": "string",
  "receiverCountry": "string",
  "receiverEmail": "string",
  "receiverID": "string",
  "senderName": "string",
  "senderCountry": "string"
}`, text)

	payload := map[string]interface{}{
		"contents": []interface{}{
			map[string]interface{}{
				"parts": []interface{}{
					map[string]interface{}{"text": prompt},
				},
			},
		},
	}

	jsonBytes, _ := json.Marshal(payload)
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + apiKey

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return models.Manifest{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var aiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &aiResp); err != nil {
		return models.Manifest{}, err
	}

	if len(aiResp.Candidates) == 0 || len(aiResp.Candidates[0].Content.Parts) == 0 {
		return models.Manifest{}, fmt.Errorf("AI returned no results")
	}

	rawJSON := aiResp.Candidates[0].Content.Parts[0].Text
	rawJSON = strings.TrimPrefix(rawJSON, "```json")
	rawJSON = strings.TrimSuffix(rawJSON, "```")
	rawJSON = strings.TrimSpace(rawJSON)

	var m models.Manifest
	if err := json.Unmarshal([]byte(rawJSON), &m); err != nil {
		return models.Manifest{}, err
	}

	return m, nil
}
