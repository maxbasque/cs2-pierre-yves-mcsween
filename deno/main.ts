// Hosted relay for the CS2 buy adviser.
//
// CS2's Game State Integration POSTs your state to  /i/<token>  (over the
// internet — the .cfg's uri points here). The browser page polls  /s/<token>
// and renders the advice. The two never talk directly; this server is a mailbox
// keyed by a random per-person token, holding only the latest payload for 5 min.
//
// Run locally:   deno task dev
// Deploy:        see deno/README.md (Deno Deploy, free tier)

import { type Advice, recommend } from "./economy.ts";

const kv = await Deno.openKv();
const TTL_MS = 5 * 60 * 1000;

const HTML = await Deno.readTextFile(
  new URL("./public/index.html", import.meta.url),
);

const TOKEN_RE = /^[A-Za-z0-9]{8,64}$/;

/** The slice of the GSI JSON we read. Everything is optional — payloads vary. */
interface GsiPayload {
  player?: { team?: string; state?: { money?: number } };
  map?: {
    phase?: string;
    round?: number;
    team_ct?: { consecutive_round_losses?: number };
    team_t?: { consecutive_round_losses?: number };
  };
}

interface State {
  have: boolean;
  money: number;
  team: string;
  losses: number;
  round: number;
  phase: string;
}
const EMPTY_STATE: State = {
  have: false,
  money: 0,
  team: "",
  losses: 0,
  round: 0,
  phase: "",
};
const EMPTY_ADVICE: Advice = {
  category: "",
  reason: "",
  loadout: [],
  cost: 0,
  save_next_round: 0,
  guaranteed_next_round: 0,
};

function json(data: unknown, status = 200): Response {
  return new Response(JSON.stringify(data), {
    status,
    headers: {
      "content-type": "application/json",
      "cache-control": "no-store",
    },
  });
}

function num(v: unknown): number {
  const n = Number(v);
  return Number.isFinite(n) ? n : 0;
}

async function handler(req: Request): Promise<Response> {
  const url = new URL(req.url);
  const path = url.pathname;

  // The page.
  if (req.method === "GET" && (path === "/" || path === "/index.html")) {
    return new Response(HTML, {
      headers: { "content-type": "text/html; charset=utf-8" },
    });
  }

  // Mint a fresh token (no storage — the KV entry is created on the first POST).
  if (req.method === "GET" && path === "/new") {
    return json({ token: crypto.randomUUID().replaceAll("-", "") });
  }

  // GSI ingest: CS2 POSTs here.
  const ingest = path.match(/^\/i\/([^/]+)$/);
  if (ingest && req.method === "POST") {
    const token = ingest[1];
    if (!TOKEN_RE.test(token)) {
      return new Response("bad token", { status: 400 });
    }

    let p: GsiPayload;
    try {
      p = (await req.json()) as GsiPayload;
    } catch {
      return new Response("bad json", { status: 400 });
    }
    // In a live match GSI only sends our own player block; ignore anything else
    // (spectating, menu, warmup with no map yet).
    if (p?.player && p?.map) {
      const team = p.player.team === "CT"
        ? "CT"
        : p.player.team === "T"
        ? "T"
        : "";
      const state: State = {
        have: true,
        money: num(p.player.state?.money),
        team,
        losses: team === "CT"
          ? num(p.map.team_ct?.consecutive_round_losses)
          : num(p.map.team_t?.consecutive_round_losses),
        round: num(p.map.round),
        phase: String(p.map.phase ?? ""),
      };
      await kv.set(["state", token], state, { expireIn: TTL_MS });
    }
    return new Response("ok");
  }

  // State read: the page polls here.
  const read = path.match(/^\/s\/([^/]+)$/);
  if (read && req.method === "GET") {
    const token = read[1];
    if (!TOKEN_RE.test(token)) {
      return json({ ...EMPTY_STATE, advice: EMPTY_ADVICE });
    }
    const role = url.searchParams.get("role") ?? "";
    const rec = await kv.get<State>(["state", token]);
    const s = rec.value ?? EMPTY_STATE;
    const advice = s.have && s.team
      ? recommend(s.money, s.team, s.losses, role)
      : EMPTY_ADVICE;
    return json({ ...s, advice });
  }

  // Stateless preview for the debug buttons — never touches stored state.
  if (req.method === "GET" && path === "/preview") {
    const q = url.searchParams;
    const team = q.get("team") === "CT" ? "CT" : "T";
    const money = num(q.get("money"));
    const losses = num(q.get("losses"));
    return json({
      have: true,
      money,
      team,
      losses,
      round: num(q.get("round")),
      phase: "preview",
      advice: recommend(money, team, losses, q.get("role") ?? ""),
    });
  }

  // Mannequin images, from the repo's assets/ dir (one level up).
  const asset = path.match(/^\/assets\/([a-z0-9-]+\.png)$/);
  if (asset && req.method === "GET") {
    try {
      const bytes = await Deno.readFile(
        new URL(`../assets/${asset[1]}`, import.meta.url),
      );
      return new Response(bytes, {
        headers: {
          "content-type": "image/png",
          "cache-control": "public, max-age=86400",
        },
      });
    } catch {
      return new Response("not found", { status: 404 });
    }
  }

  return new Response("not found", { status: 404 });
}

Deno.serve(handler);
