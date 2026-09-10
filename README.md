# cs2-pierre-yves-mcsween

A tiny Counter-Strike 2 buy adviser. *En as-tu vraiment besoin?*

It reads your money live from the match — Valve's Game State Integration, which is
read-only and carries **no VAC risk** — and calls the round: **Full buy · Half buy ·
Eco / save · Force buy · Full eco**. You get a loadout list, a **Rifler / AWPer** toggle,
and how much money you're guaranteed next round whether you buy or save.

## 1 — Get it running

### Option A: download it (no Go needed)

Grab your file from the [Releases page](https://github.com/maxbasque/cs2-pierre-yves-mcsween/releases):

| You're on | File |
|-----------|------|
| Windows | `pym-windows-amd64.exe` |
| Mac, Apple Silicon (M1/M2/M3…) | `pym-darwin-arm64` |
| Mac, Intel | `pym-darwin-amd64` |
| Linux | `pym-linux-amd64` |

Run it:

- **Windows** — double-click. SmartScreen will warn (unsigned) → *More info → Run anyway*.
- **Mac** — `xattr -d com.apple.quarantine ~/Downloads/pym-darwin-arm64` then
  `chmod +x ~/Downloads/pym-darwin-arm64 && ~/Downloads/pym-darwin-arm64`.
- **Linux** — `chmod +x pym-linux-amd64 && ./pym-linux-amd64`.

### Option B: run from source

Install [Go](https://go.dev/dl/), then from this folder:

```sh
go run .
```

---

Either way you'll see `listening on http://127.0.0.1:16000`. Open that in a browser.
It sits in a "waiting" state until the game feeds it — that's step 2.

Port 16000 taken? Add `-addr 127.0.0.1:12345`.

## 2 — Connect CS2

Copy **`gamestate_integration_pym.cfg`** (it's in the repo, and attached to each release)
into your CS2 config folder, then start CS2:

| OS | Folder |
|----|--------|
| Windows | `…\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg\` |
| Mac | `~/Library/Application Support/Steam/steamapps/common/Counter-Strike Global Offensive/game/csgo/cfg/` |
| Linux | `~/.steam/steam/steamapps/common/Counter-Strike Global Offensive/game/csgo/cfg/` |

Can't find it? In Steam: right-click **Counter-Strike 2 → Manage → Browse local files**,
then open `game/csgo/cfg`.

That's it. The panel updates every freeze time. It only sees **your own** money (the full
team economy needs spectator mode). Works in matchmaking, community servers, and demos.
No firewall changes — it's localhost only.

If you changed the port, update the `uri` line in the `.cfg` to match.

## Notes

- The numbers are a rough sanity check, not a coach: fixed buy targets (rifler ~$4,700,
  AWPer ~$6,250), no read on the enemy buy, man advantage, or clock.
- "Guaranteed next round" is the worst case — you lose the round with no kills, no plant.
- Economy constants (loss bonus $1,400 → $3,400, round wins, max $16,000) are all in
  `economy.go`.

## Cutting a release (me only)

```sh
git tag v2 && git push --tags
```

GitHub Actions (`.github/workflows/release.yml`) cross-compiles all four binaries and
publishes them to the Releases page. Nothing to build by hand.
