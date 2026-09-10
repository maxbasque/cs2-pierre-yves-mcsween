// cs2-pierre-yves-mcsween is a tiny CS2 economy adviser.
// "En as-tu vraiment besoin?" — it tells you whether to buy.
//
// It serves a one-page dashboard fed by live data from Counter-Strike 2 via Game
// State Integration (GSI). GSI is an official Valve feature: the game POSTs JSON
// to this server. It is read-only and carries no VAC risk (no injection, no
// memory reading).
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"
	"sync"
)

//go:embed index.html
var indexHTML []byte

// assetsFS holds the buy-type mannequin images, served at /assets/.
//
//go:embed assets
var assetsFS embed.FS

// liveState is the most recent state parsed from GSI. The buy advice is not
// stored here — it depends on the viewer's chosen role, so handleState computes
// it per request.
type liveState struct {
	Have   bool   `json:"have"` // has any GSI payload arrived?
	Money  int    `json:"money"`
	Team   string `json:"team"`
	Losses int    `json:"losses"`
	Round  int    `json:"round"`
	Phase  string `json:"phase"` // map phase: warmup, live, ...
}

var (
	liveMu sync.Mutex
	live   liveState
)

// gsiPayload is the subset of the GSI JSON we care about.
type gsiPayload struct {
	Map *struct {
		Phase  string `json:"phase"`
		Round  int    `json:"round"`
		TeamCT struct {
			ConsecutiveRoundLosses int `json:"consecutive_round_losses"`
		} `json:"team_ct"`
		TeamT struct {
			ConsecutiveRoundLosses int `json:"consecutive_round_losses"`
		} `json:"team_t"`
	} `json:"map"`
	Player *struct {
		Team  string `json:"team"`
		State struct {
			Money int `json:"money"`
		} `json:"state"`
	} `json:"player"`
}

func handleGSI(w http.ResponseWriter, r *http.Request) {
	var p gsiPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	// In a live match GSI only reports our own player block; ignore payloads
	// that don't have it (spectating someone else, menu, etc).
	if p.Player == nil || p.Map == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	liveMu.Lock()
	live.Have = true
	live.Money = p.Player.State.Money
	live.Team = p.Player.Team
	live.Round = p.Map.Round
	live.Phase = p.Map.Phase
	switch p.Player.Team {
	case "CT":
		live.Losses = p.Map.TeamCT.ConsecutiveRoundLosses
	case "T":
		live.Losses = p.Map.TeamT.ConsecutiveRoundLosses
	}
	liveMu.Unlock()

	w.WriteHeader(http.StatusOK)
}

// handleState returns the live state plus buy advice for the requested role
// (/api/state?role=awp; anything else, including empty, is treated as rifler).
func handleState(w http.ResponseWriter, r *http.Request) {
	liveMu.Lock()
	s := live
	liveMu.Unlock()

	var advice Advice
	if s.Have && s.Team != "" {
		advice = Recommend(s.Money, s.Team, s.Losses, r.URL.Query().Get("role"))
	}

	writeJSON(w, struct {
		liveState
		Advice Advice `json:"advice"`
	}{s, advice})
}

// handlePreview computes advice for arbitrary inputs without touching live
// state: /api/preview?money=6500&team=T&losses=0&round=5&role=awp. The debug
// buttons use it so previewing a verdict never clobbers real GSI data — the next
// /api/state poll shows the real game again.
func handlePreview(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	money, _ := strconv.Atoi(q.Get("money"))
	losses, _ := strconv.Atoi(q.Get("losses"))
	round, _ := strconv.Atoi(q.Get("round"))
	team := q.Get("team")
	if team != "CT" {
		team = "T"
	}
	writeJSON(w, struct {
		liveState
		Advice Advice `json:"advice"`
	}{
		liveState{Have: true, Money: money, Team: team, Losses: losses, Round: round, Phase: "preview"},
		Recommend(money, team, losses, q.Get("role")),
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	addr := flag.String("addr", "127.0.0.1:16000", "listen address") // 16000 = CS2 max money
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})
	http.HandleFunc("/gsi", handleGSI)
	http.HandleFunc("/api/state", handleState)
	http.HandleFunc("/api/preview", handlePreview)
	http.Handle("/assets/", http.FileServer(http.FS(assetsFS)))

	log.Printf("listening on http://%s  (GSI endpoint: http://%s/gsi)", *addr, *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
