package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connection from your frontend
	},
}

type BotRequest struct {
	Token     string `json:"token"`
	WaveCount string `json:"wave_count"`
	Stake     string `json:"stake"`
	Duration  string `json:"duration"`
}

func startBotHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BotRequest
	json.NewDecoder(r.Body).Decode(&req)

	log.Printf("Armed system: %s, Stake: %s", req.WaveCount, req.Stake)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "System Armed! Waiting for TradingView Confluence...",
	})
}

// Market Data WebSocket Proxy (True Bi-Directional)
func marketDataHandler(w http.ResponseWriter, r *http.Request) {
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer clientConn.Close()

	// Connect to Deriv from Vultr Server
	derivConn, _, err := websocket.DefaultDialer.Dial("wss://ws.binaryws.com/websockets/v3?app_id=1089", nil)
	if err != nil {
		log.Println("Failed to connect to Deriv from server:", err)
		return
	}
	defer derivConn.Close()

	// Channel 1: Read from Deriv, send to Browser
	go func() {
		for {
			msgType, message, err := derivConn.ReadMessage()
			if err != nil {
				return
			}
			if err := clientConn.WriteMessage(msgType, message); err != nil {
				return
			}
		}
	}()

	// Channel 2: Read from Browser, send to Deriv
	for {
		msgType, message, err := clientConn.ReadMessage()
		if err != nil {
			break
		}
		if err := derivConn.WriteMessage(msgType, message); err != nil {
			break
		}
	}
}
func main() {
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)
	http.HandleFunc("/api/start", startBotHandler)
	http.HandleFunc("/ws/market", marketDataHandler)

	log.Println("Server running on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
