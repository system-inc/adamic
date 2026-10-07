Built: replaced the owned inline-adjacency edge-whitespace matcher with one JS RegExp literal using the pinned Go Unicode space class.
Commit: based on pushed 341bd4de, still on current main f8013f0b; this continuation is pushed only to codex/typeaware-wave-06.
Checks: 61 prepared-node controls, 47 findings, 22093 bytes match Go, native, ASan/UBSan, source Node and emitted JavaScript; Python syntax and whitespace checks pass.
Mutants: the three rule mutants remain caught only by Go bytes; removing U+0085 from the regex also exits zero and loses independent byte equality.
Not covered: shared numeric JSX parser/driver/raw-fact adapter, source corpus parity and end-to-end timings remain blocked; no new claims.

Main and the own branch were checked after a fresh all-heads fetch. Main has not advanced beyond f8013f0b; there is no developer-tool leak-helper change to reconcile in this snapshot. No codex/lint-regex branch was present in the fetched heads, so no shared translation row was available.

The Go production helper uses unicode.IsSpace rather than a Go regexp. Its documented edge-whitespace predicate was previously implemented by hand in the prepared-node kernel. Under the new no-hand-rolled-matchers instruction, it is now one RegExp literal, translated once into an explicit Unicode class. Go IsSpace accepts U+0085 and excludes BOM, unlike JS \s, so blindly using /(?:^\s|\s$)/u would change findings. Both cases are exercised through unmodified Go source controls. No options pattern is involved. The predicate builds no finding position, so no byte/UTF-16 offset conversion belongs in it; supplied byte spans remain unchanged.

The byte total includes absolute fixture paths; the shorter scratch path explains the different total from the prior report. All 61 controls still produce 47 findings. The native kernels consume supplied syntax facts and contain numeric kind comparisons. Source analysis still refuses; the shared parser and area/stage1-lint snapshot still expose string kinds and no usable JSX node adapter. Existing React HIR claims remain parked on the previously named source-analysis blockers.

Three semantic rule mutants and the regex-class mutant compile and exit zero with empty stderr. Only full independent Go output bytes catch them. Source Node and emitted JavaScript also match normal output. Sanitizer stderr is empty. The three native source-refusal probes still exit 70. This adds no checker question or runtime handle, and does not claim a new released-handle test. Previous bridge ownership evidence remains applicable on unchanged main.

```sh
source /workspace/adamic-tools/env.sh
TMPDIR=/workspace python3 stage1/cohere/typeaware/wave_06_jsx/validate.py --scratch /workspace/wave-06-jsx-regex > /tmp/wave-06-jsx-regex.log 2>&1
```

All test output goes directly to files. Complete fixtures, generated source/emitted JavaScript, commands, timings and outputs are retained with hashes in regex_evidence/index.json. The original source-port and option/corpus limits remain in REPORT.md. No shared file, protected compiler file, submodule pin, main branch or area branch is changed.
