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
	"strconv"
	"strings"
	"time"

	"webtracker-bot/internal/models"

	"golang.org/x/time/rate"
)

var (
	// AI rate limiter: 5 requests per second
	aiRateLimiter = rate.NewLimiter(rate.Every(200*time.Millisecond), 5)

	// AI circuit breaker: 3 failures, 30s initial backoff, max 120s backoff
	aiCircuitBreaker = NewCircuitBreaker(3, 30*time.Second, 120*time.Second)
)

var stopLabels = `(?i)(?:receiver|reciver|reciever|sender|sendr|phone|mobile|mob|tel|num|contact|address|addr|country|nation|state|city|id|passport|email|cargo|item|content|weight|wgt|name|to|from|origin|dest|destination|poids|remetente|absender|empfänger|destinataire|expéditeur)`
var labelSep = `[\s]*[:\-=>]+[\s]*`

type anchor struct {
	field    string
	start    int
	end      int
	priority int
}

type labelMap struct {
	field    string
	pattern  *regexp.Regexp
	priority int
}

type uncompiledLabelMap struct {
	field    string
	pattern  string
	priority int
}

func GetLabelMappings() []uncompiledLabelMap{
	// Supports standard 's, smart ’s, plural s, and plural possessive s'
	possessive := `(?:['’]s|s['’]|s)?`
	return []uncompiledLabelMap{
		{"ReceiverName", `(?i)\b(?:receiver|recipient|reciver|recever|resiver|receive|recieve|reciever|rcvr|consignee|destinatario|destinatário|dirección|empfänger|destinataire|namen?)` + possessive + `\b(?:\s+is)?[\s\-:]+name\b[\s\-:]*`, 2},
		{"ReceiverName", `(?i)\b(?:receiver|recipient|reciver|recever|resiver|receive|recieve|reciever|rcvr|consignee|destinatario|destinatário|empfänger|destinataire|to)` + possessive + `\b(?:\s+is)?[\s\-:]*`, 2},
		{"ReceiverName", `(?i)\bname\b(?:\s+is)?[\s\-:]*`, 1},
		{"ReceiverPhone", `(?i)\b(?:receiver|recipient|reciver|recever|resiver|receive|recieve|reciever|rcvr|to|consignment|consignee|destinatario|destinatário|empfänger|destinataire)` + possessive + `\b[\s\-:]+\b(?:phone|mobile|mob|teléfono|telephon|telefone|tel|num|contact|telephone|mobil|number|ph|cell|whatsapp|handy|nr)\b[\s\-:]*`, 2},
		{"ReceiverPhone", `(?i)\b(?:phone|mobile|mob|teléfono|telephon|telefone|tel|num|contact|telephone|mobil|number|ph|cell|whatsapp|handy|nr)\b[\s\-:]*`, 1},
		{"ReceiverAddress", `(?i)\b(?:receiver|recipient|reciver|recever|resiver|receive|recieve|reciever|rcvr|to|consignment|consignee|destinatario|destinatário|empfänger|destinataire)` + possessive + `\b[\s\-:]+\b(?:address|addr|street|location|addres|addrs|dir|direction|dirección|morada|adresse|straße|strasse)\b[\s\-:]*`, 2},
		{"ReceiverAddress", `(?i)\b(?:address|addr|street|location|addres|addrs|dir|direction|dirección|morada|adresse|straße|strasse)\b[\s\-:]*`, 1},
		{"ReceiverCountry", `(?i)\b(?:receiver|recipient|reciver|recever|resiver|receive|recieve|reciever|rcvr|to|consignment|consignee|destinatario|destinatário|empfänger|destinataire)` + possessive + `\b[\s\-:]+\b(?:country|nation|state|city|pais|land|dest|destination|país|stadt|land|ort)\b[\s\-:]*`, 2},
		{"ReceiverCountry", `(?i)\b(?:country|nation|state|city|pais|land|dest|destination|país|stadt|land|ort)\b[\s\-:]*`, 2},
		{"ReceiverID", `(?i)\b(?:id|passport|passport\s*num|id\s*num|identity|identification|tin|nin|ssn|dni|passaporte|ausweis)\b[\s\-:]*`, 1},
		{"ReceiverID", `(?i)\b(?:receiver|recipient|reciver|recever|resiver|receive|recieve|reciever|rcvr|to|consignment|consignee|destinatario|destinatário|empfänger|destinataire)` + possessive + `\b[\s\-:]+\b(?:id|passport|passport\s*num|id\s*num|identity|identification|tin|nin|ssn)\b[\s\-:]*`, 2},
		{"ReceiverEmail", `(?i)\b(?:receiver|recipient|reciver|recever|resiver|receive|recieve|reciever|rcvr|to|consignment|consignee|destinatario|destinatário|empfänger|destinataire)` + possessive + `\b[\s\-:]+\b(?:email|mail|e-mail)\b[\s\-:]*`, 2},
		{"ReceiverEmail", `(?i)\b(?:email|mail|e-mail)\b[\s\-:]*`, 1},
		{"SenderName", `(?i)\b(?:sender|sendr|from|shippr|shipper|sent by|remetente|remitente|absender|expéditeur|expediteur|source|origin)` + possessive + `\b[\s\-:]+name\b[\s\-:]*`, 2},
		{"SenderName", `(?i)\b(?:sender|sendr|from|shippr|shipper|sent by|remetente|remitente|absender|expéditeur|expediteur)` + possessive + `\b[\s\-:]*`, 2},
		{"SenderCountry", `(?i)\b(?:origin|sender` + possessive + `\s*country|sender` + possessive + `\s*nation|source` + possessive + `\s*country|from` + possessive + `\s*country)\b[\s\-:]*`, 2},
		{"CargoType", `(?i)\b(?:item|content|cargo|description|type|package|commodity|conteúdo|contenido|inhalt|ware|consignment)\b(?:\s+weight)?[\s\-:]*`, 1},
		{"Weight", `(?i)\b(?:weight|wgt|mass|gross\s*weight|peso|gewicht|poids)\b[\s\-:]*`, 2},
		{"scheduled_transit_time", `(?i)\b(?:departure|transit\s*time|depart|sent\s*date|start\s*date|transit|partida|salida|abfahrt)\b[\s\-:]*`, 2},
		{"expected_delivery_time", `(?i)\b(?:arrival|delivery\s*time|arrive|expect|delivery\s*date|delivered\s*on|delivery|chegada|entrega|ankunft|zustellung)\b[\s\-:]*`, 2},
	}
}

var (
	footerRe           *regexp.Regexp
	senderLabels       *regexp.Regexp
	compiledMaps       []labelMap
	weightCleanRe      *regexp.Regexp
	stopLabelsRe       *regexp.Regexp
	expressLogisticsRe *regexp.Regexp
	dashesRe           *regexp.Regexp
	phoneLinesRe       *regexp.Regexp
	weightLinesRe      *regexp.Regexp
)

func init() {
	footerRe = regexp.MustCompile(`(?im)^[ \t_*]{3,}$|(?i)\b(?:thank\s*you|regards|best|sincerely|kind\s*regards|thanks|saludos)\b`)
	senderLabels = regexp.MustCompile(`(?i)\b(?:sender[s']*|sendr|origin|from|shippr|shipper|sent by)\b`)

	weightCleanRe = regexp.MustCompile(`([\d.]+)`)
	stopLabelsRe = regexp.MustCompile(`(?i)`+stopLabels)
	expressLogisticsRe = regexp.MustCompile(`^(?i)express\s*logistics|shipping|document`)
	dashesRe = regexp.MustCompile(`^[\-]+$`)
	phoneLinesRe = regexp.MustCompile(`(?i)(?:\+?\d[\d\s\-\(\)]{7,}\d)`)
	weightLinesRe = regexp.MustCompile(`(?i)(?:^|\s)([\d.,]+)\s*(?:kg|kgs|kilos|kg's)\b`)

	// Pre-compile all regex maps to save CPU on the VPS
	rawMaps := GetLabelMappings()
	for _, raw := range rawMaps {
		compiledMaps = append(compiledMaps, labelMap{
			field:    raw.field,
			pattern:  regexp.MustCompile(raw.pattern),
			priority: raw.priority,
		})
	}
}

// ParseRegex extracts manifest data using a segmented heuristic approach.
func ParseRegex(text string) models.Manifest {
	m := models.Manifest{}

	if loc := footerRe.FindStringIndex(text); loc != nil {
		text = text[:loc[0]]
	}
	text = CleanText(text)

	senderStartIdx := -1
	if match := senderLabels.FindStringIndex(text); match != nil {
		senderStartIdx = match[0]
	}

	anchors := findAnchors(text, compiledMaps)
	anchors = sortAndFilterAnchors(anchors)
	results := chunkAndAssign(text, anchors)

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
		cleanWeight := strings.ReplaceAll(weightStr, ",", ".")
		cleanWeight = strings.ReplaceAll(cleanWeight, " ", "")
		if match := weightCleanRe.FindString(cleanWeight); match != "" {
			if w, err := strconv.ParseFloat(match, 64); err == nil {
				m.Weight = w
			}
		}
	}

	receiverZone := text
	if senderStartIdx != -1 {
		receiverZone = text[:senderStartIdx]
	}

	if m.ReceiverPhone == "" {
		m.ReceiverPhone = extractEntity(receiverZone, `(?i)(?:phone|mobile|mob|tel|num|contact|telephone|mobil|number|ph|cell|whatsapp)?[\s\-:]*([\+\d \t\-\(\).]{7,}\d)`)
	}
	if m.ReceiverEmail == "" {
		m.ReceiverEmail = extractEntity(text, `([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`)
	}
	if m.Weight == 0 {
		weightStr := extractEntity(text, `(?i)(?:weight|wgt|mass|gross\s*weight|peso|poids)[^0-9]*([\d]+[\d., ]*)\s*(?:kg|kgs|kilos|kg's)?`)
		if weightStr != "" {
			cleanWeight := strings.ReplaceAll(weightStr, ",", ".")
			cleanWeight = strings.ReplaceAll(cleanWeight, " ", "")
			if w, err := strconv.ParseFloat(cleanWeight, 64); err == nil {
				m.Weight = w
			}
		}
	}

	if m.ReceiverName == "" || m.ReceiverPhone == "" {
		tabularText := text
		if senderStartIdx != -1 {
			tabularText = receiverZone
		}

		lines := strings.Split(tabularText, "\n")
		var cleanLines []string
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l != "" {
				cleanLines = append(cleanLines, l)
			}
		}

		if m.ReceiverName == "" && len(cleanLines) > 0 {
			for _, cl := range cleanLines {
				if len(cl) < 40 && !stopLabelsRe.MatchString(cl) && !expressLogisticsRe.MatchString(cl) && !dashesRe.MatchString(cl) {
					m.ReceiverName = cl
					break
				}
			}
		}
		if m.ReceiverPhone == "" {
			for _, cl := range cleanLines {
				if match := phoneLinesRe.FindString(cl); match != "" {
					m.ReceiverPhone = match
					break
				}
			}
		}
		if m.Weight == 0 {
			for _, cl := range cleanLines {
				if match := weightLinesRe.FindStringSubmatch(cl); len(match) > 1 {
					cleanWeight := strings.ReplaceAll(match[1], ",", ".")
					cleanWeight = strings.ReplaceAll(cleanWeight, " ", "")
					if w, err := strconv.ParseFloat(cleanWeight, 64); err == nil {
						m.Weight = w
						break
					}
				}
			}
		}
	}

	m.Validate()
	return m
}

func ParseEditPairs(text string) map[string]string {
	text = CleanText(text)
	anchors := findAnchors(text, compiledMaps)
	anchors = sortAndFilterAnchors(anchors)
	results := chunkAndAssign(text, anchors)

	dbMap := map[string]string{
		"ReceiverName":           "recipient_name",
		"ReceiverPhone":          "recipient_phone",
		"ReceiverAddress":        "recipient_address",
		"ReceiverCountry":        "destination",
		"ReceiverID":             "recipient_id",
		"ReceiverEmail":          "recipient_email",
		"SenderName":             "sender_name",
		"SenderCountry":          "origin",
		"CargoType":              "cargo_type",
		"Weight":                 "weight",
		"scheduled_transit_time": "scheduled_transit_time",
		"expected_delivery_time": "expected_delivery_time",
	}

	final := make(map[string]string)
	for k, v := range results {
		if dbField, ok := dbMap[k]; ok {
			final[dbField] = v
		}
	}
	return final
}

func findAnchors(text string, mappings []labelMap) []anchor {
	var anchors []anchor
	for _, lm := range mappings {
		matches := lm.pattern.FindAllStringIndex(text, -1)
		for _, match := range matches {
			anchorStart := match[0]
			anchorText := text[match[0]:match[1]]
			isStartOfLine := anchorStart == 0 || text[anchorStart-1] == '\n'
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
	return anchors
}

func sortAndFilterAnchors(anchors []anchor) []anchor {
	sort.Slice(anchors, func(i, j int) bool {
		if anchors[i].start != anchors[j].start {
			return anchors[i].start < anchors[j].start
		}
		if anchors[i].priority != anchors[j].priority {
			return anchors[i].priority > anchors[j].priority
		}
		return (anchors[i].end - anchors[i].start) > (anchors[j].end - anchors[j].start)
	})

	var filtered []anchor
	lastEnd := -1
	for _, a := range anchors {
		if a.start >= lastEnd {
			filtered = append(filtered, a)
			lastEnd = a.end
		}
	}
	return filtered
}

func chunkAndAssign(text string, anchors []anchor) map[string]string {
	results := make(map[string]string)
	for i, a := range anchors {
		start := a.end
		end := len(text)
		if i+1 < len(anchors) {
			end = anchors[i+1].start
		}
		val := strings.TrimSpace(text[start:end])
		if val != "" {
			// Trim common delimiters that might be between fields
			if a.field != "ReceiverAddress" && a.field != "ReceiverName" {
				val = strings.TrimRight(val, ".,;|\n\t ")
			} else {
				val = strings.TrimRight(val, ";|\n\t ")
			}
			val = strings.TrimSpace(val)

			if a.field != "ReceiverAddress" && a.field != "CargoType" {
				if idx := strings.Index(val, "\n"); idx != -1 {
					val = strings.TrimSpace(val[:idx])
				}
			}
			if _, exists := results[a.field]; !exists || a.priority > 1 {
				results[a.field] = val
			}
		}
	}
	return results
}

func extractEntity(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	if match := re.FindStringSubmatch(text); len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func CleanText(text string) string {
	if text == "" {
		return ""
	}

	var sb strings.Builder
	sb.Grow(len(text))

	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' || text[i] == '\r' {
			line := strings.TrimSpace(text[start:i])
			if line != "" || (sb.Len() > 0 && sb.String()[sb.Len()-1] != '\n') {
				sb.WriteString(line)
				sb.WriteByte('\n')
			}
			if text[i] == '\r' && i+1 < len(text) && text[i+1] == '\n' {
				i++
			}
			start = i + 1
		}
	}
	
	if start < len(text) {
		sb.WriteString(strings.TrimSpace(text[start:]))
	}

	return strings.TrimSpace(sb.String())
}

func ParseAI(ctx context.Context, text, apiKey string) (models.Manifest, error) {
	if err := aiCircuitBreaker.Allow(); err != nil {
		return models.Manifest{}, fmt.Errorf("AI parsing temporarily unavailable: %w", err)
	}

	if apiKey == "" {
		return models.Manifest{}, fmt.Errorf("AI API key is missing")
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent"

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
	req.Header.Set("x-goog-api-key", apiKey)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return models.Manifest{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		aiCircuitBreaker.RecordFailure()
		body, _ := io.ReadAll(resp.Body)
		return models.Manifest{}, fmt.Errorf("AI API error %d: %s", resp.StatusCode, string(body))
	}
	aiCircuitBreaker.RecordSuccess()
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
	aiText := strings.TrimSpace(result.Candidates[0].Content.Parts[0].Text)
	aiText = strings.TrimPrefix(aiText, "```json")
	aiText = strings.TrimPrefix(aiText, "```")
	aiText = strings.TrimSuffix(aiText, "```")
	aiText = strings.TrimSpace(aiText)
	var m models.Manifest
	if err := json.Unmarshal([]byte(aiText), &m); err != nil {
		return models.Manifest{}, err
	}
	return m, nil
}

func ValidateEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func ValidatePhone(phone string) bool {
	digits := 0
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			digits++
		}
	}
	return digits >= 5
}
