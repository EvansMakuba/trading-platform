package main

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)

// Define the incoming request from the frontend
type BotRequest struct {
	Token     string `json:"token"`
	WaveCount string `json:"wave_count"`
}

func startBotHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BotRequest
	json.NewDecoder(r.Body).Decode(&req)

	log.Printf("Received token and wave count: %s", req.WaveCount)

	// 1. Connect to Deriv WebSocket App ID 1089 (default testing ID)
	c, _, err := websocket.DefaultDialer.Dial("wss://ws.binaryws.com/websockets/v3?app_id=1089", nil)
	if err != nil {
		w.Write([]byte(`{"message": "Failed to connect to Deriv API"}`))
		return
	}
	defer c.Close()

	// 2. Send Authorization Request
	authReq := map[string]interface{}{"authorize": req.Token}
	c.WriteJSON(authReq)

	// 3. Read Deriv Response
	var authRes map[string]interface{}
	c.ReadJSON(&authRes)

	if authRes["error"] != nil {
		json.NewEncoder(w).Encode(map[string]string{"message": "Deriv Auth Failed: Invalid Token"})
		return
	}

	// 4. Fetch Balance to prove connection
	c.WriteJSON(map[string]interface{}{"balance": 1})
	var balRes map[string]interface{}
	c.ReadJSON(&balRes)
	
	balanceData := balRes["balance"].(map[string]interface{})
	currency := balanceData["currency"].(string)
	balance := balanceData["balance"].(float64)
	
	successMsg := map[string]string{
		"message": "Successfully connected to Deriv! Balance: " + currency + " " + string(rune(int(balance))),
	}
	json.NewEncoder(w).Encode(successMsg)
}

func main() {
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)
	http.HandleFunc("/api/start", startBotHandler)

	log.Println("Server starting on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}