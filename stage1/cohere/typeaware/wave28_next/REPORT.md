Built: Native .a ports of no-process-exit-after-output, no-uncleared-race-timeout, and require-blocking-standard-streams, plus raw bridge questions.
Commits: original ports 24c600f0; continuation claim d45c4d67 was pushed before implementation; implementation is the commit containing this report.
Commands/results: continuation PASS 141.812s; bridge PASS 135.090s; checker PASS 0.177s; filtered Node oracle PASS 44.361s; vet, formatting and whitespace pass.
Mutants: three rule mutants exit 0 with empty stderr and fail only the Go byte comparison; released-registry mutant, seven bridge mutants and Node's byte mutant are caught.
Not covered: full repository gate, complete production test-project matrix, shared registration/profile integration, emitted-JavaScript rule comparison, or speed parity.

The original wave's three rules were tested and pushed before continuation. Fetching all heads found 325 origin refs and 33 distinct Markdown claim blobs naming 96 ranked rules. Excluding those claims and ports on main and origin/codex/tsgo-c-library left these first three by volume. All have zero volume in both frozen corpora. No selected rule was skipped. The claim is in ../claims/wave-28.md, committed at d45c4d672a5904f9162e29bea5849e2ff22518ee before any continuation implementation.

Each rule has its own .a file. The directory also owns its AST helpers, control-flow graph, runner, Go test harness and independent production-Go oracle. The shared registration generator, shared test harness, parser and protected compiler files were not edited. The only existing shared file edit is one dispatch line in bridge/tsgo/checker/facts.go. New stream-symbol, stream-signature, stream-program and stream-file questions each have a named Go and Adamic file. They return symbol declarations and ancestor shapes, resolved signature attributes, program source/import facts and source-file mode. All lint predicates, graph construction, flow state, call following and import closure execute in native Adamic; the bridge returns no lint verdict, message, fix or suggestion. The owned runner imports and invokes the rules without changing shared registrations.

The Go oracle invokes the unmodified pinned production rules through their registry with an independent loader and walk. It imports no bridge. Both runners serialize complete finding spans, rule names, IDs, messages, fixes and suggestions, sorted identically. These three production rules have no repair suggestions or fixes, and those zero-length fields are compared too. The comparison includes explicit .a roots through AllowNonTsExtensions.

| Population | Roots | Findings | Identical output bytes | Native sanitizer stderr bytes |
| --- | ---: | ---: | ---: | ---: |
| Expanded production-source controls | 132 | 108 | 67,395 | 0 |
| Isolated timer and flow controls | 18 | 14 | 8,672 | 0 |
| Frozen repository | 287 | 0 | 18,485 | 0 |
| TypeScript src/compiler | 77 | 0 | 5,010 | 0 |

All 132 generated expanded sources passed the Go syntax filter; none was excluded. They are mechanically extracted literal sources from the pinned production tests plus additional controls. They are run together, so their ambient augmentations can affect one another. In particular, production sources that augment the global timer make Go conservatively decline timer findings elsewhere in that program. The separate 18-root program has six positive exit-after-output, six timeout, and two blocking-standard-streams findings. It includes shadowed APIs, unread and retained handles, one-level writer calls, and a started async function with an await before output and exit. Expanded controls additionally cover loops, switch, try/catch/finally, imported process aliases, destructuring, callbacks, generator functions, decorators, and computed imports. These extracted snippets do not reproduce every multi-file helper/project setup in the production test suites. Both populations and both corpora match in normal and ASan/UBSan/LSan builds.

The comparison caught incorrect optional-node handling in my code. A top-level await control first exposed the shared parser's unset module await context; the owned runner now obtains raw external-module mode and sets that existing parser context. No parser edit was required. The compiler corpus then caught my raw symbol serializer calling Node.Text on a destructuring binding pattern. The new question file now extracts textual names only from textual name kinds; binding patterns remain represented by their ancestor kind and source span. A direct checker regression control covers that case. The final run passes these cases and all 77 compiler roots.

| Mutant run | Catch |
| --- | --- |
| Report every reachable exit regardless of preceding writes | Exit 0, empty stderr; Go comparison differs at byte 1,837 |
| Reverse timer identifier read/write classification | Exit 0, empty stderr; Go comparison differs at byte 4,886 |
| Treat every file as requiring ordered blocking analysis | Exit 0, empty stderr; Go comparison differs at byte 8,035 |
| Keep released checker handle in registry | Mutant exits 0; required panic 70 is absent |
| C ABI input or output length +1 | ASan heap-buffer-overflow |
| Bridge released handle retained | Stale-handle assertion |
| Bridge type taken from SourceFile | Independent Go mismatch at byte 6 |
| Remove bridge link opt-in guard | Refusal test |
| Remove C output free | LeakSanitizer |
| Allocate region entry result on heap | LeakSanitizer |
| Change one Node oracle output byte | Differential oracle mismatch |

The normal and sanitized stream-symbol released-handle probes both exit 70 with exactly `adamic: panic: invalid or released checker handle`. The checker package verifies question framing, source mode, imported declaration origins, module resolution, destructuring, wrong-node refusals and unsupported questions. The bridge gate also verifies 100 C ABI queries, outputs surviving release, stale/zero-handle rejection and 162 independent positions producing 3,261 identical bytes under sanitizers. The filtered Node oracle runs eight fixtures through Node, JavaScript backend and sanitized native, plus its byte mutant.

Three fresh-process rounds compare timed stdout again. These are wall-clock observations with warm filesystem caches on this worker; they include loading and serialization, and run Go then native each round.

| Population | Median native seconds | Median Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 0.411181 | 0.169295 | 2.43 |
| Compiler | 4.072790 | 0.585792 | 6.95 |

Native is slower. Every round is retained in validation/timing.json. Complete canonical streams, their SHA-256 hashes and stderr lengths are in validation/. Absolute file headers affect hashes on another machine. Portable manifests and root-content hashes are copied from the original wave's frozen validation. TypeScript is 050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3); cohere is 715ba94f3608a6500086b1076ce5cb7e51b836db; typescript-go is 8d550c837c90bd1805b047b7eeccc2baac2d5e7a.

Setup was completed for this wave: Go 1.27.1 ready 0s; clang 20.1.8 ready 1s; Node v24.19.0 ready 1s; submodules 1s; build cache warm 114s; done 114s. nproc is 5, CPU quota 4, memory 17.6 GB. See ../WAVE_28_REPORT.md and its setup evidence. No second setup was needed for continuation.

Commands from the repository root, with each test's output written to a file:

```bash
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE28_NEXT_ARTIFACTS=/workspace/wave-28-next-validation \
  go test ./stage1/cohere/typeaware/wave28_next -count=1 -timeout=25m -v \
  > /tmp/wave-28-next-final.log 2>&1
go test ./bridge/tsgo -count=1 -timeout=15m -v > /tmp/wave-28-next-bridge.log 2>&1
go test ./bridge/tsgo/checker -count=1 > /tmp/wave-28-next-checker-final.log 2>&1
go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  -count=1 -timeout=10m -v > /tmp/wave-28-next-node.log 2>&1
go vet ./stage1/cohere/typeaware/wave28_next ./bridge/tsgo/... > /tmp/wave-28-next-final-vet.log 2>&1
```

The continuation harness uses /workspace/wave-28-artifacts/{repository,compiler}.manifest and /workspace/wave-28-corpus/src/compiler/tsconfig.json when present. Materialize the portable manifests in validation/ at those locations to reproduce this complete run. It explicitly logs absent external corpora on other machines. Full go test ./... and full vet were not run; the touched packages and filtered oracle above were used. The shared generator and emitted-JavaScript rule harness are left for codex/lint-harness-dot-a integration. The pinned ordinary cohere CLI rejects .a inputs, already demonstrated in the original wave's lint-refusal evidence; this is not a clean production lint claim. None of those integration gaps blocked native .a compilation or the independent full diagnostic comparisons here.
