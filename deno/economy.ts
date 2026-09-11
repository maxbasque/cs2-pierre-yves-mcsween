// Port of economy.go — CS2 buy adviser. Keep the two in sync.

export const MAX_MONEY = 16000;
const PRICE_AWP = 4750;
const PRICE_KEVLAR = 650; // vest only

/** End-of-round money for the losing team, by consecutive losses so far. */
export function lossBonus(priorConsecutiveLosses: number): number {
  const b = 1400 + 500 * priorConsecutiveLosses;
  return Math.max(1400, Math.min(3400, b));
}

/** "Kitted rifle" target: rifle + full armour + utility. */
function fullBuyCost(team: string): number {
  return team === "CT" ? 5000 : 4700;
}

/** Proper AWP buy: AWP + full armour + a little utility (AWPers run lean on nades). */
function awpFullCost(team: string): number {
  return team === "CT" ? 6650 : 6250;
}

export interface Advice {
  category: string;
  reason: string;
  loadout: string[];
  cost: number;
  save_next_round: number;
  guaranteed_next_round: number;
}

/**
 * money  - cash in hand right now (freeze time)
 * team   - "CT" or "T"
 * losses - your team's current consecutive-loss streak
 * role   - "awp" for AWP-first economy, anything else = rifler
 */
export function recommend(
  money: number,
  team: string,
  losses: number,
  role: string,
): Advice {
  team = team === "CT" ? "CT" : "T";
  money = Math.max(0, Math.floor(money) || 0);
  losses = Math.max(0, Math.floor(losses) || 0);

  const mk = (
    category: string,
    reason: string,
    loadout: string[],
    cost: number,
  ): Advice => {
    const spend = Math.min(cost, money);
    return {
      category,
      reason,
      loadout,
      cost: spend,
      save_next_round: Math.min(money + lossBonus(losses), MAX_MONEY),
      guaranteed_next_round: Math.min(
        Math.max(money - spend + lossBonus(losses), 0),
        MAX_MONEY,
      ),
    };
  };

  const saveProjection = Math.min(money + lossBonus(losses), MAX_MONEY);
  const rifleFull = fullBuyCost(team);
  const half = Math.floor((rifleFull * 3) / 5); // ~2820 T / 3000 CT

  let rifle = "AK-47 ($2700)";
  let pistolUp = "Tec-9 / Glock upgrade ($500)";
  let kit = "";
  if (team === "CT") {
    rifle = "M4A4 / M4A1-S ($2900)";
    pistolUp = "Five-SeveN ($500)";
    kit = " + defuse kit ($400)";
  }

  if (role === "awp") {
    const awpFull = awpFullCost(team);
    const awpLean = PRICE_AWP + PRICE_KEVLAR; // AWP + vest, no helmet/nades

    if (money >= awpFull) {
      return mk(
        "Full buy",
        "AWP, full armour and a little utility — take your angle early.",
        [
          "AWP ($4750)",
          "Kevlar + helmet ($1000)",
          "Flash + smoke" + kit,
        ],
        awpFull,
      );
    }
    if (money >= awpLean) {
      return mk(
        "Full buy",
        "AWP + vest. Skip the helmet and nades to lock in the pick.",
        [
          "AWP ($4750)",
          "Kevlar, no helmet ($650)",
          "No nades this round",
        ],
        awpLean,
      );
    }
    if (money >= rifleFull) {
      return mk(
        "Half buy",
        "Not enough to kit the AWP — take a rifle this round, AWP next.",
        [
          rifle,
          "Kevlar + helmet ($1000)",
          "1 nade if you have spare" + kit,
        ],
        rifleFull,
      );
    }
    if (saveProjection >= awpLean) {
      return mk(
        "Eco / save",
        `Buy nothing — even on a loss you'll have ~$${saveProjection} next round, enough to AWP.`,
        [
          "No weapons — keep your pistol",
          "Let a teammate take the space",
          `Save → AWP next round (~$${saveProjection})`,
        ],
        0,
      );
    }
    if (money >= half) {
      return mk(
        "Half buy",
        "No AWP economy yet — buy light with the team and stay even.",
        [
          "Kevlar ($650)",
          pistolUp + " or SMG (MP9 / MAC-10 ~$1100)",
          "One nade",
        ],
        2000,
      );
    }
    if (money >= 2000) {
      return mk(
        "Force buy",
        "Can't save into the AWP from here — force with the team.",
        [
          "Kevlar ($650, skip helmet)",
          pistolUp + " / cheap SMG",
          "Grab a dropped AWP if one falls",
        ],
        1400,
      );
    }
    return mk("Full eco", "Pistol only. Stack for the AWP — 2+ rounds away.", [
      "Default pistol only — no buy",
      "Group up to trade kills",
      "AWP once the bank recovers",
    ], 0);
  }

  // Rifler economy.
  if (money >= rifleFull) {
    const loadout = [
      rifle,
      "Kevlar + helmet ($1000)",
      "Smoke + flash + molotov/incendiary",
    ];
    if (team === "CT") loadout.push("Defuse kit ($400)");
    return mk(
      "Full buy",
      "Rifle, full armour and utility for everyone.",
      loadout,
      rifleFull,
    );
  }
  if (money >= half) {
    return mk(
      "Half buy",
      "Armour and an upgraded pistol or SMG — skip the rifle.",
      [
        "Kevlar ($650), + helmet if affordable",
        pistolUp + " or SMG (MP9 / MAC-10 ~$1100)",
        "One nade (flash or smoke)",
        "No rifle — bank the difference",
      ],
      2200,
    );
  }
  if (saveProjection >= rifleFull) {
    return mk(
      "Eco / save",
      "Buy nothing now — a loss still leaves you a guaranteed full buy.",
      [
        "No weapons — keep your pistol",
        "Armour only if you have spare cash",
        `Save → ~$${saveProjection} next round`,
      ],
      0,
    );
  }
  if (money >= 2000) {
    return mk(
      "Force buy",
      "Can't afford to save into a strong enemy buy — commit as a team.",
      [
        "Kevlar ($650, skip helmet)",
        pistolUp + " or cheap SMG",
        "Nades only if the whole team can buy",
        "Push together and trade kills",
      ],
      1400,
    );
  }
  return mk("Full eco", "Pistol only. Stack for a full buy in 1-2 rounds.", [
    "Default pistol only — no buy",
    "Group up, rush one site to trade",
    "Full buy in 1-2 rounds",
  ], 0);
}
