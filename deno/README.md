# Hosted version (Deno Deploy)

Runs the adviser as a website. Friends install nothing — they open a link and
drop one `.cfg` into CS2.

## How it works

`main.ts` is a tiny relay:

| Route                  | Who calls it          | Does                                 |
| ---------------------- | --------------------- | ------------------------------------ |
| `POST /i/<token>`      | CS2 (GSI, over HTTPS) | stores your latest state             |
| `GET /s/<token>?role=` | the page, every 2 s   | returns state + buy advice           |
| `GET /new`             | the page              | mints a fresh random token           |
| `GET /`                | the browser           | the page itself                      |
| `GET /assets/*.png`    | the page              | mannequin images (from `../assets/`) |

State lives in Deno KV for 5 minutes per token, then expires. Nothing is logged.
CS2 and the browser never talk to each other — this is just a mailbox keyed by a
per-person token.

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

## Deploy (free)

1. Push this branch to GitHub.
2. <https://dash.deno.com> → **New Project** → **Deploy from GitHub repo**. If
   you land on the newer dashboard, pick **Deploy Classic** — this uses Deno KV,
   which is a Classic feature.
3. Select the repo, then:
   - **Branch:** `deno-hosted-relay` (or `main` once merged)
   - **Entrypoint:** `deno/main.ts`
   - Build command / install step: leave empty
4. Deploy. You get `https://<name>.deno.dev`; KV is provisioned automatically.
   Every `git push` redeploys.

Free tier: 1M requests/month, 100 GB bandwidth — far more than a few friends
(with `throttle "1.0"` in the cfg, ≈1,000 requests per person per match).

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
