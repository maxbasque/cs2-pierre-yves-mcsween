# cs2-pierre-yves-mcsween

A tiny Counter-Strike 2 economy adviser. *En as-tu vraiment besoin?*

Enter your money / side / loss streak and it tells you: **Full buy**, **Half buy**,
**Eco / save**, **Force buy**, or **Full eco** — with the reasoning and your
projected money next round if you save.

Optionally it reads live data straight from the game via **Game State Integration
(GSI)**, an official Valve feature. GSI is read-only: the game POSTs JSON to this
local server. No injection, no memory reading, **no VAC risk**.

## Run

```sh
export PATH="/home/linuxbrew/.linuxbrew/bin:$PATH"   # this box only
go run .
```

Open <http://127.0.0.1:3000>. The manual calculator works immediately.

## Live game data (optional)

Copy `gamestate_integration_pym.cfg` into your CS2 config folder:

```
.../steamapps/common/Counter-Strike Global Offensive/game/csgo/cfg/
```

Launch CS2. The "Live" panel lights up once you're in a match. Works in
matchmaking, community servers, and demo playback. In a live match GSI only
reports **your own** money (the full economy table needs spectator/observer).

## Known simplifications (v0)

- Kill rewards assume $300 (rifle/pistol/SMG). AWP is $100, knife $1500, shotgun $900.
- Buy thresholds are fixed "kitted rifle" targets (~$4700 T / ~$5000 CT), not
  situational (enemy buy, man advantage, bomb down, time on clock).
- No teammate economy — advice is for you, not the team.

## Economy rules encoded

See `economy.go`. Loss bonus $1400 → $3400 (+$500 per consecutive loss). Win
rewards: elimination $3250, T detonation $3500, CT defuse $3250. Plant bonus $800
(kept on a lost round). Max money $16000.
