Built: `.a` port of `nexus/correctness-no-uncleared-race-timeout`, plus partial native process-output state; two process rules remain incomplete.
Commits: claim `952a900c`; implementation `80e56233`; all code and evidence pushed on `codex/typeaware-wave-30`.
Commands and outputs: timer gate PASS 41.805s, state gate PASS 13.775s, vet PASS; setup 21s, nproc 5.
Mutants: timer range caught at byte 68; catch-state caught at byte 5; both exited 0; sanitizers and released handles passed.
Not covered: complete process-exit and blocking-streams ports, their diagnostic mutants, the full repository gate or the full oracle matrix.

The timer port uses the existing symbol-lineage, symbol-identities and binding-declarations
facts. All timer/race/handle decisions execute in native Adamic. It adds no bridge
question and changes no shared registration or harness file. The private suite and
oracle are new files. The oracle calls the pinned Go cohere production registry and
serializes every finding, fix and suggestion. This rule supplies zero fixes and zero
suggestions; both counts are compared.

`TestWave30ThirdAgreementAndMutants` compares 20 authored `.a` controls under DOM
and Node ambient declarations. DOM: 15 findings, 9513 identical bytes. Node's merged
`setTimeout` namespace: 14 findings, 8965 identical bytes. Both sanitizer runs agree.
The 77-file TypeScript compiler corpus agrees on 5318 bytes and zero findings; the
287-file repository corpus agrees on 18485 bytes and zero findings. Both corpus
sanitizer runs agree. These are the existing corpus manifests, not an exhaustive
scan of all repository sources. Unicode, parentheses, const versus let aliases,
shadowing, nested functions/classes, shorthand reads, dropped/assigned handles,
DOM receiver methods and Node global augmentation are covered.

The timer-range mutant changes the reported end by one byte. Its native executable
exits 0 with empty stderr, and the independent Go byte comparison catches byte 68.
Released handles for all three existing questions refuse with the required panic
and exit 70. The initial implementation's missing-node sentinel panic was corrected
before the final passing gate.

Commands (source `/workspace/adamic-tools/env.sh` first):

- `bash cloud/setup.sh`: Go 0s, clang 0s, Node 0s, submodules 0s, cache warm 21s, total 21s; nproc 5, quota 4 cores.
- `go test ./stage1/cohere/typeaware -run '^TestWave30ThirdAgreementAndMutants$' -count=1 -v`, with `ADAMIC_WAVE_30_THIRD_ARTIFACTS=/workspace/wave-30-third-validation-reviewed`, both `ADAMIC_WAVE_30_THIRD_*_MANIFEST` variables and `ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-30-typescript`: PASS 41.805s.
- `go test ./stage1/cohere/typeaware -run '^TestWave30ProcessOutputStateAgreement$' -count=1 -v`: PASS 13.775s.
- `go vet ./stage1/cohere/typeaware` and `git diff --check`: PASS.
- `python3 bridge/tsgo/profile/volume_bench.py <native> <oracle> <directory> --rounds 3 --corpus compiler <config> <manifest> --corpus repository <config> <manifest>`: all counts agree.

Three quiet alternating benchmark rounds give median whole-process native/Go times:

| Corpus | Native | Go cohere | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 1.629154s | 0.293601s | 5.55 |
| Repository | 0.232594s | 0.123679s | 1.88 |

Native parsing/traversal dominates these zero-finding corpora; bridge timing reports
zero queries there. These measurements establish no speed-parity claim.

The process-output state is a partial component, not either complete process rule.
Its minimal catch-chain set, subset reduction, catch filtering and canonical key
agree with the unchanged production Go state implementation on 216 transitions,
including sanitizer runs. A reversed catch filter exits 0 and differs at byte 5.
The state mutant does not substitute for a per-rule diagnostic mutant.

The remaining work is concrete:

- `correctness-no-process-exit-after-output` requires native control-flow roots,
  successor blocks and expression/statement events, including exception routing,
  catch binding evaluation, finally replay, loops and optional chains. It also
  requires raw ancestor names to distinguish `NodeJS.Process`, and a resolved-call
  declaration/body fact to follow the exact checker-selected callee, respecting
  async/generator/never-return refusals. Existing `signature-shape` exports types,
  not that declaration. The reusable native state is ready, but no diagnostic
  implementation is asserted.
- `correctness-require-blocking-standard-streams` additionally requires the complete
  program's source files and resolved static/dynamic import/require targets. The
  present bridge has no complete module-resolution fact interface. Entry-file
  selection, transitive stream reachability, import-load blocking and ordered
  callback/callee analysis cannot be established from per-file symbols alone.

These are missing analysis/bridge components, not a reported `.a` loader or
suggestion-serialization harness failure. Both incomplete rules retain their claims;
no later batch was claimed. No silent no-op rule or approximate findings implementation
was registered. No shared harness, registration generator, emitter, lowering or
native compiler file was changed.

Evidence is in `validation-wave-30-third`: fetched origin head and claim blob
snapshot, corrected token-boundary ranking selection, compressed exact byte streams
and authored controls, hashes, gate logs and timing records. The selection snapshot
reconstructs the pre-claim branch head `b64a21b3`; selection still yields exactly the
three rules in the pushed claim. Pinned cohere's standalone CLI `.a` limitation remains
as reported in the previous batch; the independent production-registry loader and
native suite both support the authored `.a` controls.
