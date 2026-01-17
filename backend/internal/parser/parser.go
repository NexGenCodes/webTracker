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
	rxReceiverName    = regexp.MustCompile(`(?i)(?:Receiver|Reciver|Receive|Recieve)(?:[''’’]s)?(?:\s*:\s*|\s+)(?:Name)?[:\s]*\s*([^\n\r]+)|(?:^|\n)Name:\s*([^\n\r]+)`)
	rxReceiverPhone   = regexp.MustCompile(`(?i)(?:Receiver|Reciver|Receive|Recieve)(?:[''’’]s)?(?:\s*:\s*|\s+)(?:Phone|Mobile|Tel)[:\s]*\s*([^\n\r]+)|(?:^|\n)(?:Phone|Mobile|Tel|TEL):\s*([^\n\r]+)`)
	rxReceiverAddress = regexp.MustCompile(`(?i)(?:Receiver|Reciver|Receive|Recieve)(?:[''’’]s)?(?:\s*:\s*|\s+)(?:Address|Addr)[:\s]*\s*([^\n\r]+)|(?:^|\n)(?:Address|Addr|ADDR):\s*([^\n\r]+)`)
	rxReceiverCountry = regexp.MustCompile(`(?i)(?:Receiver|Reciver|Receive|Recieve)(?:[''’’]s)?(?:\s*:\s*|\s+)(?:Country|Destination|To)[:\s]*\s*([^\n\r]+)|(?:^|\n)(?:Country|Destination|To|COUNTRY):\s*([^\n\r]+)`)
	rxReceiverEmail   = regexp.MustCompile(`(?i)(?:Receiver|Reciver|Receive|Recieve)(?:[''’’]s)?(?:\s*:\s*|\s+)(?:Email|Mail)[:\s]*\s*([^\n\r]+)|(?:^|\n)(?:Email|Mail):\s*([^\n\r]+)`)
	rxReceiverID      = regexp.MustCompile(`(?i)(?:Receiver|Reciver|Receive|Recieve)(?:[''’’]s)?(?:\s*:\s*|\s+)(?:ID|Passport|NIN)[:\s]*\s*([^\n\r]+)|(?:^|\n)(?:ID|Passport|NIN):\s*([^\n\r]+)`)
	rxSenderName      = regexp.MustCompile(`(?i)Sender(?:[''’’]s)?(?:\s*:\s*|\s+)(?:Name)?[:\s]*\s*([^\n\r]+)|(?:^|\n)Sender Name:\s*([^\n\r]+)`)
	rxSenderCountry   = regexp.MustCompile(`(?i)Sender(?:[''’’]s)?(?:\s*:\s*|\s+)(?:Country|Origin|From)[:\s]*\s*([^\n\r]+)|(?:^|\n)(?:Origin|From|Sender Country):\s*([^\n\r]+)`)
)

func extract(re *regexp.Regexp, text string) string {
	m := re.FindStringSubmatch(text)
	if len(m) > 1 {
		for _, match := range m[1:] {
			if match != "" {
				val := strings.TrimSpace(match)
				// Clean up trailing labels that might have been captured
				labels := []string{"Name:", "Phone:", "Mobile:", "Tel:", "Address:", "Addr:", "Country:", "Destination:", "To:", "Email:", "ID:", "Passport:", "NIN:", "Origin:", "From:", "Sender:", "Receiver:", "Name ", "Phone ", "Mobile ", "Tel ", "Address ", "Addr ", "Country ", "Destination ", "To ", "Email ", "ID ", "Passport ", "NIN ", "Origin ", "From ", "Sender ", "Receiver "}
				for _, label := range labels {
					if idx := strings.Index(strings.ToLower(val), strings.ToLower(label)); idx != -1 {
						val = strings.TrimSpace(val[:idx])
					}
				}
				return val
			}
		}
	}
	return ""
}

func isReceiver(lower string) bool {
	return strings.Contains(lower, "receiver") || strings.Contains(lower, "reciver") || strings.Contains(lower, "receive") || strings.Contains(lower, "recieve")
}

func isSender(lower string) bool {
	return strings.Contains(lower, "sender")
}

func ParseRegex(text string) models.Manifest {
	m := models.Manifest{}
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		lower := strings.ToLower(line)

		// Name
		if isReceiver(lower) && strings.Contains(lower, "name") {
			m.ReceiverName = cleanLine(line, "name")
		} else if isSender(lower) && strings.Contains(lower, "name") {
			m.SenderName = cleanLine(line, "name")
		} else if strings.HasPrefix(lower, "name:") {
			if m.ReceiverName == "" {
				m.ReceiverName = strings.TrimSpace(line[5:])
			}
		} else if isReceiver(lower) && !strings.Contains(lower, "country") && !strings.Contains(lower, "phone") && !strings.Contains(lower, "address") && !strings.Contains(lower, "email") && !strings.Contains(lower, "id") {
			// Case like "Receiver: John Doe" or "Reciver's: John Doe"
			if m.ReceiverName == "" {
				m.ReceiverName = cleanLine(line, "receiver", "reciver", "receive", "recieve")
			}
		}

		// Phone
		if isReceiver(lower) && (strings.Contains(lower, "phone") || strings.Contains(lower, "mobile") || strings.Contains(lower, "tel")) {
			m.ReceiverPhone = cleanLine(line, "phone", "mobile", "tel")
		} else if strings.Contains(lower, "phone:") || strings.Contains(lower, "tel:") || strings.Contains(lower, "mobile:") {
			if m.ReceiverPhone == "" {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) > 1 {
					m.ReceiverPhone = strings.TrimSpace(parts[1])
				}
			}
		}

		// Address
		if isReceiver(lower) && (strings.Contains(lower, "address") || strings.Contains(lower, "addr")) {
			m.ReceiverAddress = cleanLine(line, "address", "addr")
		} else if strings.Contains(lower, "address:") || strings.HasPrefix(lower, "addr:") {
			if m.ReceiverAddress == "" {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) > 1 {
					m.ReceiverAddress = strings.TrimSpace(parts[1])
				}
			}
		}

		// Country
		if isReceiver(lower) && (strings.Contains(lower, "country") || strings.Contains(lower, "destination") || strings.Contains(lower, "to")) {
			m.ReceiverCountry = cleanLine(line, "country", "destination", "to")
		} else if isSender(lower) && (strings.Contains(lower, "country") || strings.Contains(lower, "origin") || strings.Contains(lower, "from")) {
			m.SenderCountry = cleanLine(line, "country", "origin", "from")
		} else if strings.Contains(lower, "country:") || strings.Contains(lower, "destination:") || strings.Contains(lower, "origin:") {
			if strings.Contains(lower, "origin") || (isSender(lower) && strings.Contains(lower, "country")) {
				m.SenderCountry = cleanLine(line, "country", "origin", "from")
			} else if m.ReceiverCountry == "" {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) > 1 {
					m.ReceiverCountry = strings.TrimSpace(parts[1])
				}
			}
		}

		// Email
		if isReceiver(lower) && (strings.Contains(lower, "email") || strings.Contains(lower, "mail")) {
			m.ReceiverEmail = cleanLine(line, "email", "mail")
		} else if strings.Contains(lower, "email:") || strings.Contains(lower, "mail:") {
			if m.ReceiverEmail == "" {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) > 1 {
					m.ReceiverEmail = strings.TrimSpace(parts[1])
				}
			}
		}

		// ID
		if isReceiver(lower) && (strings.Contains(lower, "id") || strings.Contains(lower, "passport") || strings.Contains(lower, "nin")) {
			m.ReceiverID = cleanLine(line, "id", "passport", "nin")
		} else if strings.Contains(lower, "id:") || strings.Contains(lower, "passport:") || strings.Contains(lower, "nin:") {
			if m.ReceiverID == "" {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) > 1 {
					m.ReceiverID = strings.TrimSpace(parts[1])
				}
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
