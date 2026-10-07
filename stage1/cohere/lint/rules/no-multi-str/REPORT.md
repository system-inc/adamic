# Wave 1 slot 08, sixth batch

Ported `no-multi-str`, `no-nonoctal-decimal-escape` and `no-octal` in independent directories using `.a` modules and directory registration metadata. The decimal escape rule preserves both ordinary suggestions and all three suggestions following a null escape, with exact message IDs, replacement text and ranges. None of these rules supplies an automatic fix.

Selection inspected 341 fetched origin refs, excluding main ports and claims on every origin branch. The original 46 helper-ready rules were reserved; these are syntax inventory positions 30, 33 and 34. Claim commit `8670eaac` was pushed before implementation. See `selection.json` and the shared claim file.

## Verification

All successful logs below are fresh uncached runs against the owned scratch overlay. Production shared sources were not edited. Source Node, emitted JavaScript and native with ASan/UBSan matched Go byte for byte, including suggestions, edit ranges, replacement text and suggestion-applied source:

- Compiler plus stage1: 246 files, 13,091,844 serialized bytes; PASS.
- Owned witnesses: 29,360 bytes; PASS.
- Upstream: 74 cases (15 multiline, 31 decimal escape, 28 octal), 31,191 bytes; PASS.
- Three comparison-only mutants: PASS, each compiled and ran on all three port backends before comparison rejected it.
- Directory registry and `go vet ./...` with overlay: PASS.
- Filtered external oracle `TestTheOracleCatchesOneByte`: PASS.

The pinned TypeScript checkout is `050880ce59e30b356b686bd3144efe24f875ebc8`. Reproduce successful gates after sourcing `/workspace/adamic-tools/env.sh`:

```
python3 stage1/cohere/lint/rules/no-multi-str/validate.py --scratch /tmp/wave08-sixth-reproduce --typescript /tmp/lint-wave1-08-typescript
```

The runner prepares scratch copies of shared registry, harness and profiler using the previously owned compatibility patch. It adds `.a` loading, emitted comparison and rich suggestion serialization. This verifies candidates; default shared harness integration remains dependent on `codex/lint-harness-dot-a`.

The owned Go comparison driver deliberately runs rule listeners on recovered ASTs, as upstream rule tests do, without rejecting parse diagnostics. When no automatic fix exists it returns unchanged source instead of invoking Go's fixer, which rejects legacy grammar even for an empty edit set. Suggestions remain independently applied and compared. Rule bodies are upstream Go, unchanged. Earlier failed logs in scratch exposed an escaping error in the initial port and the empty-edit fixer limitation; those runs are not credited as passes.

## Mutants

| Rule | Mutant | Comparison result |
| --- | --- | --- |
| no-multi-str | `multiline_suppressed` disables the LF report | Missing finding caught on Node, emitted JavaScript and native |
| no-nonoctal-decimal-escape | `decimal_suggestion_wrong` changes replacement digit to zero | Identical finding, wrong suggestion edit caught on all three backends |
| no-octal | `octal_suppressed` changes leading-zero guard | Missing finding caught on all three backends |

## Throughput

Best of five rotating process measurements, 78 files (77 compiler files plus 1,000 positive findings), release native, source Node and Go. Every run checked count equality. These are wall-clock process rates, including startup and parsing, on this shared cloud machine.

| Rule | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: |
| no-multi-str | 624.76 | 821.94 | 3760.85 |
| no-nonoctal-decimal-escape | 627.78 | 832.69 | 3906.92 |
| no-octal | 679.36 | 886.31 | 4195.79 |

`bash cloud/setup.sh`: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 32s, total 32s. `nproc`: 5. Go 1.27.1, clang 20.1.8, Node 24.19.0.

## Exact gaps

Full upstream probe selected 83 cases and failed on shared JSX parsing: expected GreaterThanToken, got Identifier at position 13. Six multiline JSX fixtures are excluded from the successful 74-case comparison. Three octal recovery fixtures independently fail in the shared parser before listeners: `var a = 01.5;` expects a semicolon at 10; `var a = 0777.5;` and `var a = 0755n;` expect one at 12. See gap and recovery logs. Those nine cases are not certified on any port backend. No shared parser or harness file was changed.

Reproduce the expected failures after preparing the overlay:

```
go test -overlay=/tmp/wave08-sixth-reproduce/overlay.json ./stage1/cohere/lint -run '^(TestSixthFull|TestSixthRecovery)$' -count=1 -v -timeout=10m > /tmp/wave08-sixth-gaps.log 2>&1
```

The full repository test gate was not run. The owned candidates, registry, whole-repository vet and filtered external oracle were run. Logs are under `evidence/`.
