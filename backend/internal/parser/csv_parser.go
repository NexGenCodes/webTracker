package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"webtracker-bot/internal/models"
)

// ParseCSV extracts multiple shipments from a standard CSV payload.
// Expected Headers (case-insensitive, roughly):
// SenderName, SenderPhone, Origin, ReceiverName, ReceiverPhone, Destination, CargoType, Weight
func ParseCSV(payload string) ([]models.Manifest, error) {
	reader := csv.NewReader(strings.NewReader(payload))
	reader.TrimLeadingSpace = true

	// Read headers
	headers, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("empty CSV")
		}
		return nil, err
	}

	for i, h := range headers {
		headers[i] = strings.ToLower(strings.TrimSpace(h))
	}

	var manifests []models.Manifest

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		m := models.Manifest{}
		for i, val := range record {
			if i >= len(headers) {
				continue
			}
			col := headers[i]
			val = strings.TrimSpace(val)

			if strings.Contains(col, "sender") && strings.Contains(col, "name") {
				m.SenderName = val
			} else if strings.Contains(col, "receiver") || strings.Contains(col, "recipient") {
				if strings.Contains(col, "name") {
					m.ReceiverName = val
				} else if strings.Contains(col, "phone") {
					m.ReceiverPhone = val
				} else if strings.Contains(col, "address") {
					m.ReceiverAddress = val
				} else if strings.Contains(col, "email") {
					m.ReceiverEmail = val
				}
			} else if strings.Contains(col, "origin") || (strings.Contains(col, "sender") && strings.Contains(col, "country")) {
				m.SenderCountry = val
			} else if strings.Contains(col, "dest") || (strings.Contains(col, "receiver") && strings.Contains(col, "country")) {
				m.ReceiverCountry = val
			} else if strings.Contains(col, "cargo") || strings.Contains(col, "type") || strings.Contains(col, "item") {
				m.CargoType = val
			} else if strings.Contains(col, "weight") {
				fmt.Sscanf(val, "%f", &m.Weight)
			}
		}

		// Set defaults if empty
		if m.SenderCountry == "" {
			m.SenderCountry = "Processing Center"
		}
		if m.ReceiverCountry == "" {
			m.ReceiverCountry = "Local Delivery"
		}
		if m.CargoType == "" {
			m.CargoType = "consignment box"
		}
		if m.Weight == 0 {
			m.Weight = 15.0
		}

		m.Validate()
		manifests = append(manifests, m)
	}

	return manifests, nil
}
