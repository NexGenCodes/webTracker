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

// ParseRegex extracts manifest data using regex patterns
func ParseRegex(text string) models.Manifest {
	m := models.Manifest{}

	// Common variations for "receiver" with flexible suffix
	rxVariations := `(?:receiver|reciver|recever|resiver|recieve|reciever)[s']*`

	// Receiver Name
	// 1. Specific prefix (handle "Receiver's: Name" case)
	// Simplified separator logic: match any separators + optional label + any separators
	if match := regexp.MustCompile(fmt.Sprintf(`(?i)%s[\s:']*(?:name[\s:']*)?([^\n]+)`, rxVariations)).FindStringSubmatch(text); len(match) > 1 {
		m.ReceiverName = strings.TrimSpace(match[1])
	} else {
		// 2. Generic "Name:" fallback
		if match := regexp.MustCompile(`(?i)(?:^|\n)\s*name[\s:']*([^\n]+)`).FindStringSubmatch(text); len(match) > 1 {
			m.ReceiverName = strings.TrimSpace(match[1])
		}
	}

	// Receiver Phone
	phonePatterns := []string{
		fmt.Sprintf(`(?i)%s[\s:']*(?:phone|mobile|tel|num|contact|telephone|mobil|number)[\s:']*([\+\d\s\-\(\)]+)`, rxVariations),
		`(?i)(?:phone|mobile|tel|num|contact|telephone|mobil|number)[\s:']*([\+\d\s\-\(\)]+)`,
	}
	for _, pattern := range phonePatterns {
		if match := regexp.MustCompile(pattern).FindStringSubmatch(text); len(match) > 1 {
			m.ReceiverPhone = strings.TrimSpace(match[1])
			break
		}
	}

	// Receiver Address
	if match := regexp.MustCompile(fmt.Sprintf(`(?i)%s[\s:']*(?:address|addr)[\s:']*(.+?)(?:\n|$)`, rxVariations)).FindStringSubmatch(text); len(match) > 1 {
		m.ReceiverAddress = strings.TrimSpace(match[1])
	} else if match := regexp.MustCompile(`(?i)(?:^|\n)\s*(?:address|addr)[\s:']*(.+?)(?:\n|$)`).FindStringSubmatch(text); len(match) > 1 {
		m.ReceiverAddress = strings.TrimSpace(match[1])
	}

	// Receiver Country
	if match := regexp.MustCompile(fmt.Sprintf(`(?i)%s[\s:']*country[\s:']*([^\n]+)`, rxVariations)).FindStringSubmatch(text); len(match) > 1 {
		m.ReceiverCountry = strings.TrimSpace(match[1])
	} else if match := regexp.MustCompile(`(?i)(?:^|\n)\s*country[\s:']*([^\n]+)`).FindStringSubmatch(text); len(match) > 1 {
		// Only if not "Sender Country" (simple check, sender usually explicit)
		if !strings.Contains(strings.ToLower(match[0]), "sender") {
			m.ReceiverCountry = strings.TrimSpace(match[1])
		}
	}

	// Receiver Email
	if match := regexp.MustCompile(fmt.Sprintf(`(?i)%s[\s:']*email[\s:']*([^\s]+@[^\s]+)`, rxVariations)).FindStringSubmatch(text); len(match) > 1 {
		m.ReceiverEmail = strings.TrimSpace(match[1])
	} else if match := regexp.MustCompile(`(?i)(?:^|\n)\s*email[\s:']*([^\s]+@[^\s]+)`).FindStringSubmatch(text); len(match) > 1 {
		m.ReceiverEmail = strings.TrimSpace(match[1])
	}

	// Receiver ID
	if match := regexp.MustCompile(fmt.Sprintf(`(?i)%s[\s:']*(?:id|ID)[\s:']*([A-Za-z0-9\-]+)`, rxVariations)).FindStringSubmatch(text); len(match) > 1 {
		m.ReceiverID = strings.TrimSpace(match[1])
	} else if match := regexp.MustCompile(`(?i)(?:^|\n)\s*(?:passport|id|nin|gov id)[\s:']*([A-Za-z0-9\-]+)`).FindStringSubmatch(text); len(match) > 1 {
		m.ReceiverID = strings.TrimSpace(match[1])
	}

	// Sender Name
	if match := regexp.MustCompile(`(?i)sender(?:'s|s)?\s*(?:name)?:?\s*([^\n]+)`).FindStringSubmatch(text); len(match) > 1 {
		m.SenderName = strings.TrimSpace(match[1])
	}

	// Sender Country
	if match := regexp.MustCompile(`(?i)sender(?:'s|s)?\s*country:?\s*([^\n]+)`).FindStringSubmatch(text); len(match) > 1 {
		m.SenderCountry = strings.TrimSpace(match[1])
	}

	m.Validate()
	return m
}

// ParseAI uses Gemini AI to extract manifest data with rate limiting
func ParseAI(text, apiKey string) (models.Manifest, error) {
	// Rate limit AI requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := aiRateLimiter.Wait(ctx); err != nil {
		return models.Manifest{}, fmt.Errorf("AI rate limit exceeded: %w", err)
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + apiKey

	prompt := fmt.Sprintf(`Extract shipping information from this text and return ONLY a JSON object with these exact fields:
{
  "receiverName": "",
  "receiverAddress": "",
  "receiverPhone": "",
  "receiverCountry": "",
  "receiverEmail": "",
  "receiverID": "",
  "senderName": "",
  "senderCountry": ""
}

Text: %s

Return ONLY the JSON, no explanations.`, text)

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
