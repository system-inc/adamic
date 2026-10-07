Rebased all 30 owned commits onto area b46914832, including main c7991b900, without conflicts.
Rebased source c949027e9896719e1f712bc438abef3a07bc492e; published predecessor 32268860dd6c275f9a2f9c049daa5e59f81d76f2.
Original oracle PASS 110.678s; next seven option profiles and both corpora PASS; bridge/registry/options and all partial kernels PASS.
All existing rule, provenance, released-handle, graph, listener, resolution and option mutants caught; sanitizer streams match.
Full source adapters, React analysis and configured regex remain blocked; full gate and external required-input checks not run; no new claims.

Correction to earlier refresh reports: the repository's fetch configuration only
refreshed wave-29 with plain git fetch origin. An explicit all-heads refspec
fetched area b46914832 and its main c7991b900 integration, resolving the prior
rebase blocker without editing shared files. Future checks must fetch all heads.

Commands (each stdout/stderr redirected to preserved compressed logs):

- git fetch origin '+refs/heads/*:refs/remotes/origin/*'
- git rebase --onto origin/area/stage1-lint d65a8f931c98655936ae04c6899f38f14862b73e
- source /workspace/adamic-tools/env.sh; go run ./cmd/lint-registry; go build -o /workspace/wave29-area-adamic ./cmd/adamic
- ADAMIC_WAVE29_ARTIFACTS=/workspace/wave29-area-default ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave29-typescript go test ./stage1/cohere/typeaware -run '^TestWave29AgreementAndMutants$' -count=1 -timeout=30m -v
- ADAMIC_COMPILER=/workspace/wave29-area-adamic python3 -u stage1/cohere/typeaware/wave-29-next/check.py /workspace/wave29-area-next
- With the same compiler: wave-29-fourth/check.py; wave-29-third/check_static_core.py, check_control_core.py, check_listeners.py; wave-29-configured/check_regex_contract.py; wave-29-fourth/rules/react-jsx-no-undef/testdata/check_resolution.py; wave-29-fourth/prove_gaps.py; wave-29-third/prove_gaps.py. Artifact directories reuse the corresponding /workspace/wave29-area-* paths from the previous area run.
- go test ./bridge/tsgo/checker -count=1 -timeout=30m: PASS 0.152s
- go test ./stage1/cohere/lint/registry -count=1 -timeout=30m: PASS 0.087s
- go test ./stage1/cohere/lint -run '^TestDecodedOptionsAndMutant$' -count=1 -timeout=30m -v: PASS 43.430s; 47 bytes match Go/Node/emitted JavaScript/native; ignored decoded-option mutant caught on Node and native. No guard changed or bypassed.

Original control bytes 9840/15 findings; corpus 77 compiler files/5241 bytes and
287 repository files/18485 bytes match, including sanitizer streams. Denylist,
match, concurrency and provenance mutants compile and finish normally and are
caught by bytes; released-handle mutant caught by required panic. Native/Go
compiler 1.849713894s/0.387566663s (4.77x); repository
0.271564127s/0.142197805s (1.91x).

Next controls 400/252 findings/seven groups; three compiling mutants caught.
All configured/corpus ASan/UBSan/leak streams match; combined SHA-256
 a7d2e5c0ea0e5c53b2e5831eccf1d13adb03b273d3894272f3a48164bd89e254.
Native/Go compiler 1.968103373s/0.362211364s (5.43x); repository
0.293461015s/0.153117691s (1.92x). Concurrent process timing remains an
observation, not an isolated benchmark or proof of shared driver performance.

All partial kernels re-green: JSX 79 rows/7359 bytes; static 34 graphs/27
findings; control 17796 graphs/28 sources; nine named declarations/562 bytes;
resolution 148 controls/28 findings/18408 bytes. Existing compiling mutants
caught in each suite, including seven resolution mutations. Source probes
confirm parser availability but no native source-to-SSA React analysis entry.
Regex Node factory and constant-pattern mutant hold; nonconstant native RegExp
and five Go/JS dialect witnesses still block configured id-match. No full
checker-backed source port is claimed by the raw-facts test kernels.

Setup was not rerun; existing sourced toolchain was reused. Prior setup was
103s and nproc 5. Full gate, the 17 external-input correctness checks, and broad
new-main compiler proof/record tests were not run in this focused worker gate.
No skipped test was counted as passing, relaxed or removed.
