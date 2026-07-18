// Command babelgrid-sfu is the selective-forwarding unit for the transcription
// grid: it upgrades speaker connections to WebSocket, pushes audio frames into a
// worker pool that runs the Rust DSP + Python transcription pipeline, and
// broadcasts transcripts (with translations) to every listener in the room.
package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"time"

	"babelgrid/sfu/internal/config"
	"babelgrid/sfu/internal/sfu"
	"babelgrid/sfu/internal/ws"
)

// sineFrame builds one 20 ms PCM frame at the given frequency (synthetic speaker).
func sineFrame(freq float64, sr int, amp float64) []int16 {
	n := sr / 50 // 20 ms
	out := make([]int16, n)
	for i := 0; i < n; i++ {
		v := amp * math.Sin(2*math.Pi*freq*float64(i)/float64(sr))
		out[i] = int16(v * 32767)
	}
	return out
}

func main() {
	cfg := config.Load()
	mgr := sfu.NewManager(cfg)

	if cfg.Synth {
		log.Print("synthetic speaker enabled")
		go func() {
			freqs := []float64{160, 200, 240, 300}
			t := time.NewTicker(400 * time.Millisecond)
			defer t.Stop()
			i := 0
			for range t.C {
				mgr.Enqueue(sfu.AudioFrame{
					Room:    "call1",
					Speaker: "synth",
					Lang:    "en",
					SR:      16000,
					PCM:     sineFrame(freqs[i%len(freqs)], 16000, 0.6),
				})
				i++
			}
		}()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := ws.Upgrade(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		q := r.URL.Query()
		room := or(q.Get("room"), "call1")
		speaker := or(q.Get("speaker"), "anon")
		lang := or(q.Get("lang"), "en")
		mgr.Serve(conn, room, speaker, lang)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		rooms, clients := mgr.Stats()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "ok", "rooms": rooms, "clients": clients,
		})
	})

	log.Printf("babelgrid-sfu on %s (dsp=%s grid=%s workers=%d)",
		cfg.Addr, cfg.DSPURL, cfg.GridURL, cfg.Workers)
	if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
		log.Fatal(err)
	}
}

func or(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
