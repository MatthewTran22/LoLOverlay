package lcu

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// EventType represents LCU WebSocket event types
type EventType int

const (
	EventTypeSubscribe   EventType = 5
	EventTypeUnsubscribe EventType = 6
	EventTypeEvent       EventType = 8
)

// ChampSelectSession represents the champion select session data
type ChampSelectSession struct {
	GameID     int64              `json:"gameId"`
	Timer      ChampSelectTimer   `json:"timer"`
	MyTeam     []ChampSelectPlayer `json:"myTeam"`
	TheirTeam  []ChampSelectPlayer `json:"theirTeam"`
	Actions    [][]ChampSelectAction `json:"actions"`
	LocalPlayerCellID int `json:"localPlayerCellId"`
}

type ChampSelectTimer struct {
	Phase            string `json:"phase"`
	TotalTimeInPhase int    `json:"totalTimeInPhase"`
	TimeLeftInPhase  int    `json:"timeLeftInPhase"`
}

type ChampSelectPlayer struct {
	CellID           int    `json:"cellId"`
	ChampionID       int    `json:"championId"`
	SummonerID       int64  `json:"summonerId"`
	AssignedPosition string `json:"assignedPosition"`
	Position         string `json:"position"`         // Alternative field
	SelectedPosition string `json:"selectedPosition"` // Another alternative
	Team             int    `json:"team"`
}

// GetPosition returns the player's position from available fields
func (p *ChampSelectPlayer) GetPosition() string {
	if p.AssignedPosition != "" {
		return p.AssignedPosition
	}
	if p.Position != "" {
		return p.Position
	}
	if p.SelectedPosition != "" {
		return p.SelectedPosition
	}
	return ""
}

type ChampSelectAction struct {
	ID          int  `json:"id"`
	ActorCellID int  `json:"actorCellId"`
	ChampionID  int  `json:"championId"`
	Type        string `json:"type"` // "pick", "ban"
	Completed   bool `json:"completed"`
	IsInProgress bool `json:"isInProgress"`
}

// WebSocketMessage represents a message from LCU WebSocket
type WebSocketMessage struct {
	Type    EventType
	Event   string
	Payload json.RawMessage
}

// ChampSelectHandler is called when champ select state changes
type ChampSelectHandler func(session *ChampSelectSession, inChampSelect bool)

// WebSocketClient handles LCU WebSocket connection
type WebSocketClient struct {
	conn            *websocket.Conn
	credentials     *Credentials
	mu              sync.Mutex
	isConnected     bool
	stopChan        chan struct{}
	champSelectHandler ChampSelectHandler
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient() *WebSocketClient {
	return &WebSocketClient{
		stopChan: make(chan struct{}),
	}
}

// Connect establishes WebSocket connection to LCU
func (w *WebSocketClient) Connect(creds *Credentials) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isConnected {
		return nil
	}

	w.credentials = creds

	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	url := fmt.Sprintf("wss://127.0.0.1:%s", creds.Port)
	header := http.Header{}
	header.Set("Authorization", "Basic "+basicAuth("riot", creds.Password))

	conn, _, err := dialer.Dial(url, header)
	if err != nil {
		return fmt.Errorf("failed to connect to LCU WebSocket: %w", err)
	}

	w.conn = conn
	w.isConnected = true

	// Subscribe to champ select events
	if err := w.subscribe("OnJsonApiEvent_lol-champ-select_v1_session"); err != nil {
		w.conn.Close()
		w.isConnected = false
		return fmt.Errorf("failed to subscribe to champ select: %w", err)
	}

	// Start listening for messages
	go w.listen()

	return nil
}

// subscribe sends a subscription message for an event
func (w *WebSocketClient) subscribe(event string) error {
	msg := []interface{}{EventTypeSubscribe, event}
	return w.conn.WriteJSON(msg)
}

// listen reads messages from the WebSocket
func (w *WebSocketClient) listen() {
	defer func() {
		w.mu.Lock()
		w.isConnected = false
		if w.conn != nil {
			w.conn.Close()
		}
		w.mu.Unlock()
	}()

	for {
		select {
		case <-w.stopChan:
			return
		default:
			_, message, err := w.conn.ReadMessage()
			if err != nil {
				return
			}

			w.handleMessage(message)
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (w *WebSocketClient) handleMessage(data []byte) {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return
	}

	if len(raw) < 3 {
		return
	}

	var eventType EventType
	if err := json.Unmarshal(raw[0], &eventType); err != nil {
		return
	}

	if eventType != EventTypeEvent {
		return
	}

	var eventName string
	if err := json.Unmarshal(raw[1], &eventName); err != nil {
		return
	}

	if eventName == "OnJsonApiEvent_lol-champ-select_v1_session" {
		w.handleChampSelectEvent(raw[2])
	}
}

// handleChampSelectEvent processes champ select events
func (w *WebSocketClient) handleChampSelectEvent(payload json.RawMessage) {
	var eventData struct {
		EventType string          `json:"eventType"`
		URI       string          `json:"uri"`
		Data      json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(payload, &eventData); err != nil {
		return
	}

	if w.champSelectHandler == nil {
		return
	}

	switch eventData.EventType {
	case "Create", "Update":
		var session ChampSelectSession
		if err := json.Unmarshal(eventData.Data, &session); err != nil {
			fmt.Printf("Failed to parse session: %v\n", err)
			return
		}

		// Debug: print myTeam data
		fmt.Printf("LocalPlayerCellID: %d, MyTeam count: %d\n", session.LocalPlayerCellID, len(session.MyTeam))
		for i, p := range session.MyTeam {
			fmt.Printf("  Player %d: CellID=%d, ChampID=%d, Assigned='%s', Pos='%s', Selected='%s'\n",
				i, p.CellID, p.ChampionID, p.AssignedPosition, p.Position, p.SelectedPosition)
		}

		w.champSelectHandler(&session, true)
	case "Delete":
		w.champSelectHandler(nil, false)
	}
}

// SetChampSelectHandler sets the callback for champ select events
func (w *WebSocketClient) SetChampSelectHandler(handler ChampSelectHandler) {
	w.champSelectHandler = handler
}

// Disconnect closes the WebSocket connection
func (w *WebSocketClient) Disconnect() {
	w.mu.Lock()
	defer w.mu.Unlock()

	close(w.stopChan)
	if w.conn != nil {
		w.conn.Close()
		w.conn = nil
	}
	w.isConnected = false
	w.stopChan = make(chan struct{})
}

// IsConnected returns whether the WebSocket is connected
func (w *WebSocketClient) IsConnected() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.isConnected
}

// basicAuth encodes credentials for basic auth
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64Encode(auth)
}

func base64Encode(s string) string {
	const base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var result []byte
	for i := 0; i < len(s); i += 3 {
		var n uint32
		remaining := len(s) - i
		switch remaining {
		case 1:
			n = uint32(s[i]) << 16
			result = append(result, base64Table[n>>18], base64Table[(n>>12)&0x3F], '=', '=')
		case 2:
			n = uint32(s[i])<<16 | uint32(s[i+1])<<8
			result = append(result, base64Table[n>>18], base64Table[(n>>12)&0x3F], base64Table[(n>>6)&0x3F], '=')
		default:
			n = uint32(s[i])<<16 | uint32(s[i+1])<<8 | uint32(s[i+2])
			result = append(result, base64Table[n>>18], base64Table[(n>>12)&0x3F], base64Table[(n>>6)&0x3F], base64Table[n&0x3F])
		}
	}
	return string(result)
}
