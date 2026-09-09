# cs2-pierre-yves-mcsween

A tiny Counter-Strike 2 economy adviser. *En as-tu vraiment besoin?*

Enter your money / side / loss streak and it tells you: **Full buy**, **Half buy**,
**Eco / save**, **Force buy**, or **Full eco** — with the reasoning and your
projected money next round if you save.

Optionally it reads live data straight from the game via **Game State Integration
(GSI)**, an official Valve feature. GSI is read-only: the game POSTs JSON to this
local server. No injection, no memory reading, **no VAC risk**.

Pure Go standard library, no OS-specific code — runs on **Linux, Windows, and
macOS** the same way. You need [Go](https://go.dev/dl/) 1.21+ installed.

## Run

```sh
go run .
```

On this Bazzite box, Go lives in Homebrew and isn't on the default PATH, so prefix
that one machine's commands with:

```sh
export PATH="/home/linuxbrew/.linuxbrew/bin:$PATH"
```

Open <http://127.0.0.1:16000> (port 16000 = CS2 max money). The manual calculator
works immediately. Override with `-addr 127.0.0.1:PORT`.

To build a standalone binary instead: `go build -o pym .` (or `go build -o pym.exe .`
on Windows), then run `./pym`.

## Live game data (optional)

Copy `gamestate_integration_pym.cfg` into your CS2 `cfg` folder, then launch CS2.
The folder is under your Steam library:

| OS | Path |
|----|------|
| Windows | `C:\Program Files (x86)\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg\` |
| Linux | `~/.steam/steam/steamapps/common/Counter-Strike Global Offensive/game/csgo/cfg/` |
| macOS | `~/Library/Application Support/Steam/steamapps/common/Counter-Strike Global Offensive/game/csgo/cfg/` |

Easiest way to find it on any OS: in Steam, right-click **Counter-Strike 2 →
Manage → Browse local files**, then open `game\csgo\cfg`.

The "Live" panel lights up once you're in a match. Works in matchmaking, community
servers, and demo playback. In a live match GSI only reports **your own** money
(the full economy table needs spectator/observer).

The `.cfg` points GSI at `http://127.0.0.1:16000/gsi` — if you run the server on a
different port or host, edit the `uri` line to match. Keep it on `127.0.0.1`
(localhost); no firewall changes are needed for loopback on any OS.

## Known simplifications (v0)

- Kill rewards assume $300 (rifle/pistol/SMG). AWP is $100, knife $1500, shotgun $900.
- Buy thresholds are fixed "kitted rifle" targets (~$4700 T / ~$5000 CT), not
  situational (enemy buy, man advantage, bomb down, time on clock).
- No teammate economy — advice is for you, not the team.

## Economy rules encoded

See `economy.go`. Loss bonus $1400 → $3400 (+$500 per consecutive loss). Win
rewards: elimination $3250, T detonation $3500, CT defuse $3250. Plant bonus $800
(kept on a lost round). Max money $16000.
