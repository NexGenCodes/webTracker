package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"

	"golang.org/x/time/rate"
)

// AI rate limiter: 5 requests per second
var aiRateLimiter = rate.NewLimiter(rate.Every(200*time.Millisecond), 5)

var stopLabels = `(?:receiver|reciver|sender|sendr|phone|mobile|tel|num|contact|address|addr|country|nation|state|city|id|passport|email|cargo|item|content|weight|wgt|name|to|from|origin|dest|destination)`
var sep = `[\s\-:]*`
var labelSep = `[\s]*[:\-]+[\s]*`

type anchor struct {
	field    string
	start    int
	end      int
	priority int // Higher priority labels win if they overlap
}

// ParseRegex extracts manifest data using a segmented heuristic approach.
func ParseRegex(text string) models.Manifest {
	m := models.Manifest{}
	text = CleanText(text)

	// 1. Define Label Mappings
	// Priority: 2 (Specific Label like 'Receiver Name'), 1 (Generic Label like 'Name')
	labelMaps := []struct {
		field    string
		pattern  string
		priority int
	}{
		{"ReceiverName", `(?i)\b(?:receiver|reciver|recever|resiver|receive|recieve|rcvr|to|consignment|consignee)[s']*\b[\s\-:]+name\b[\s\-:]*`, 2},
		{"ReceiverName", `(?i)\b(?:receiver|reciver|recever|resiver|receive|recieve|rcvr|to|consignment|consignee)[s']*\b[\s\-:]*`, 2},
		{"ReceiverName", `(?i)\bname\b[\s\-:]*`, 1},
		{"ReceiverPhone", `(?i)\b(?:receiver|reciver|recever|resiver|receive|recieve|rcvr|to|consignment|consignee)[s']*\b[\s\-:]+\b(?:phone|mobile|tel|num|contact|telephone|mobil|number|ph|cell|whatsapp)\b[\s\-:]*`, 2},
		{"ReceiverPhone", `(?i)\b(?:phone|mobile|tel|num|contact|telephone|mobil|number|ph|cell|whatsapp)\b[\s\-:]*`, 1},
		{"ReceiverAddress", `(?i)\b(?:receiver|reciver|recever|resiver|receive|recieve|rcvr|to|consignment|consignee)[s']*\b[\s\-:]+\b(?:address|addr|street|location|addres|addrs|dir|direction)\b[\s\-:]*`, 2},
		{"ReceiverAddress", `(?i)\b(?:address|addr|street|location|addres|addrs|dir|direction)\b[\s\-:]*`, 1},
		{"ReceiverCountry", `(?i)\b(?:receiver|reciver|recever|resiver|receive|recieve|rcvr|to|consignment|consignee)[s']*\b[\s\-:]+\b(?:country|nation|state|city|pais|land|dest|destination)\b[\s\-:]*`, 2},
		{"ReceiverCountry", `(?i)\b(?:country|nation|state|city|pais|land|dest|destination)\b[\s\-:]*`, 1},
		{"ReceiverID", `(?i)\b(?:id|passport|passport\s*num|id\s*num|identity|identification|tin|nin|ssn)\b[\s\-:]*`, 1},
		{"ReceiverID", `(?i)\b(?:receiver|reciver|recever|resiver|receive|recieve|rcvr|to|consignment|consignee)[s']*\b[\s\-:]+\b(?:id|passport|passport\s*num|id\s*num|identity|identification|tin|nin|ssn)\b[\s\-:]*`, 2},
		{"ReceiverEmail", `(?i)\b(?:receiver|reciver|recever|resiver|receive|recieve|rcvr|to|consignment|consignee)[s']*\b[\s\-:]+\b(?:email|mail|e-mail)\b[\s\-:]*`, 2},
		{"ReceiverEmail", `(?i)\b(?:email|mail|e-mail)\b[\s\-:]*`, 1},
		{"SenderName", `(?i)\b(?:sender|sendr|origin|from|shippr|shipper|sent by)[s']*\b[\s\-:]+name\b[\s\-:]*`, 2},
		{"SenderName", `(?i)\b(?:sender|sendr|origin|from|shippr|shipper|sent by)[s']*\b[\s\-:]*`, 2},
		{"SenderCountry", `(?i)\b(?:sender|sendr|origin|from|shippr|shipper|sent by)[s']*\b[\s\-:]+\b(?:country|nation|state|city|pais|land|dest|destination)\b[\s\-:]*`, 2},
		{"CargoType", `(?i)\b(?:item|content|cargo|description|type|package|commodity)\b[\s\-:]*`, 1},
		{"Weight", `(?i)\b(?:weight|wgt|mass|gross\s*weight)\b[\s\-:]*`, 1},
	}

	// 2. Identify Anchors
	var anchors []anchor
	for _, lm := range labelMaps {
		re := regexp.MustCompile(lm.pattern)
		matches := re.FindAllStringIndex(text, -1)
		for _, match := range matches {
			anchorStart := match[0]
			anchorText := text[match[0]:match[1]]

			// Quality Check:
			// A label in the middle of a line should generally have a colon/hyphen
			// to be considered a robust label, unless it's a very specific one.
			isStartOfLine := anchorStart == 0 || text[anchorStart-1] == '\n' || (anchorStart > 1 && text[anchorStart-1] == ' ' && text[anchorStart-2] == '\n')
			hasStrongSep := regexp.MustCompile(labelSep).MatchString(anchorText)

			if isStartOfLine || hasStrongSep || lm.priority > 1 {
				anchors = append(anchors, anchor{
					field:    lm.field,
					start:    match[0],
					end:      match[1],
					priority: lm.priority,
				})
			}
		}
	}

	// 3. Filter and Sort Anchors
	// Remove overlapping anchors (keep higher priority or earlier match)
	sort.Slice(anchors, func(i, j int) bool {
		if anchors[i].start != anchors[j].start {
			return anchors[i].start < anchors[j].start
		}
		return anchors[i].priority > anchors[j].priority
	})

	var filtered []anchor
	lastEnd := -1
	for _, a := range anchors {
		if a.start >= lastEnd {
			filtered = append(filtered, a)
			lastEnd = a.end
		}
	}
	anchors = filtered

	// 4. Chunk and Assign
	results := make(map[string]string)
	for i, a := range anchors {
		start := a.end
		end := len(text)
		if i+1 < len(anchors) {
			end = anchors[i+1].start
		}
		val := strings.TrimSpace(text[start:end])
		// Optimization: if the value is empty or just punctuation, don't overwrite
		if val != "" {
			// User request: remove trailing periods from saved values
			val = strings.TrimRight(val, ".")
			val = strings.TrimSpace(val)

			if _, exists := results[a.field]; !exists || a.priority > 1 {
				results[a.field] = val
			}
		}
	}

	// 5. Build Manifest
	m.ReceiverName = results["ReceiverName"]
	m.ReceiverPhone = results["ReceiverPhone"]
	m.ReceiverAddress = results["ReceiverAddress"]
	m.ReceiverCountry = results["ReceiverCountry"]
	m.ReceiverID = results["ReceiverID"]
	m.ReceiverEmail = results["ReceiverEmail"]
	m.SenderName = results["SenderName"]
	m.SenderCountry = results["SenderCountry"]
	m.CargoType = results["CargoType"]

	if weightStr, ok := results["Weight"]; ok {
		// Clean weight string (e.g., "70 kg" -> "70")
		re := regexp.MustCompile(`([\d.]+)`)
		if match := re.FindString(weightStr); match != "" {
			fmt.Sscanf(match, "%f", &m.Weight)
		}
	}

	// 6. Entity Fallback (for fields that are still empty)
	if m.ReceiverPhone == "" {
		m.ReceiverPhone = extractEntity(text, `(?i)(?:phone|mobile|tel|num|contact|telephone|mobil|number|ph|cell|whatsapp)?[\s\-:]*([\+\d \t\-\(\).]{7,}\d)`)
	}
	if m.ReceiverEmail == "" {
		m.ReceiverEmail = extractEntity(text, `([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`)
	}
	if m.Weight == 0 {
		weightStr := extractEntity(text, `(?i)(?:weight|wgt|mass|gross\s*weight)[\s\-:]*([\d.]+)\s*(?:kg|kgs|kilos|kg's)?`)
		if weightStr != "" {
			fmt.Sscanf(weightStr, "%f", &m.Weight)
		}
	}

	m.Validate()
	return m
}

func extractEntity(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	if match := re.FindStringSubmatch(text); len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func CleanText(text string) string {
	// Standardize line endings and remove weird invisible chars
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return text
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
            "senderCountry": string,
            "weight": number
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
