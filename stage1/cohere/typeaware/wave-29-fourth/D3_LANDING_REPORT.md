Rebased 31 owned commits onto area d3a37422c, containing main b6b1538b0, without conflicts.
Rebased source c58cf2e3ad91b31254b6eef218d1f1f20232ed9d; predecessor 6574d02b3232fb236f9507172fd555fbb037a844.
Original oracle PASS 107.529s; next configured/corpus sanitizer comparisons and all available partial kernels PASS; checker PASS 0.171s.
All existing rule, provenance, release, graph, listener and resolution mutants caught; inherited typeof mutant oracle PASS 4.792s.
Full source/checker adapters, React analysis and configured regex remain blocked; full gate and required external-input checks not run; no new claims.

Fetched all origin heads explicitly, then rebased onto the integrated area tip.
No shared implementation edits or guard changes were made. Main's typeof null,
constructor, string-literal and lookup-presence changes warranted fresh compiler
and native kernel comparisons. The source setup toolchain was reused; setup
was not repeated (prior run 103s, nproc 5).

Commands, all redirected to the compressed logs in validation-d3:

- git fetch origin '+refs/heads/*:refs/remotes/origin/*'
- git rebase --onto origin/area/stage1-lint b46914832d70e00847d82d5d221ab7bb24040c53
- source /workspace/adamic-tools/env.sh; go run ./cmd/lint-registry; go build -o /workspace/wave29-area-adamic ./cmd/adamic
- ADAMIC_WAVE29_ARTIFACTS=/workspace/wave29-area-default ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave29-typescript go test ./stage1/cohere/typeaware -run '^TestWave29AgreementAndMutants$' -count=1 -timeout=30m -v
- ADAMIC_COMPILER=/workspace/wave29-area-adamic python3 -u stage1/cohere/typeaware/wave-29-next/check.py /workspace/wave29-area-next
- With that compiler: wave-29-fourth/check.py; wave-29-third/check_static_core.py, check_control_core.py, check_listeners.py; wave-29-configured/check_regex_contract.py; wave-29-fourth/rules/react-jsx-no-undef/testdata/check_resolution.py. Each uses the corresponding /workspace/wave29-area-* artifact directory documented in the preceding report.
- go test ./bridge/tsgo/checker -count=1 -timeout=30m: PASS 0.171s
- go test ./internal/oracle -run '^TestTypeOf(NullMutant|NullSlotPresenceMutant|ConstructorMutant|StringLiteralMutant)$' -count=1 -timeout=30m -v: PASS 4.792s. Node comparisons catch restored null-as-undefined behavior, missing-vs-present null slots, constructor-as-object and string-literal-as-undefined mutants. Cache observations: native hits 5/misses 5, Node hits 1/misses 7, probe hits 0/misses 0.

Original controls: 9840 bytes/15 findings. Corpus: 77 compiler files/5241 bytes,
287 repository files/18485 bytes, zero findings, identical sanitizer streams.
Denylist, match, concurrency and provenance compiling mutants caught only by
bytes; released-handle mutant caught by required panic. Native/Go compiler
2.080769432s/0.341989733s (6.08x); repository 0.303722981s/0.162081885s (1.87x).

Next controls: 400 cases/252 findings/seven option profiles; the same two corpora
match. Globals, setter and shadow compiling mutants caught; ASan/UBSan/leaks
match every configured/corpus stream. Native/Go compiler
1.888233789s/0.331850159s (5.69x); repository 0.283996748s/0.151979029s (1.87x).
These concurrent process timings remain observations, not isolated benchmarks.

Partial kernels: JSX 79 rows/7359 bytes; static 34 graphs/27 findings; control
17796 graphs/28 sources; nine named listener declarations/562 bytes; resolution
148 controls/28 findings/18408 bytes. Existing mutants caught, including seven
resolution mutations; their logs identify each mutation and comparison.
No test-only raw-fact kernel is certified as a production source adapter.
Regex contract test still catches the constant-pattern mutant on Node; dynamic
native RegExp refusal and five Go/JS dialect differences remain blockers.
The unchanged shared options guard was not modified; its 43.430s green mutant
proof on b469 remains recorded in B469_LANDING_REPORT.md and was not rerun here.

Availability scan after the all-heads fetch: 657 origin refs, 33 unique claim
blobs, 197 ranked checker-dependent rules, zero unclaimed/unported candidates.
Full gate and the 17 required external-input checks were not run in this focused
worker gate; no skip was counted as green or any check relaxed or removed.
