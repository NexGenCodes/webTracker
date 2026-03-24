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

	"webtracker-bot/internal/models"

	"golang.org/x/time/rate"
)

// AI rate limiter: 5 requests per second
var aiRateLimiter = rate.NewLimiter(rate.Every(200*time.Millisecond), 5)

var stopLabels = `(?:receiver|reciver|sender|sendr|phone|mobile|tel|num|contact|address|addr|country|nation|state|city|id|passport|email|cargo|item|content|weight|wgt|name|to|from|origin|dest|destination)`
var labelSep = `[\s]*[:\-]+[\s]*`

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

func GetLabelMappings() []uncompiledLabelMap {
	return []uncompiledLabelMap{
		{"ReceiverName", `(?i)\b(?:receiver|reciver|recever|resiver|receive|recieve|rcvr|consignment|consignee|destinatario|destinatário|dirección|empfänger|namen?)[s']*\b[\s\-:]+name\b[\s\-:]*`, 2},
		{"ReceiverName", `(?i)\b(?:receiver|reciver|recever|resiver|receive|recieve|rcvr|consignment|consignee|destinatario|destinatário|empfänger)\b[\s\-:]*`, 2},
		{"ReceiverName", `(?i)\bname\b[\s\-:]*`, 1},
		{"ReceiverPhone", `(?i)\b(?:receiver|reciver|recever|resiver|receive|recieve|rcvr|to|consignment|consignee|destinatario|destinatário|empfänger)[s']*\b[\s\-:]+\b(?:phone|mobile|tel|num|contact|telephone|mobil|number|ph|cell|whatsapp|telefone|teléfono|telephon|handy|nr)\b[\s\-:]*`, 2},
		{"ReceiverPhone", `(?i)\b(?:phone|mobile|tel|num|contact|telephone|mobil|number|ph|cell|whatsapp|telefone|teléfono|telephon|handy|nr)\b[\s\-:]*`, 1},
		{"ReceiverAddress", `(?i)\b(?:receiver|reciver|recever|resiver|receive|recieve|rcvr|to|consignment|consignee|destinatario|destinatário|empfänger)[s']*\b[\s\-:]+\b(?:address|addr|street|location|addres|addrs|dir|direction|dirección|morada|adresse|straße|strasse)\b[\s\-:]*`, 2},
		{"ReceiverAddress", `(?i)\b(?:address|addr|street|location|addres|addrs|dir|direction|dirección|morada|adresse|straße|strasse)\b[\s\-:]*`, 1},
		{"ReceiverCountry", `(?i)\b(?:receiver|reciver|recever|resiver|receive|recieve|rcvr|to|consignment|consignee|destinatario|destinatário|empfänger)[s']*\b[\s\-:]+\b(?:country|nation|state|city|pais|land|dest|destination|país|stadt|land|ort)\b[\s\-:]*`, 2},
		{"ReceiverCountry", `(?i)\b(?:country|nation|state|city|pais|land|dest|destination|to|país|stadt|land|ort)\b[\s\-:]*`, 2},
		{"ReceiverID", `(?i)\b(?:id|passport|passport\s*num|id\s*num|identity|identification|tin|nin|ssn|dni|passaporte|ausweis)\b[\s\-:]*`, 1},
		{"ReceiverID", `(?i)\b(?:receiver|reciver|recever|resiver|receive|recieve|rcvr|to|consignment|consignee)[s']*\b[\s\-:]+\b(?:id|passport|passport\s*num|id\s*num|identity|identification|tin|nin|ssn)\b[\s\-:]*`, 2},
		{"ReceiverEmail", `(?i)\b(?:receiver|reciver|recever|resiver|receive|recieve|rcvr|to|consignment|consignee)[s']*\b[\s\-:]+\b(?:email|mail|e-mail)\b[\s\-:]*`, 2},
		{"ReceiverEmail", `(?i)\b(?:email|mail|e-mail)\b[\s\-:]*`, 1},
		{"SenderName", `(?i)\b(?:sender|sendr|origin|from|shippr|shipper|sent by|remetente|remitente|absender)[s']*\b[\s\-:]+name\b[\s\-:]*`, 2},
		{"SenderName", `(?i)\b(?:sender|sendr|origin|from|shippr|shipper|sent by|remetente|remitente|absender)\b[\s\-:]*`, 2},
		{"SenderCountry", `(?i)\b(?:sender|sendr|origin|from|shippr|shipper|sent by)[s']*\b[\s\-:]+\b(?:country|nation|state|city|pais|land|dest|destination)\b[\s\-:]*`, 2},
		{"CargoType", `(?i)\b(?:item|content|cargo|description|type|package|commodity|conteúdo|contenido|inhalt|ware)\b[\s\-:]*`, 1},
		{"Weight", `(?i)\b(?:weight|wgt|mass|gross\s*weight|peso|gewicht)\b[\s\-:]*`, 1},
		{"scheduled_transit_time", `(?i)\b(?:departure|transit\s*time|depart|sent\s*date|start\s*date|transit|partida|salida|abfahrt)\b[\s\-:]*`, 2},
		{"expected_delivery_time", `(?i)\b(?:arrival|delivery\s*time|arrive|expect|delivery\s*date|delivered\s*on|delivery|chegada|entrega|ankunft|zustellung)\b[\s\-:]*`, 2},
	}
}

var (
	footerRe     *regexp.Regexp
	senderLabels *regexp.Regexp
	compiledMaps []labelMap
)

func init() {
	footerRe = regexp.MustCompile(`(?im)^[\s\-_*]{3,}$|(?i)\b(?:thank\s*you|regards|best|sincerely|kind\s*regards|thanks|saludos)\b`)
	senderLabels = regexp.MustCompile(`(?i)\b(?:sender|sendr|origin|from|shippr|shipper|sent by)\b`)

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
		re := regexp.MustCompile(`([\d.]+)`)
		if match := re.FindString(weightStr); match != "" {
			fmt.Sscanf(match, "%f", &m.Weight)
		}
	}

	receiverZone := text
	if senderStartIdx != -1 {
		receiverZone = text[:senderStartIdx]
	}

	if m.ReceiverPhone == "" {
		m.ReceiverPhone = extractEntity(receiverZone, `(?i)(?:phone|mobile|tel|num|contact|telephone|mobil|number|ph|cell|whatsapp)?[\s\-:]*([\+\d \t\-\(\).]{7,}\d)`)
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
			if len(cleanLines[0]) < 40 && !regexp.MustCompile(`(?i)`+stopLabels).MatchString(cleanLines[0]) {
				m.ReceiverName = cleanLines[0]
			}
		}
		if m.ReceiverPhone == "" {
			phoneRe := regexp.MustCompile(`(?i)(?:\+?\d[\d\s\-\(\)]{7,}\d)`)
			for i := 0; i < len(cleanLines) && i < 4; i++ {
				if match := phoneRe.FindString(cleanLines[i]); match != "" {
					m.ReceiverPhone = match
					break
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
			val = strings.TrimRight(val, ".,;|\n\t ")
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
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return text
}

func ParseAI(text, apiKey string) (models.Manifest, error) {
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
	aiText := strings.TrimSpace(result.Candidates[0].Content.Parts[0].Text)
	aiText = strings.Trim(aiText, "```json")
	aiText = strings.Trim(aiText, "```")
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
