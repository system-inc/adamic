# TypeScript meter inputs

| Input | Pin |
| --- | --- |
| Baseline compiler | `491e020de2aa138d7b608611c40a641571a5a1eb` |
| Candidate checker and lowerer | `7852f8a484ac135c34fc6abd5c855e0aa6063e9b` |
| Candidate on the next main tip | `154e142d26a991681907c2a8c2b649795f913b7f` |
| Cohere | `5360753e19e4417675a90e92a1823bea66b13918` |
| Cohere TypeScript submodule | `d92d9bfee114c80be2c375d72edae966176e3a4f` |
| TypeScript source under measurement | v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8` |

The rebase from `7852f8a48` to `154e142d2` brought in a JavaScript-backend change. The 79-root census has no JavaScript or native backend dependency. The entry meter also consumes checker and lowerer observations only. The rebase left both tools' sources, the lowerer, submodule pins and adapters identical. The subsequent main commit `7ed9da77f6125271b18de835f5391abf056dc274` changed object-literal dictionary lowering and is outside this comparison.

The daily meter's locked `npm ci --ignore-scripts --no-audit --no-fund` and `stage3/apply.sh` prepared one adapted tree under the pinned upstream checkout. Both runs use those exact bytes. First stops come from `stage3/census/tool` pointed at `src/compiler` with `CENSUS_MODULE_FORMAT=CommonJSBundle`. That directory walk also reports three non-TypeScript entries and an aggregate row. The per-root comparison uses the 79 TypeScript roots.

The entry binaries use `stage3/census/latent/make_overlay.py` and `stage3/meter/entry_overlay.py` with formatted overlay sources and `go build -buildvcs=false -overlay=...`. Their input is `src/tsc/tsc.ts`, with `LATENT_FULL=0`, `LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_TAIL=1` and `LATENT_TIMES` enabled. The candidate's entry pass is queued after the baseline completion trailer.
