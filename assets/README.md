# Buy-type images

One pre-composed image per buy category. The server embeds this folder and serves
it at `/assets/`; the page swaps the whole image when the advice changes.

| File | Category |
|------|----------|
| `full-buy.png` | Full buy |
| `half-buy.png` | Half buy |
| `eco.png` | Eco / save **and** Full eco |
| `force-buy.png` | Force buy |

`base.png` is not used by the app.

AWPer overrides (optional): add `full-buy-awp.png` / `half-buy-awp.png` / etc. —
same name with `-awp`. Used when the AWPer toggle is on; falls back to the base
image if absent.

All files are 1920×1080 PNG. Keep new ones the same size so the panel doesn't
resize between states.
