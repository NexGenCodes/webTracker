package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"webtracker-bot/internal/localdb"
	manifestParser "webtracker-bot/internal/parser"
	"webtracker-bot/internal/shipment"
	"webtracker-bot/internal/utils"
)

type Server struct {
	ldb           *localdb.Client
	token         string
	geminiKey     string
	allowedOrigin string
}

func NewServer(ldb *localdb.Client, token string, geminiKey string, allowedOrigin string) *Server {
	return &Server{ldb: ldb, token: token, geminiKey: geminiKey, allowedOrigin: allowedOrigin}
}

// Start launches the HTTP server
func (s *Server) Start(port string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/track/", s.handleTrack)
	mux.HandleFunc("/api/public/status", s.handlePublicStatus)
	mux.HandleFunc("/api/shipments", s.handleShipments)
	mux.HandleFunc("/api/shipments/", s.handleShipmentItem)
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/api/parse", s.handleParse)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      s.middleware(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return server.ListenAndServe()
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := s.allowedOrigin
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/shipments") || strings.HasPrefix(r.URL.Path, "/api/stats") || strings.HasPrefix(r.URL.Path, "/api/parse") {
			authHeader := r.Header.Get("Authorization")
			expected := "Bearer " + s.token

			if s.token != "" && authHeader != expected {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleTrack(w http.ResponseWriter, r *http.Request) {
	trackingID := r.URL.Path[len("/api/track/"):]
	if trackingID == "" {
		http.Error(w, `{"error": "tracking_id required"}`, http.StatusBadRequest)
		return
	}

	ship, err := s.ldb.GetShipment(r.Context(), trackingID)
	if err != nil {
		http.Error(w, `{"error": "shipment not found"}`, http.StatusNotFound)
		return
	}

	timeline := s.generateTimeline(ship)

	response := map[string]interface{}{
		"tracking_id":       ship.TrackingID,
		"status":            ship.Status,
		"origin":            ship.Origin,
		"destination":       ship.Destination,
		"recipient_country": ship.Destination,
		"timeline":          timeline,
		"weight":            ship.Weight,
		// Redact sensitive fields for public tracking
		"sender_name":       redactName(ship.SenderName),
		"recipient_name":    redactName(ship.RecipientName),
		"recipient_address": "Redacted for privacy",
	}

	json.NewEncoder(w).Encode(response)
}

func redactName(name string) string {
	if name == "" {
		return "N/A"
	}
	parts := strings.Split(name, " ")
	if len(parts[0]) <= 2 {
		return parts[0] + "***"
	}
	return parts[0][:2] + "******"
}

func (s *Server) handlePublicStatus(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "webtracker-api"})
}

func (s *Server) handleShipments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		shipments, err := s.ldb.ListShipments(r.Context())
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(shipments)

	case "POST":
		var input struct {
			SenderName      string  `json:"senderName"`
			SenderCountry   string  `json:"senderCountry"`
			ReceiverName    string  `json:"receiverName"`
			ReceiverCountry string  `json:"receiverCountry"`
			ReceiverNumber  string  `json:"number"`
			ReceiverEmail   string  `json:"email"`
			ReceiverAddress string  `json:"address"`
			CargoType       string  `json:"cargoType"`
			Weight          float64 `json:"weight"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, `{"error": "invalid input"}`, http.StatusBadRequest)
			return
		}

		trackingID := utils.GenerateTrackingID("AWB")
		nowUTC := time.Now().UTC()
		transitTime, outForDeliveryTime, deliveryTime := shipment.CalculateSchedule(nowUTC, input.SenderCountry, input.ReceiverCountry)

		newShip := &shipment.Shipment{
			TrackingID:           trackingID,
			UserJID:              "admin-ui",
			Status:               shipment.StatusPending,
			CreatedAt:            nowUTC,
			ScheduledTransitTime: transitTime,
			OutForDeliveryTime:   outForDeliveryTime,
			ExpectedDeliveryTime: deliveryTime,

			SenderName: input.SenderName,
			Origin:     input.SenderCountry,

			RecipientName:    input.ReceiverName,
			RecipientPhone:   input.ReceiverNumber,
			RecipientEmail:   input.ReceiverEmail,
			RecipientAddress: input.ReceiverAddress,
			Destination:      input.ReceiverCountry,

			CargoType: input.CargoType,
			Weight:    input.Weight,
		}

		// Deduplication Layer
		// Check for existing shipments for this user/recipient combo created recently (using existing FindSimilar logic)
		existingID, err := s.ldb.FindSimilarShipment(r.Context(), "admin-ui", input.ReceiverNumber)
		if err == nil && existingID != "" {
			existingShip, err := s.ldb.GetShipment(r.Context(), existingID)
			if err == nil {
				// Window: 2 minutes (120 seconds) - prevent double clicks or quick re-submits
				if time.Since(existingShip.CreatedAt) < 2*time.Minute {
					// Strict check: Same SenderName and ReceiverName
					if existingShip.SenderName == input.SenderName && existingShip.RecipientName == input.ReceiverName {
						fmt.Printf("[Deduplication] Prevented duplicate for %s -> %s\n", input.SenderName, input.ReceiverName)
						json.NewEncoder(w).Encode(map[string]string{"tracking_id": existingID})
						return
					}
				}
			}
		}

		newShip.Weight = 15.0 // STRICT: Always 15kg as per policy

		if err := s.ldb.CreateShipment(r.Context(), newShip); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"tracking_id": trackingID})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleShipmentItem(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/shipments/"):]
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "PATCH":
		var input shipment.Shipment
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, `{"error": "invalid input"}`, http.StatusBadRequest)
			return
		}

		existing, err := s.ldb.GetShipment(r.Context(), id)
		if err != nil {
			http.Error(w, `{"error": "shipment not found"}`, http.StatusNotFound)
			return
		}

		if input.Status != "" {
			existing.Status = input.Status
		}
		if input.SenderName != "" {
			existing.SenderName = input.SenderName
		}
		if input.RecipientName != "" {
			existing.RecipientName = input.RecipientName
		}
		if input.Destination != "" {
			existing.Destination = input.Destination
		}
		if input.Origin != "" {
			existing.Origin = input.Origin
		}
		// ... add other fields if needed, but these are the main ones from Admin UI

		if err := s.ldb.UpdateShipment(r.Context(), existing); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		if id == "cleanup" {
			if _, err := s.ldb.RunAgedCleanup(r.Context()); err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
				return
			}
		} else {
			if err := s.ldb.DeleteShipment(r.Context(), id); err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusOK)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	shipments, err := s.ldb.ListShipments(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
		return
	}

	stats := map[string]int{
		"total":          len(shipments),
		"pending":        0,
		"intransit":      0,
		"outfordelivery": 0,
		"delivered":      0,
		"canceled":       0,
	}

	for _, s := range shipments {
		switch s.Status {
		case shipment.StatusPending:
			stats["pending"]++
		case shipment.StatusIntransit:
			stats["intransit"]++
		case shipment.StatusOutForDelivery:
			stats["outfordelivery"]++
		case shipment.StatusDelivered:
			stats["delivered"]++
		case shipment.StatusCanceled:
			stats["canceled"]++
		}
	}

	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "invalid input"}`, http.StatusBadRequest)
		return
	}

	if input.Text == "" {
		http.Error(w, `{"error": "text is required"}`, http.StatusBadRequest)
		return
	}

	// AI Fallback
	manifest, err := manifestParser.ParseAI(input.Text, s.geminiKey)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(manifest)
}

func (s *Server) generateTimeline(ship *shipment.Shipment) []shipment.TimelineEvent {
	now := time.Now().UTC()
	var events []shipment.TimelineEvent

	events = append(events, shipment.TimelineEvent{
		Status:      "Order Placed",
		Timestamp:   ship.CreatedAt,
		Description: fmt.Sprintf("Shipment registered at %s", ship.Origin),
		IsCompleted: true,
	})

	inTransitCompleted := now.After(ship.ScheduledTransitTime) || ship.Status == shipment.StatusIntransit || ship.Status == shipment.StatusDelivered
	events = append(events, shipment.TimelineEvent{
		Status:      "In Transit",
		Timestamp:   ship.ScheduledTransitTime,
		Description: "Package has left the origin facility and is on its way",
		IsCompleted: inTransitCompleted,
	})

	deliveredCompleted := now.After(ship.ExpectedDeliveryTime) || ship.Status == shipment.StatusDelivered
	events = append(events, shipment.TimelineEvent{
		Status:      "Delivered",
		Timestamp:   ship.ExpectedDeliveryTime,
		Description: "Package has arrived at the destination",
		IsCompleted: deliveredCompleted,
	})

	return events
}
