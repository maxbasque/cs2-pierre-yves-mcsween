package main

import "fmt"

// CS2 economy ruleset (2023+ Counter-Strike 2 values).
const (
	MaxMoney = 16000

	// Round-win rewards.
	WinElimination  = 3250
	WinBombDetonate = 3500 // T win by detonation
	WinDefuse       = 3250 // CT win by defuse
	WinTimeExpired  = 3250 // CT win, no plant

	PlantBonus  = 800 // to the planter; the T team keeps this even on a lost round
	DefuseBonus = 300 // extra, to the defuser

	// Rough per-kill reward. GSI does not tell us the weapon used for each kill,
	// so we assume a rifle/pistol/SMG kill. AWP is $100, knife $1500, shotgun $900.
	KillReward = 300
)

// LossBonus is the end-of-round money for the losing team, by how many rounds
// they have lost in a row already (before this one): 0->1400, 1->1900 ... cap 3400.
func LossBonus(priorConsecutiveLosses int) int {
	b := 1400 + 500*priorConsecutiveLosses
	switch {
	case b > 3400:
		return 3400
	case b < 1400:
		return 1400
	default:
		return b
	}
}

// fullBuyCost is a "kitted rifle" target: rifle + full armour + utility.
func fullBuyCost(team string) int {
	if team == "CT" {
		return 5000 // M4 ~3000 + kevlar+helmet 1000 + ~2 nades + kit
	}
	return 4700 // AK 2700 + kevlar+helmet 1000 + ~2 nades
}

// Advice is a buy recommendation for the upcoming round.
type Advice struct {
	Category      string `json:"category"`
	Reason        string `json:"reason"`
	SaveNextRound int    `json:"save_next_round"` // projected money next round if you buy nothing and lose
}

// Recommend gives simple money-based buy advice.
//
//	money  - cash in hand right now (freeze time)
//	team   - "CT" or "T"
//	losses - your team's current consecutive-loss streak
func Recommend(money int, team string, losses int) Advice {
	full := fullBuyCost(team)
	saveProjection := min(money+LossBonus(losses), MaxMoney)

	switch {
	case money >= full:
		return Advice{"Full buy", "Rifle, full armour and utility for everyone.", saveProjection}

	case money >= full*3/5: // ~2800-3000
		return Advice{"Half buy", "Armour + upgraded pistol or SMG, one nade. Skip the rifle.", saveProjection}

	case saveProjection >= full:
		return Advice{"Eco / save", fmt.Sprintf(
			"Buy nothing (maybe a pistol). Even if you lose you'll have ~$%d next round — a guaranteed full buy.", saveProjection), saveProjection}

	case money >= 2000:
		return Advice{"Force buy", "Armour + pistols/SMGs + nades as a team. You can't afford to save into a strong enemy buy.", saveProjection}

	default:
		return Advice{"Full eco", "Pistol only, stack for a full buy in 1-2 rounds. Consider a group rush to trade kills.", saveProjection}
	}
}
