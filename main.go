// cs2-pierre-yves-mcsween is a tiny CS2 economy adviser.
// "En as-tu vraiment besoin?" — it tells you whether to buy.
//
// It serves a manual calculator page and, optionally, ingests live data from
// Counter-Strike 2 via Game State Integration (GSI). GSI is an official Valve
// feature: the game POSTs JSON to this server. It is read-only and carries no
// VAC risk (no injection, no memory reading).
package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"
	"sync"
)

//go:embed index.html
var indexHTML []byte

// liveState is the most recent state parsed from GSI.
type liveState struct {
	Have   bool   `json:"have"` // has any GSI payload arrived?
	Money  int    `json:"money"`
	Team   string `json:"team"`
	Losses int    `json:"losses"`
	Round  int    `json:"round"`
	Phase  string `json:"phase"` // map phase: warmup, live, ...
	Advice Advice `json:"advice"`
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
	if live.Team != "" {
		live.Advice = Recommend(live.Money, live.Team, live.Losses)
	}
	liveMu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func handleState(w http.ResponseWriter, r *http.Request) {
	liveMu.Lock()
	snapshot := live
	liveMu.Unlock()
	writeJSON(w, snapshot)
}

// handleCalc is the manual calculator: /api/calc?money=800&team=T&losses=2
func handleCalc(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	money, _ := strconv.Atoi(q.Get("money"))
	losses, _ := strconv.Atoi(q.Get("losses"))
	team := q.Get("team")
	if team != "CT" {
		team = "T"
	}
	writeJSON(w, map[string]any{
		"money":      money,
		"team":       team,
		"losses":     losses,
		"loss_bonus": LossBonus(losses),
		"advice":     Recommend(money, team, losses),
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	addr := flag.String("addr", "127.0.0.1:3000", "listen address")
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
	http.HandleFunc("/api/calc", handleCalc)

	log.Printf("listening on http://%s  (GSI endpoint: http://%s/gsi)", *addr, *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
