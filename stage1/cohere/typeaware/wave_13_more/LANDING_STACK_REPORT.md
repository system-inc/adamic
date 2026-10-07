Built: rebased wave 13 onto current main 4e0bfda5, then merged lint area bb2ece56; retained runtime, developer-tools and option-bearing witness integration; no new claims.
Commits: integrated code e245b161 replaces pushed 3eb3703d; this evidence commit is pushed only to codex/typeaware-wave-13.
Commands and outputs: original PASS 333.341s; next 129/129, more 430/433 normal and ASan; corpora, compiler parity, Node, runtime, options, metadata and scoped vet pass.
Mutants: nine rule, two export-question, one released-registry, 27 metadata, three map-hash, one decoded-options and the Node one-byte comparison check caught again.
Not covered: three shared parser recovery inputs, typed supplied-node checker context, emitted-JavaScript typed-rule comparison, full repository gate and other required-input gates.

Both integration operations completed cleanly. Main brings the runtime and developer
tools stack; the lint area adds option-bearing witnesses and recovery classification
for profile snapshots. Main and the area are not linear ancestors, so the own branch
was rebased onto main and then merged the newer area commits. No main or area push
occurred. All shared changes were retained. CLAUDE.md's new counted macOS leak check
was read and retained; Linux uses sanitizers here. This unit edited no shared source.

Stage 0 and both continuation suites were rebuilt, including ASan builds. Wave 13
sources, shared parser, RuleContext and the raw checker bridge are unchanged from
3eb3703d, so the checker archives were reused. Explicitly scoped obsolete scratch
ELF binaries and C archives were removed to free space; logs, source and evidence
were preserved. Setup was reused, with previously reported total 217s and nproc 5.

TestWave13AgreementAndMutants passes in 333.341s. Controls have 19,008 identical
bytes and 32 findings; compiler 7,087 bytes/four findings and repository 19,144
bytes/one finding match Go in normal and sanitized execution. Released handles
exit 70 with the required diagnostic; its registry mutant exits zero and is caught.
Both continuation trios match Go over 77 compiler and 287 repository roots: 4,933
and 18,485 bytes, respectively, with zero findings. Positive controls supply
findings and suggestions. Next passes 129/129 controls in both modes. More matches
430/433 in both modes; every fixture ran, including the three parser refusals.
No completed native control stream has ASan, LeakSanitizer or UBSan diagnostics.

| Mutant | Required catcher |
| --- | --- |
| unassigned | Go bytes at offset 1348 |
| caught | Go bytes at offset 4146 |
| exports | Go bytes at offset 14997 |
| export-chain-flags | Go bytes at offset 14715 |
| export-module-lookup | Go bytes at offset 17159 |
| process-state | Go bytes at offset 134 |
| race-handle | Go bytes at offset 118 |
| blocking-order | Go bytes at offset 1690 |
| namespace-alias | Go bytes at offset 72 |
| literal-parentheses | Go bytes at offset 587 |
| executor-span | Go bytes at offset 101 |
| released-registry | required released-handle refusal |
| nine source listener names | production Go listener comparison |
| nine JSON listener names | descriptor comparison |
| nine numeric JSON listeners | named-kind decoding |
| map old hash | deterministic integer probe bound |
| map zero normalization | numeric hash invariant probe |
| map NaN normalization | NaN hash invariant probe |
| ignored decoded option | Node and native findings against Go |
| one-byte oracle change | byte comparison guard |

Rule and export-question mutants compile, exit zero and have empty stderr. Their
only catcher is byte comparison. Map mutants compile and must fail the specific
probe diagnostic, without sanitizer errors. The ignored-option mutant compiles
and disagrees with Go on both Node and native. The option guard was not weakened.

Selected commands, with output written directly to logs:

* TestWave13AgreementAndMutants with stage-0/archive/artifact environment variables,
  compiler and repository manifests, and -count=1 -v -timeout=30m: PASS 333.341s.
* Own validate.py, corpora.py and mutants.py for both continuation families, with
  production Go adapters and freshly rebuilt native suites: results above.
* go test ./stage1/cohere/typeaware/wave_13_more -count=1 -v: PASS 0.060s.
* go vet ./stage1/cohere/typeaware/...: PASS.
* ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave13-corpus go test
  ./stage1/typescript/parser -run '^TestCompilerExpressionsAgree$' -count=1
  -timeout=15m -v: PASS 73.714s, 77 files/28,836,875 expression-tree bytes.
* go test ./internal/native -run
  '^Test(RuntimeReleasePaths|RuntimeStringEquality|MapHashProbeBound|MapHashProbeCatchesMutants)$'
  -count=1 -v -timeout=15m: PASS 54.580s.
* go test ./stage1/cohere/lint -run '^TestDecodedOptionsAndMutant$'
  -count=1 -v -timeout=15m: PASS 138.408s. Go, Node, emitted JavaScript and native
  agree on 47 bytes before mutation.
* Filtered internal/oracle test for the one-byte guard and eight named Node fixtures
  (empty-path, runtime_last_index_of, lint_runtime_release_chain,
  lint_runtime_release_shared, lint_runtime_equal_headers, string_build_calls,
  map_foreach_named_objects, borrow_loop_calls): PASS 4.657s. Ordinary valid
  harness cache use is recorded in node.log; this is not an uncached integration gate.

Required compiler inputs were supplied and no selected test skipped. Other
required-input checks and the full repository gate were not run. All logs, canonical
JSON results and raw streams are retained under validation/landing-stack.

Single compiler whole-process observations: original native 9.920s versus Go
3.384s (2.93x); next 3.419s versus 0.550s (6.22x); more 3.621s versus 0.516s (7.02x).
Verification overlapped some measurements, so these are not a controlled benchmark
or evidence of speed gains. Native remains slower. Previous alternating measurements
remain in LANDING_PROOF_REPORT.md.

The three remaining parser refusals are 53d1c0a9ffc2fefa (JSX-like slash in .ts),
aff6ced2ca61fe89 (JSX-like element in .ts), and defcb4c4ce6a2921 (yield label and
break yield). Go exits zero; native exits 70 before rule execution. Complete inputs
remain in the frozen controls, with exact diagnostics retained. Shared RuleContext
still has no checker session or bridge-query context for these typed factories;
own suites retain per-file traversal, so listener metadata alone does not prove
supplied-node integration. Shared harness and parser changes are outside this
unit's authorized territory. The React parking exception does not apply. No new
claim was taken and full-green status is not asserted. No production Go regex
occurs in the nine rules and no regex matcher was added.
