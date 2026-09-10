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

	// Buy-menu prices used by the adviser.
	priceAWP    = 4750
	priceKevlar = 650  // vest only
	priceArmor  = 1000 // vest + helmet
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

// awpFullCost is the target for a proper AWP buy: AWP + full armour + a little
// utility (AWPers tend to run lean on nades).
func awpFullCost(team string) int {
	if team == "CT" {
		return 6650 // AWP 4750 + armour 1000 + utility + kit
	}
	return 6250 // AWP 4750 + armour 1000 + 1-2 nades
}

// Advice is a buy recommendation for the upcoming round.
type Advice struct {
	Category string   `json:"category"` // one of the 5 buckets below
	Reason   string   `json:"reason"`   // one-line summary
	Loadout  []string `json:"loadout"`  // itemised shopping list, in buy order

	Cost                int `json:"cost"`                  // what following this advice costs now
	SaveNextRound       int `json:"save_next_round"`       // money next round if you buy nothing and lose
	GuaranteedNextRound int `json:"guaranteed_next_round"` // money next round if you buy this and lose
}

// Recommend gives simple money-based buy advice.
//
//	money  - cash in hand right now (freeze time)
//	team   - "CT" or "T"
//	losses - your team's current consecutive-loss streak
//	role   - "awp" for AWP-first economy, anything else = rifler
func Recommend(money int, team string, losses int, role string) Advice {
	// advice finishes an Advice: it fills in the money projections from the
	// category's cash cost (capped at what you actually have).
	advice := func(category, reason string, loadout []string, cost int) Advice {
		spend := min(cost, money)
		return Advice{
			Category:            category,
			Reason:              reason,
			Loadout:             loadout,
			Cost:                spend,
			SaveNextRound:       min(money+LossBonus(losses), MaxMoney),
			GuaranteedNextRound: min(max(money-spend+LossBonus(losses), 0), MaxMoney),
		}
	}

	saveProjection := min(money+LossBonus(losses), MaxMoney)
	rifleFull := fullBuyCost(team)

	rifle := "AK-47 ($2700)"
	pistolUp := "Tec-9 / Glock upgrade ($500)"
	kit := ""
	if team == "CT" {
		rifle = "M4A4 / M4A1-S ($2900)"
		pistolUp = "Five-SeveN ($500)"
		kit = " + defuse kit ($400)"
	}

	if role == "awp" {
		awpFull := awpFullCost(team)
		awpLean := priceAWP + priceKevlar // AWP + vest, no helmet/nades

		switch {
		case money >= awpFull:
			return advice("Full buy",
				"AWP, full armour and a little utility — take your angle early.",
				[]string{
					"AWP ($4750)",
					"Kevlar + helmet ($1000)",
					"Flash + smoke" + kit,
				}, awpFull)

		case money >= awpLean:
			return advice("Full buy",
				"AWP + vest. Skip the helmet and nades to lock in the pick.",
				[]string{
					"AWP ($4750)",
					"Kevlar, no helmet ($650)",
					"No nades this round",
				}, awpLean)

		case money >= rifleFull:
			return advice("Half buy",
				"Not enough to kit the AWP — take a rifle this round, AWP next.",
				[]string{
					rifle,
					"Kevlar + helmet ($1000)",
					"1 nade if you have spare" + kit,
				}, rifleFull)

		case saveProjection >= awpLean:
			return advice("Eco / save",
				fmt.Sprintf("Buy nothing — even on a loss you'll have ~$%d next round, enough to AWP.", saveProjection),
				[]string{
					"No weapons — keep your pistol",
					"Let a teammate take the space",
					fmt.Sprintf("Save → AWP next round (~$%d)", saveProjection),
				}, 0)

		case money >= rifleFull*3/5:
			return advice("Half buy",
				"No AWP economy yet — buy light with the team and stay even.",
				[]string{
					"Kevlar ($650)",
					pistolUp + " or SMG (MP9 / MAC-10 ~$1100)",
					"One nade",
				}, 2000)

		case money >= 2000:
			return advice("Force buy",
				"Can't save into the AWP from here — force with the team.",
				[]string{
					"Kevlar ($650, skip helmet)",
					pistolUp + " / cheap SMG",
					"Grab a dropped AWP if one falls",
				}, 1400)

		default:
			return advice("Full eco",
				"Pistol only. Stack for the AWP — 2+ rounds away.",
				[]string{
					"Default pistol only — no buy",
					"Group up to trade kills",
					"AWP once the bank recovers",
				}, 0)
		}
	}

	// Rifler economy.
	switch {
	case money >= rifleFull:
		loadout := []string{
			rifle,
			"Kevlar + helmet ($1000)",
			"Smoke + flash + molotov/incendiary",
		}
		if team == "CT" {
			loadout = append(loadout, "Defuse kit ($400)")
		}
		return advice("Full buy", "Rifle, full armour and utility for everyone.", loadout, rifleFull)

	case money >= rifleFull*3/5: // ~2800-3000
		return advice("Half buy", "Armour and an upgraded pistol or SMG — skip the rifle.", []string{
			"Kevlar ($650), + helmet if affordable",
			pistolUp + " or SMG (MP9 / MAC-10 ~$1100)",
			"One nade (flash or smoke)",
			"No rifle — bank the difference",
		}, 2200)

	case saveProjection >= rifleFull:
		return advice("Eco / save", "Buy nothing now — a loss still leaves you a guaranteed full buy.", []string{
			"No weapons — keep your pistol",
			"Armour only if you have spare cash",
			fmt.Sprintf("Save → ~$%d next round", saveProjection),
		}, 0)

	case money >= 2000:
		return advice("Force buy", "Can't afford to save into a strong enemy buy — commit as a team.", []string{
			"Kevlar ($650, skip helmet)",
			pistolUp + " or cheap SMG",
			"Nades only if the whole team can buy",
			"Push together and trade kills",
		}, 1400)

	default:
		return advice("Full eco", "Pistol only. Stack for a full buy in 1-2 rounds.", []string{
			"Default pistol only — no buy",
			"Group up, rush one site to trade",
			"Full buy in 1-2 rounds",
		}, 0)
	}
}
