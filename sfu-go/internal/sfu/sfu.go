// Package sfu is the selective-forwarding unit: it accepts audio frames from
// many speakers over WebSocket, hands each frame to a worker pool that runs the
// Rust DSP + Python transcription pipeline, and fans the resulting transcript
// (with translations) back to every listener in the room.
package sfu

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"babelgrid/sfu/internal/config"
	"babelgrid/sfu/internal/ws"
)

// AudioFrame is one chunk of PCM from a speaker.
type AudioFrame struct {
	Room    string
	Speaker string
	Lang    string
	SR      int
	PCM     []int16
}

type dspFeatures struct {
	RMS        float64  `json:"rms"`
	Voiced     bool     `json:"voiced"`
	DominantHz float64  `json:"dominant_hz"`
	Tokens     []uint32 `json:"tokens"`
}

type pipelineResp struct {
	Text         string            `json:"text"`
	Translations map[string]string `json:"translations"`
}

type Client struct {
	conn    *ws.Conn
	Speaker string
	Lang    string
	out     chan []byte
}

func (c *Client) send(b []byte) {
	select {
	case c.out <- b:
	default:
	}
}

type room struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

func (r *room) broadcast(msg []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for c := range r.clients {
		c.send(msg)
	}
}

// Manager owns the rooms, the worker pool, and the pipeline HTTP clients.
type Manager struct {
	cfg   config.Config
	http  *http.Client
	jobs  chan AudioFrame
	mu    sync.Mutex
	rooms map[string]*room
}

func NewManager(cfg config.Config) *Manager {
	m := &Manager{
		cfg:   cfg,
		http:  &http.Client{Timeout: 3 * time.Second},
		jobs:  make(chan AudioFrame, 256),
		rooms: make(map[string]*room),
	}
	for i := 0; i < cfg.Workers; i++ {
		go m.worker()
	}
	return m
}

func (m *Manager) getRoom(id string) *room {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rooms[id]
	if !ok {
		r = &room{clients: make(map[*Client]struct{})}
		m.rooms[id] = r
	}
	return r
}

func (m *Manager) postJSON(url string, in, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	resp, err := m.http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return io.EOF
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// worker runs the DSP + transcription pipeline for one frame at a time.
func (m *Manager) worker() {
	for f := range m.jobs {
		var feat dspFeatures
		if err := m.postJSON(m.cfg.DSPURL+"/process",
			map[string]any{"pcm": f.PCM, "sr": f.SR}, &feat); err != nil {
			continue
		}
		if !feat.Voiced || len(feat.Tokens) == 0 {
			continue // silence — nothing to transcribe
		}
		var pr pipelineResp
		if err := m.postJSON(m.cfg.GridURL+"/pipeline",
			map[string]any{"tokens": feat.Tokens, "targets": m.cfg.Targets}, &pr); err != nil {
			continue
		}
		out, _ := json.Marshal(map[string]any{
			"type":         "transcript",
			"room":         f.Room,
			"speaker":      f.Speaker,
			"text":         pr.Text,
			"translations": pr.Translations,
			"dominant_hz":  feat.DominantHz,
			"ts":           float64(time.Now().UnixNano()) / 1e9,
		})
		m.getRoom(f.Room).broadcast(out)
	}
}

// Enqueue submits a frame to the worker pool, dropping it if the pool is
// saturated (backpressure — a slow pipeline never blocks ingestion).
func (m *Manager) Enqueue(f AudioFrame) {
	select {
	case m.jobs <- f:
	default:
	}
}

type inMsg struct {
	Type    string  `json:"type"`
	Room    string  `json:"room"`
	Speaker string  `json:"speaker"`
	Lang    string  `json:"lang"`
	SR      int     `json:"sr"`
	PCM     []int16 `json:"pcm"`
}

// Serve runs the read/write loops for one client.
func (m *Manager) Serve(conn *ws.Conn, roomID, speaker, lang string) {
	r := m.getRoom(roomID)
	c := &Client{conn: conn, Speaker: speaker, Lang: lang, out: make(chan []byte, 64)}

	r.mu.Lock()
	r.clients[c] = struct{}{}
	r.mu.Unlock()

	go func() {
		for msg := range c.out {
			if err := conn.WriteText(msg); err != nil {
				return
			}
		}
	}()

	defer func() {
		r.mu.Lock()
		delete(r.clients, c)
		r.mu.Unlock()
		close(c.out)
		conn.Close()
	}()

	for {
		raw, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var msg inMsg
		if json.Unmarshal(raw, &msg) != nil || msg.Type != "audio" {
			continue
		}
		sr := msg.SR
		if sr == 0 {
			sr = 16000
		}
		m.Enqueue(AudioFrame{
			Room:    roomID,
			Speaker: firstNonEmpty(msg.Speaker, speaker),
			Lang:    firstNonEmpty(msg.Lang, lang),
			SR:      sr,
			PCM:     msg.PCM,
		})
	}
}

// Stats reports room and client counts for /healthz.
func (m *Manager) Stats() (rooms, clients int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.rooms {
		r.mu.RLock()
		clients += len(r.clients)
		r.mu.RUnlock()
	}
	return len(m.rooms), clients
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
