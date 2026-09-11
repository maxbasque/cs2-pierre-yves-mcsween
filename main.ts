// Deno Deploy entrypoint.
//
// The dashboard's GitHub integration has no field to point it at a file in a
// subdirectory — it auto-detects a `main.ts` at the repo root. The actual app
// lives in deno/ (see deno/README.md); this just runs it.
import "./deno/main.ts";
