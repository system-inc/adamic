Built: rebased wave 23 onto area d65a8f931; added react/jsx-fragments, bringing completed native rules to 17; four analysis rules are parked.
Commits: rebase foundation b11889cc451d576e7a3282dedc6bcaab2d361cf6; this report accompanies the final rule/evidence commit on codex/typeaware-wave-23.
Commands and outputs: seven existing gates PASS 645.160s; fragments PASS 52.887s; bridge PASS 74.552s/checker 0.123s; filtered Node oracle PASS 6.906s; go vet clean.
Mutants: all existing rule/output/listener/handle mutants were caught again; fragments compiled and exited normally but differed at byte 2172; malformed-question mutant failed its contract test; released fragment handle exited 70.
Not covered: full repository gate and native context-value/capture analysis; no unclaimed ranking entries remain in the fetched origin snapshot.

The branch includes origin/main 39638d9e278d38bb5aeae887f46d55a70e47aaad and origin/area/stage1-lint d65a8f931c98655936ae04c6899f38f14862b73e. Shared harness, runtime optimizations and allocator checks from integration are retained. No shared generator or test harness was edited. The only shared bridge edit is one registration line in facts.go; the new jsx-declaration-syntax question and its native decoder live in their own files.

react/jsx-fragments listens to JsxElement, JsxSelfClosingElement and JsxFragment, declared by name in its owned rule.json, with node:true. Its probe dispatches by those kinds and passes the existing node. Declaration syntax facts distinguish React imports, variable initializers and destructuring. Both default syntax mode and element mode match production Go findings, fixes and suggestions. Go deliberately supplies no fixes or suggestions for this rule. The comparison covers 18 generated controls, 77 frozen compiler roots and 287 frozen repository roots; sanitizer executions agree as well. The independent Go adapter imports the production rule, not the native implementation.

A negative control exposed a bug in the first native initializer helper: searching for any variable declaration picked a declaration inside a function or class initializer. The saved fragments-negative.log records the byte mismatch. The corrected helper follows the outer variable statement's declaration directly. The final gate includes both negative witnesses and passes. An earlier optional-chain compiler limitation was handled with explicit traversal in the owned helper, without compiler changes.

Native versus Go single-run timings for JSX fragments: compiler 1.263560609s versus 0.316406022s; repository 0.249305543s versus 0.165852169s. Native remains slower. Existing rule timings and every mutant outcome are preserved in rules.log.gz. The new rule mutant changes the accepted require module from react to preact; it compiles, exits zero with empty stderr, and only the independent comparison catches it at byte 2172. The released-handle probe exits 70 with the required invalid-or-released checker panic. The bridge guard mutant accepts a malformed question and is caught by TestJsxDeclarationSyntax.

react/jsx-no-constructed-context-values is PARKED, alongside the three previously parked hooks rules. Reading its companion stability module showed that memo dependency decisions require resolved callee return evaluation and holder identity, capture and escape analysis. The initial AST/checker-only assessment was incomplete. The committed Go witness adapter reports the fresh-return dependency, but reports neither a cached return nor a closure-captured return. Those outputs are under context-capture in the evidence directory. Native equivalents are blocked on the analysis work identified by #dnv6f2c. No native parity is claimed for this parked rule; per the user's parking instruction it counts as finished for landing-first.

The selection snapshot scanned 609 origin refs and 33 distinct claim blobs, finding 172 ranked claimed names. Together with the original 26 ported rules, these cover all 197 ranking entries (one overlaps). Remaining candidates: zero. No additional rules were claimed.

Reproduction uses source /workspace/adamic-tools/env.sh and the frozen manifests /workspace/wave-23/compiler.manifest and repository.manifest, with ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-23/typescript. Existing gate command:

    go test ./stage1/cohere/typeaware -run '^TestWave23((Next|Core|Constructor|Behavior)?AgreementAndMutants|NumericListeners|JsxNoUndefAgreementAndMutants)$' -count=1 -timeout=30m -v

New rule command:

    go test ./stage1/cohere/typeaware -run '^TestWave23JsxFragmentsAgreementAndMutants$' -count=1 -timeout=15m -v

Bridge command (ADAMIC_TSGO_CORPUS=typescript):

    go test ./bridge/tsgo/... -count=1 -timeout=15m -v

Node command:

    go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|functions|closures|sorting|call_targets_.*|devirtualize|047cb0d_n_.*|library_map_set_iterator_.*|override_same_representation|inherited_static_field_read|runtime_last_index_of)\.a$|^TestRuntimeLastIndex' -count=1 -timeout=10m -v

Each test writes to its own log, without a pipe. The initial Node attempt exhausted scratch disk space before compilation; node.log preserves the failure. Removing only generated ELF binaries and archives from completed scratch gates allowed node-retry.log to pass. Source, diagnostics and logs were preserved. go vet ./... and git diff --check pass. The toolchain setup was not repeated: prior timing lines were Go/clang/Node/submodules 0s, warm cache 83s, total 83s; nproc 5, quota 4. Raw evidence is compressed with stable gzip timestamps and hashes of the uncompressed bytes in validation-wave23-d65a/raw-sha256.json.
