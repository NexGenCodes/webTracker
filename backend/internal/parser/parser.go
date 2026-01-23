package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"

	"golang.org/x/time/rate"
)

// AI rate limiter: 5 requests per second
var aiRateLimiter = rate.NewLimiter(rate.Every(200*time.Millisecond), 5)

// ParseRegex extracts manifest data using fuzzy regex patterns to minimize AI reliance.
func ParseRegex(text string) models.Manifest {
	m := models.Manifest{}
	text = CleanText(text)

	// Receiver Variations (Fuzzy)
	rxLabel := `(?i)(?:receiver|reciver|recever|resiver|receive|recieve|rcvr|to|dest|destination|consignment|consignee)[s']*`
	// Sender Variations (Fuzzy)
	sxLabel := `(?i)(?:sender|sendr|origin|from|shippr|shipper|sent by)[s']*`

	// 1. Receiver Name
	m.ReceiverName = extractField(text, rxLabel+`[\s:']*(?:name[\s:']*)?`, `([^\n]+)`)
	if m.ReceiverName == "" {
		m.ReceiverName = extractField(text, `(?i)(?:^|\n)\s*name[\s:']*`, `([^\n]+)`)
	}

	// 2. Receiver Phone
	phoneLabel := `(?:phone|mobile|tel|num|contact|telephone|mobil|number|ph|cell|whatsapp)`
	m.ReceiverPhone = extractField(text, rxLabel+`[\s:']*`+phoneLabel+`[\s:']*`, `([\+\d\s\-\(\).]+)`)
	if m.ReceiverPhone == "" {
		m.ReceiverPhone = extractField(text, `(?i)`+phoneLabel+`[\s:']*`, `([\+\d\s\-\(\).]+)`)
	}

	// 3. Receiver Address
	addrLabel := `(?:address|addr|street|location|addres|addrs|dir|direction)`
	m.ReceiverAddress = extractField(text, rxLabel+`[\s:']*`+addrLabel+`[\s:']*`, `(.+?)(?:\n|$)`)
	if m.ReceiverAddress == "" {
		m.ReceiverAddress = extractField(text, `(?i)`+addrLabel+`[\s:']*`, `(.+?)(?:\n|$)`)
	}

	// 4. Receiver Country
	countryLabel := `(?:country|nation|state|origin|city|pais|land)`
	m.ReceiverCountry = extractField(text, rxLabel+`[\s:']*`+countryLabel+`[\s:']*`, `([^\n]+)`)
	if m.ReceiverCountry == "" {
		// Only pick if not sender's
		match := extractField(text, `(?i)`+countryLabel+`[\s:']*`, `([^\n]+)`)
		if match != "" && !strings.Contains(strings.ToLower(text), "sender") {
			m.ReceiverCountry = match
		}
	}

	// 5. Receiver ID
	idLabel := `(?:id|passport|passport\s*num|id\s*num|identity|identification|tin|nin|ssn)`
	m.ReceiverID = extractField(text, rxLabel+`[\s:']*`+idLabel+`[\s:']*`, `([A-Z0-9\s-]+)`)
	if m.ReceiverID == "" {
		m.ReceiverID = extractField(text, `(?i)`+idLabel+`[\s:']*`, `([A-Z0-9\s-]+)`)
	}

	// 6. Receiver Email
	emailLabel := `(?:email|mail|e-mail)`
	m.ReceiverEmail = extractField(text, rxLabel+`[\s:']*`+emailLabel+`[\s:']*`, `([^\n\s]+@[^\n\s]+\.[^\n\s]+)`)
	if m.ReceiverEmail == "" {
		m.ReceiverEmail = extractField(text, `(?i)`+emailLabel+`[\s:']*`, `([^\n\s]+@[^\n\s]+\.[^\n\s]+)`)
	}

	// 7. Sender Name
	m.SenderName = extractField(text, sxLabel+`[\s:']*(?:name[\s:']*)?`, `([^\n]+)`)

	// 8. Sender Country
	m.SenderCountry = extractField(text, sxLabel+`[\s:']*`+countryLabel+`[\s:']*`, `([^\n]+)`)

	// 9. Cargo Type
	cargoLabel := `(?:item|content|cargo|description|type|package|commodity)`
	m.CargoType = extractField(text, `(?i)`+cargoLabel+`[\s:']*`, `([^\n]+)`)

	m.Validate()
	return m
}

func CleanText(text string) string {
	// Standardize line endings and remove weird invisible chars
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return text
}

func extractField(text, labelPattern, valuePattern string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(?i)%s\s*%s`, labelPattern, valuePattern))
	if match := re.FindStringSubmatch(text); len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

// ParseAI uses Gemini AI to extract manifest data with rate limiting
func ParseAI(text, apiKey string) (models.Manifest, error) {
	// Rate limit AI requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := aiRateLimiter.Wait(ctx); err != nil {
		return models.Manifest{}, fmt.Errorf("AI rate limit exceeded: %w", err)
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent?key=" + apiKey

	prompt := `You are a logistics data extraction assistant. Extract shipping information from user text and return JSON matching the schema below.
        
        TARGET SCHEMA:
        {
            "receiverName": string,
            "receiverAddress": string,
            "receiverCountry": string,
            "receiverPhone": string,
            "receiverEmail": string,
            "receiverID": string,
            "senderName": string,
            "senderCountry": string
        }

        RULES:
        1. Extract the fields from the input text.
        2. If a field is missing, use an empty string "" - DO NOT return null.
        3. Infer countries if city names are well-known (e.g. "Paris" -> "France").
        4. Phone numbers: Extract as is.
        
        Extract from this:
        ` + text

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return models.Manifest{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return models.Manifest{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return models.Manifest{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.Manifest{}, fmt.Errorf("AI API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.Manifest{}, err
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return models.Manifest{}, fmt.Errorf("no AI response")
	}

	aiText := result.Candidates[0].Content.Parts[0].Text
	aiText = strings.TrimSpace(aiText)
	aiText = strings.Trim(aiText, "```json")
	aiText = strings.Trim(aiText, "```")
	aiText = strings.TrimSpace(aiText)

	var m models.Manifest
	if err := json.Unmarshal([]byte(aiText), &m); err != nil {
		logger.Warn().Str("ai_response", aiText).Msg("Failed to parse AI JSON")
		return models.Manifest{}, err
	}

	return m, nil
}

// ValidateEmail checks if the string superficially resembles an email.
func ValidateEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// ValidatePhone checks if the string looks like a phone number (mostly digits).
func ValidatePhone(phone string) bool {
	// Must have at least a few digits
	digits := 0
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			digits++
		}
	}
	// Arbitrary minimum of 5 digits to be a "phone number"
	return digits >= 5
}
