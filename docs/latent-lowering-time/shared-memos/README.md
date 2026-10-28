# Shared predicate proofs and assigned symbols in the latent overlay

Performance's build, on Lower's yes (#pv935yw, 09:47): in the latent overlay only, let every lowering
latentRun's first-error loop makes for one (program, checker) pair share `predicateProofs` (`f27967ea`) and
`assignedSymbols` (`761ee21c`). Built on the four-head tree `4a9c4665` (`a1ce733b` plus `761ee21c`
cherry-picked). No compiler branch is pushed; the change is [shared-memos.diff](shared-memos.diff), against
`4a9c4665`, under `stage3/census/latent` only.

## The change

- `lower.go.txt`: `latentSharedFacts`, one store keyed by the `*load.Program` and `*checker.Checker` pointers.
  `fresh()` seeds a new lowering from it when both pointers match; when either differs, the store empties and
  takes the new key, so facts are recomputed. After a unit (the per-file refusal scan, and each statement) the
  store takes the lowering's facts back, only when the unit returned normally **and** recovered no panic anywhere.
- The three registration handlers that swallowed panics (`defer func() { _ = recover() }()`, two in
  `lower.go.txt`, one in `make_overlay.py`'s enum registration text) now call `latentSwallow`, which still
  swallows but marks the unit unclean.
- The entries memo (`entriesRecords`, `entriesProgramRecords*`) stays per lowering, as Lower said.

## Lower's conditions

1. **Keyed by both pointers, recomputed when either differs.** `seed` compares `program` and `checker`; on a
   mismatch it empties the store before anything is seeded, and `take` keeps facts only under the store's own key.
2. **No AST change between statements (read, not measured).** Of `fresh()`'s inputs, the program and its files hold
   still: `statementrewrite` renames the compiler's Go `lowering.statement` to `latentStatementRaw` behind a
   wrapper, and `refusalrewrite` clones the refusal scan's Go, so both rewrite the compiler, never the program.
   The overlay's `Lower`, whose async normalization is the one source rewrite in production, never runs (its
   body is cut). A search of `internal/lower` and the generated overlay for writes to `.Parent`, `.Flags`, node
   lists, `.Text` or `.Kind` finds only IR writes (`ir.Function.Body` and the like). The state copier
   (`latentSnapshot`) runs only in full and speculative modes, not first-error. Nothing I found edits or excises
   source between statements, so I built it. The comment on `latentSharedFacts` says this.
   `l.symbol`, which `assignedSymbol` keys by, reads only the checker (`GetSymbolAtLocation`, `GetAliasedSymbol`,
   `GetExportSymbolOfSymbol`).
3. **Proof:** the test and mutants below, and the byte diff and back-to-back wall time in "The run" below.

The trap Performance named, and one more: `assignedSymbol` fills `l.assignedSymbols` during its walk, so a walk
cut short leaves a partial map. latentRun's own `recover` is one way out, which `take`'s `clean` covers. The other:
`latentRegister` swallows panics around `enumLocal`, `needsStatics`/`declareLocal`/`staticStorage` and
`declareLocal`, and the unit then returns normally. Had one of those reached `assignedSymbol` and panicked
mid-walk, the partial map would pass a plain "returned normally" check. So `clean` also requires that no
swallowed panic happened in the unit (`latentUnitRecovered`). `predicateFixedPoint` stores its declarations and
each mode only once they're complete; it takes the same rule anyway. A seeded lowering gets its own copy of the
proofs' mode table, so a mode a failed unit adds never reaches the store. Seeded assigned symbols are shared as they
are, since `assignedSymbol` writes only into a nil map.

## Test and mutants (observed)

No harness ran an overlay test as is: `stage3/census/speculative/audit_snapshot.py` overlays
`scalar_copy_test.go.txt` into `internal/lower`, but on today's tree the package's own tests stop building under
the overlay (`undefined: lowerChecked`, since the overlay cuts `Lower`'s body). The smallest harness that works is
the same shape, plus one thing: `audit_shared_facts.py` (in [shared-memos.diff](shared-memos.diff)) maps every other
`internal/lower/*_test.go` to an empty replacement, which deletes it from the build, and maps
`shared_facts_test.go.txt` in as `internal/lower/latent_shared_facts_test.go`. It runs
`go test -overlay=... ./internal/lower -run '^(three tests)$'` on the meter's gofmt'd overlay, then once per mutant.

```
python3 stage3/census/latent/audit_shared_facts.py <tree> <latent-overlay> <out>
```

[audit/audit-summary.txt](audit/audit-summary.txt):

| run | result | the test that went red |
|---|---|---|
| baseline | PASS, all three | |
| shares across a changed checker (the store keyed by program alone) | caught | `TestLatentSharedFactsKeyedByProgramAndChecker`: "a lowering of another checker was seeded with the first checker's facts" |
| takes from every unit, the panic path included | caught | `TestLatentSharedFactsTakeOnlyCleanUnits`: "an unclean unit's facts reached the store" |
| seeds the store's own mode table | caught | `TestLatentSharedFactsTakeOnlyCleanUnits`: "a mode an unclean unit added reached the store" |
| a swallowed registration panic leaves the unit clean | caught | `TestLatentSwallowMarksTheUnit`: "a swallowed panic left its unit clean" |

Each mutant turns exactly one test red and leaves the other two green ([audit/](audit)). Each is caught by the test
it's meant for, not by the compiler: every mutant builds. `go test -count=1 ./stage3/census/latent/...` on the tree
is green ([audit/go-test-stage3-census-latent.log](audit/go-test-stage3-census-latent.log)).

What the test doesn't cover: it drives `seed`, `take` and `latentSwallow` directly with distinct pointers. It does
not run latentRun on a real program with two checkers; whether `program.Checker` hands one file a different checker
than another is not exercised on tsc either, where the meter's run decides only that the output is unchanged.

## The run (observed)

The shared-memo tree built as the meter builds it (its own `make_overlay.py`, `gofmt -w`, `go build -overlay`),
plus the timing lines, same adapted tree. The four-head tree `4a9c4665` ran right after it on the same box:

| run | wall | user s | peak RSS | loop sum wall s (user s) |
|---|--:|--:|--:|--:|
| shared memos | 14:33.0 (873.0 s) | 961.7 | 4.09 GB | 868.3 (957.0) |
| four heads `4a9c4665`, right after | 15:32.5 (932.5 s) | 1,027.4 | 4.04 GB | 927.8 (1,022.8) |
| shared / four heads | 0.94 | 0.94 | 1.01 | |

**Output: identical.** `cmp` finds no difference in `latent.jsonl` (20,131,373 bytes, sha256 `395d5157...` for both),
and the 79 stderr decision lines are identical ([diff-against-four-heads.txt](diff-against-four-heads.txt)). The
four-head run here also matches its own earlier run (15:32.2) and output.

Per file ([per-file-against-four-heads.txt](per-file-against-four-heads.txt), [shared/](shared),
[four-heads/](four-heads)): `utilities.ts` 240.1 to 225.3 s, `commandLineParser.ts` 88.5 to 79.9,
`utilitiesPublic.ts` 35.9 to 26.5, `factory/utilities.ts` 19.6 to 13.8, and the small files lose most of what they
had (`builderStatePublic.ts` 0.5 to 0.1). `types.ts` (1.06) and `program.ts` (1.10) come out slower, within the
box's noise.

A 6% gain, which is what the four-head profile allows: what this shares was already small there,
`provePredicate` 6.8% of samples and `assignedSymbol` 0.1%. What remains is the per-lowering entries walk
(39.6%), `enumNeverValue` (29.6%) and the per-statement walks over all files (19.4%), which this change leaves
alone. The box drifts by up to 12% over hours, but this pair ran back to back; a 6% difference from one pair is
near the floor of what this box can tell.
