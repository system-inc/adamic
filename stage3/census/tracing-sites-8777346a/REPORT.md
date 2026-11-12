# tracing.ts latent census at 8777346a

Base: `8777346abe4d888c0182e301b1ffec76bd45ae04`. Branch: `typescript/tracing-sites-8777346a`. This is the single-file full-mode census requested for Optimizer task #ff5eddz.

The census child exited **0** in 42.688 seconds, with **zero checker diagnostics**. The stream header says `latent_mode: full`, `checker_rejected: false`, and empty `diagnostics` and `diagnostic_sites`. The final `relation_report` is present. The overlay's fixed measurement label still says “measured on a checker-rejected program”; the actual header establishes that this run is checker-clean.

**16 distinct tracing.ts sites: 13 NotYet and 3 Refused**, comprising 18 phase-specific observations. The two method views occur in both phases. Deduplication follows the README: `(kind, where, reason, text)`; each row retains all observed phases. Attribution uses the finding's `where`, not the attempting unit: seven findings in checker.ts are not tracing.ts sites. The 21 Boundary markers are excluded from this ledger; raw.jsonl preserves the complete stream. All 23 catalogued units were attempted.

## Requested checks

- `311:34` — `recursionIdentityMap.get(recursionIdentity)`: **absent from the stream** (neither a NotYet nor a Refused finding).
- `313:38` — `recursionIdentityMap.size`: **absent from the stream** (neither a NotYet nor a Refused finding).
- `314:21` — `recursionIdentityMap.set(recursionIdentity, recursionToken)`: **absent from the stream** (neither a NotYet nor a Refused finding).
- `340:14` — `fs as { writeSync(fd: number, data: string | undefined): number }`: **Refused**, refusal_scan, lowering; exact reason: "a function taking string seen as one taking string | undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take".
- `359:10` — `fs as { writeFileSync(path: string, data: string | undefined): void }`: **Refused**, refusal_scan, lowering; exact reason: "a function taking string seen as one taking string | undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take".
- `380:1` — `declare const require: (specifier: "fs") => any;`: **Refused**, refusal_scan; exact reason: "an ambient value declaration".
- `381:1` — `declare const process: { readonly pid: number };`: **absent from the stream** (neither a NotYet nor a Refused finding).

The three map reads remain in the source; their census findings are absent. The map declaration is now at line 226 and is also absent as a site. The old line numbers are not used to infer current findings. Absence is an observation from this full-mode stream, not proof that every child expression was examined or that whole-program lowering succeeds. Full mode can stop at failed constructs, and boundary markers document those limits.

## Complete site ledger

Columns are one-based UTF-16 positions as emitted by the census. Source expressions are copied from the adapted file; the complete source line is also in sites.json. Reasons below are the exact `reason` field, without paraphrase.

| Line:column | Phase | Kind | Exact reason | Source expression |
| --- | --- | --- | --- | --- |
| 66:22 | lowering | NotYet | a call returning any | `require("fs")` |
| 98:19 | refusal_scan | NotYet | a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | `tracingEnabled` |
| 98:19 | lowering | NotYet | reading tracingEnabled | `tracingEnabled` |
| 166:9 | lowering | NotYet | assigning a field of a value | `eventStack.length--` |
| 173:9 | lowering | NotYet | assigning a field of a value | `eventStack.length = 0` |
| 185:44 | lowering | NotYet | a spread into record storage from a fixed object | `...args` |
| 196:67 | lowering | NotYet | JSON.stringify recursive array types | `args` |
| 252:46 | lowering | NotYet | reading indexedAccessType | `indexedAccessType` |
| 261:39 | lowering | NotYet | reading referenceType | `referenceType` |
| 271:43 | lowering | NotYet | reading conditionalType | `conditionalType` |
| 282:43 | lowering | NotYet | reading substitutionType | `substitutionType` |
| 291:46 | lowering | NotYet | reading reverseMappedType | `reverseMappedType` |
| 301:47 | lowering | NotYet | reading evolvingArrayType | `evolvingArrayType` |
| 340:14 | refusal_scan, lowering | Refused | a function taking string seen as one taking string \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | `fs as { writeSync(fd: number, data: string \| undefined): number }` |
| 359:10 | refusal_scan, lowering | Refused | a function taking string seen as one taking string \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | `fs as { writeFileSync(path: string, data: string \| undefined): void }` |
| 380:1 | refusal_scan | Refused | an ambient value declaration | `declare const require: (specifier: "fs") => any;` |

## Reproduction and evidence

Adapted source: `/tmp/tracing-8777346a/adapted/src/compiler/tracing.ts`; SHA-256 `e6d9a83d1df04c864a1b1c311a3270b538e13e2ba4cf228362cf993bf2034787` (15958 bytes). `tracing.ts.txt` preserves those exact bytes. TypeScript source pin: `050880ce59e30b356b686bd3144efe24f875ebc8` (v6.0.3). The worktree started at the exact base above; no compiler sources were modified. Recursive submodules and the stage3/api lockfile packages were initialized.

Commands from the pinned worktree, with `/workspace/adamic-tools/env.sh` sourced and `ADAMIC_TYPESCRIPT_SOURCE=/tmp/tracing-8777346a/typescript-source`:

```sh
bash stage3/apply.sh /tmp/tracing-8777346a/adapted
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/tracing-8777346a/overlay-final
gofmt -w /tmp/tracing-8777346a/overlay-final/*.go
go build -buildvcs=false -overlay=/tmp/tracing-8777346a/overlay-final/overlay.json -o /tmp/tracing-8777346a/latent ./stage3/census/latent/tool
python3 stage3/census/latent/audit.py /tmp/tracing-8777346a/latent
python3 /tmp/tracing-8777346a/run.py census /tmp/tracing-8777346a/latent /tmp/tracing-8777346a/adapted/src/compiler /tmp/tracing-8777346a/raw.jsonl
```

`run.py` follows `stage3/adapt/99-correctness-axolotl-s/clear-census.py`: it removes every inherited LATENT_ variable and sets only the following for the census child:

```json
{
  "LATENT_FULL": "1",
  "LATENT_ONLY_FILE": "/tmp/tracing-8777346a/adapted/src/compiler/tracing.ts",
  "LATENT_ASSERT_NO_OUTPUT": "1",
  "LATENT_TIMES": "/tmp/tracing-8777346a/unit-times.jsonl"
}
```

Every build/run wrapper writes its child's own exit code and exits with that code (signals map to 128 + signal). Each child has a 1740-second limit, followed by at most five seconds of termination grace. The successful apply took 638.587 seconds, overlay generation 7.759 seconds, build 241.012 seconds, basic audit 0.521 seconds, and census 42.688 seconds. No process was allowed to run for 30 minutes.

The README basic audit passed. An additional `full_audit.py` run exited 1: its exact expected reason set is `{a string as a condition, a number as a condition, var}`, but the actual set was `{var}`. This extra audit is **not claimed to pass**; its log and result are included. It does not alter the requested census result or its zero-diagnostic gate.

The initial apply failed because ADAMIC_TYPESCRIPT_SOURCE was unset; initial overlay attempts failed because the temporary directory and nested checker submodule were missing. Those setup inputs were supplied, and the successful build results above supersede those attempts.

Evidence includes raw.jsonl, unit-times.jsonl, the bounded wrapper, build/audit/census logs and child results, the exact adapted source, and the generated scratch overlay. Every Go evidence file is named `.go.txt`. `sites.json` retains every occurrence, phase, exact reason, and source expression. The generated overlay retains its original absolute-path mapping for provenance; its paths must be regenerated when reproducing elsewhere.
