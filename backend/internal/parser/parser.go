package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"webtracker-bot/internal/models"
)

func isReceiver(lower string) bool {
	return strings.Contains(lower, "receiver") || strings.Contains(lower, "reciver") || strings.Contains(lower, "receive") || strings.Contains(lower, "recieve")
}

func isSender(lower string) bool {
	return strings.Contains(lower, "sender")
}

func ParseRegex(text string) models.Manifest {
	m := models.Manifest{}
	lines := strings.Split(text, "\n")

	type field struct {
		target   *string
		keywords []string
		isPrefix bool // Fallback for standard "Name:" format
	}

	receiverFields := []field{
		{&m.ReceiverName, []string{"name"}, true},
		{&m.ReceiverPhone, []string{"phone", "mobile", "tel"}, false},
		{&m.ReceiverAddress, []string{"address", "addr"}, false},
		{&m.ReceiverCountry, []string{"country", "destination", "to"}, false},
		{&m.ReceiverEmail, []string{"email", "mail"}, false},
		{&m.ReceiverID, []string{"id", "passport", "nin"}, false},
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)

		// 1. Check for Sender Fields (Unique prefix)
		if strings.Contains(lower, "sender") {
			if strings.Contains(lower, "name") {
				m.SenderName = cleanLine(line, "name")
			} else if strings.Contains(lower, "country") || strings.Contains(lower, "origin") || strings.Contains(lower, "from") {
				m.SenderCountry = cleanLine(line, "country", "origin", "from")
			}
			continue
		}

		// 2. Check for Receiver Fields with Person Keywords (Robust)
		if isReceiver(lower) {
			found := false
			for _, f := range receiverFields {
				for _, kw := range f.keywords {
					if strings.Contains(lower, kw) {
						*f.target = cleanLine(line, kw)
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			// Special case: "Receiver: John Doe" (No specific field name)
			if !found && m.ReceiverName == "" {
				m.ReceiverName = cleanLine(line, "receiver", "reciver", "receive", "recieve")
			}
			continue
		}

		// 3. Fallback to Standard "Field:" format (No receiver/sender prefix)
		for _, f := range receiverFields {
			if *f.target != "" {
				continue
			}
			for _, kw := range f.keywords {
				prefix := kw + ":"
				if strings.HasPrefix(lower, prefix) {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) > 1 {
						*f.target = strings.TrimSpace(parts[1])
					}
					break
				}
			}
		}

		// Special case fallbacks for Sender
		if m.SenderCountry == "" && (strings.HasPrefix(lower, "origin:") || strings.HasPrefix(lower, "from:")) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 1 {
				m.SenderCountry = strings.TrimSpace(parts[1])
			}
		}
	}

	m.Validate()
	return m
}

func cleanLine(line string, keywords ...string) string {
	lower := strings.ToLower(line)
	bestIdx := -1
	for _, kw := range keywords {
		if idx := strings.Index(lower, kw); idx != -1 {
			if bestIdx == -1 || idx > bestIdx {
				bestIdx = idx + len(kw)
			}
		}
	}
	if bestIdx == -1 {
		return ""
	}
	res := line[bestIdx:]
	res = strings.TrimLeft(res, " '’s:：") // Clean up apostrophes, colons, and spaces
	return strings.TrimSpace(res)
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
