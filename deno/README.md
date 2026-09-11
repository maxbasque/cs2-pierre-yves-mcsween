# Hosted version (Deno Deploy)

Runs the adviser as a website. Friends install nothing — they open a link and
drop one `.cfg` into CS2.

## How it works

`main.ts` is a tiny relay:

| Route                  | Who calls it          | Does                                 |
| ---------------------- | --------------------- | ------------------------------------ |
| `POST /i/<token>`      | CS2 (GSI, over HTTPS) | stores your latest state             |
| `GET /s/<token>?role=` | the page, every 2 s   | returns state + buy advice           |
| `GET /preview?…`       | the debug buttons     | advice for made-up inputs, stateless |
| `GET /new`             | the page              | mints a fresh random token           |
| `GET /`                | the browser           | the page itself                      |
| `GET /assets/*.png`    | the page              | mannequin images (from `../assets/`) |

No database — the latest payload per token is held in memory for 5 minutes. Deno
Deploy may run several isolates, so writes are fanned out to the others over a
`BroadcastChannel`. CS2 and the browser never talk to each other — this is just
a mailbox keyed by a per-person token. Nothing is logged.

## Run locally

```sh
deno task dev          # http://localhost:8000
```

Test without the game:

```sh
TOK=$(curl -s localhost:8000/new | grep -o '[a-f0-9]\{32\}')
curl -s -X POST "localhost:8000/i/$TOK" -H 'content-type: application/json' \
  -d '{"map":{"phase":"live","round":5,"team_t":{"consecutive_round_losses":1}},"player":{"team":"T","state":{"money":5500}}}'
curl -s "localhost:8000/s/$TOK?role=rifle"
```

## Deploy (free, no database to set up)

1. Push this branch to GitHub.
2. <https://app.deno.com> → your org → **New app** → connect the GitHub repo.
3. In the app config:
   - **Entrypoint:** `deno/main.ts`
   - **Branch:** `deno-hosted-relay` (or `main` once merged)
   - Build / install command: leave empty
   - No env vars, no database — the app stores state in memory.
4. Deploy. You get `https://<name>.deno.dev`. Every `git push` redeploys.

Free tier is far more than a few friends need (with `throttle "1.0"` in the cfg,
≈1,000 requests per person per match).

## Onboard a friend

Send them `https://<name>.deno.dev`. They click **Generate my link**, get a
personalised `.cfg` + a download button, save it as
`gamestate_integration_pym.cfg` in
`…/Counter-Strike Global Offensive/game/csgo/cfg/`, launch CS2, and bookmark
their link (their token is the `#…` part of the URL).

## Notes

- The token is a bearer secret — anyone with the link can watch that person's
  live economy or push junk into their view. Harmless, but don't post links
  publicly.
- GSI POSTing to a remote URL is the same outbound webhook as POSTing to
  `127.0.0.1`; no VAC implication. The only change is the data leaves the PC.
- `economy.ts` is a hand port of `../economy.go` — keep the two in sync.
