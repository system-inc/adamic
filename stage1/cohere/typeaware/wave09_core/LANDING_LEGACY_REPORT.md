Built: rebased wave-09 onto lint area b46914832, preserving its legacy-rule registry migration and current main c7991b900.
Commits: old remote 1360c061a; rebased implementation c93e98589; evidence commit follows on wave-09 only.
Commands/output: original rules with frozen corpora, seven owned rule/contract verifiers, registry and vet PASS.
Mutants: fresh original semantic, scope, handed-node, suggestion, listener and released-handle checks all detected their mutations.
Not covered: complete regex rules or full gate; unchanged bridge, parser and pure-helper checks retain the preceding landing evidence.

Area advanced to b46914832d70e00847d82d5d221ab7bb24040c53, merging
lint-rules/legacy at 91879f86. Current main remains c7991b900. Rebase
completed without conflicts. The compiler was rebuilt before verification.
Shared changes were retained and no shared files were edited. The area diff
changes only stage1/cohere/lint sources; internal, bridge, stage1/typescript
and stage1/cohere/typeaware sources are unchanged. Accordingly their clean
required compiler/bridge/pure-component results in LANDING_C799_REPORT.md
are retained, not described as newly rerun.

After sourcing /workspace/adamic-tools/env.sh:

```
git rebase origin/area/stage1-lint
go build -o /workspace/wave-09-core/adamic ./cmd/adamic
ADAMIC_TYPESCRIPT_SOURCE=/workspace/TypeScript-050880ce59e30b356b686bd3144efe24f875ebc8 ADAMIC_WAVE09_REPOSITORY_MANIFEST=/workspace/wave-09-validation/repository.manifest ADAMIC_WAVE09_COMPILER_MANIFEST=/workspace/wave-09-validation/compiler.manifest ADAMIC_WAVE09_ARTIFACTS=/workspace/wave-09-legacy-original-artifacts go test ./stage1/cohere/typeaware -run '^TestWave09' -count=1 -timeout=30m -v
go test ./stage1/cohere/lint/registry -count=1
go vet ./stage1/cohere/typeaware ./stage1/cohere/lint/registry
```

The seven Python verifiers and their log paths are enumerated in
validation-landing-legacy/checks.json: label scope, label handed-node,
flags frontend, flags handed-node, complete literal slice, listener metadata
and dynamic RegExp blocker. Each command writes its own log, never a pipe.
Original package PASS 162.512 seconds, registry PASS 0.074 seconds, vet
output empty. Original controls 44 findings and compiler/repository 14/4
findings match full Go finding/fix/suggestion bytes in normal and sanitized
native runs. Compiler/repository sources remain the frozen 77/287 manifests.

Fresh clean-running mutants differ from Go at byte 64 (prefer-const), 10017
(radix), 12914 (type-parameter flag), 12314 (non-nullable assertion fix).
Released-registry retention removes required panic 70 and is caught. Label
membership, scope value-mask, handle retention and handed-node refetch are
caught by their independent Go/refusal oracles. Label controls retain 18
findings/7700 bytes, both corpora zero/7859 and 18485 bytes. One parser-
refused label fixture remains explicitly recorded. Flags controls retain 35
findings/6020 bytes; report-refetch differs at byte 58. Literal slice retains
22 default and 17 allow-escape findings with suggestions, both corpora zero,
normal/sanitized native and source/emitted Node. Its suggestion-range mutant
is caught. Named listener checks retain their byte-276 wrong-kind catcher.
All fresh logs retain precise mutant differences.

Whole-process native/Go time: compiler 7.258040/1.248913 seconds;
repository 0.578165/0.398701 seconds. These concurrent runs do not establish
isolated throughput. Setup remains the reused successful 88-second run,
nproc 5. Unrelated required checks and the full gate were not run; no skip,
input pin or correctness check was weakened.

Dynamic new RegExp(pattern, 'u') remains refused. Source Node passes and
the static-pattern mutant builds/runs under sanitizers, proving the refusal
assertion can fail. Shared regex runtime tip ca71f1deb still explicitly
reports unfinished lowering. Both regex claims remain incomplete, shared
checker installation and remaining visitor migration are unfinished, and no
React parking exception applies. No new matcher, parser, rule claim or
shared registration edit. Publication uses an exact lease on old remote
1360c061a3a2c11b5af4ba13091297c5c257166f exclusively to
codex/typeaware-wave-09, never main or any area branch.
