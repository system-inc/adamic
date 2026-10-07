Migrated exhaustive-deps effect classification to the shared JS RegExp literal.
Tested code c8b74e6c2be86dbe6fd218ad60103c7a14692d6f; branch remains on current main c01907a7.
605 controls/437 findings and both corpora match Go bytes, normal and sanitized.
Four rule mutants exit 0; Go bytes catch them; released-handle checks pass.
Runtime additionalHooks patterns remain blocked; shared harness is not on main.

# Regex instruction follow-up

All twelve default-option ports were already green and pushed at 72c882862 on
c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. Two fresh main fetches confirm the base
is unchanged, so no new landing rebase is necessary. The only changed rule is
exhaustive-deps; its complete owned batch was revalidated. No new claims were made.

The actual Go regexp sites among the twelve rules are the exhaustive-deps fixed
pattern at react/exhaustive_deps.go:451 and additionalHooks at line 463. The fixed
translation comes from origin/codex/lint-regex, tip
071fb012848ce0408428c61aba0857cca472236f, and is preserved in regex-rows.json:
`Effect($|[^a-z])` becomes `/Effect((?![\s\S])|[^a-z])/u`.
The shared table's g flag is for enumerating matches; Go MatchString is stateless,
so the rule's boolean test uses u without g. The rule performs no translation at
runtime and uses no hand-written matcher. The regex result is only a boolean;
there is no match offset to convert before building a finding.

The required dynamic constructor is written in additional_hooks_pattern.a as
`new RegExp(pattern, 'u')`. It is deliberately isolated from the default suite:
importing it in a real runtime-pattern probe causes native compilation to exit 1
with `stage 0 can't lower RegExp with a nonconstant pattern yet`. The corresponding
Node constructor probe exits 0 and prints true. This is an observed lowering
blocker, not a claim that arbitrary Go option dialects already equal JS Unicode
RegExp. No matcher fallback is supplied, and additionalHooks option integration
remains unfinished. Fixing compiler lowering is outside this rule directory.

The current shared harness tip is exactly ab70f38d47de1d4974082b38f84a56af2368b7af.
Its seven-argument report and wire protocol remain, beside reportNode/reportRange
and the optional node metadata flag. Its registration directory is absent from
current origin/main. This unit does not import an unlanded area branch or edit
shared registration/test files. The announced leak-helper changes are likewise
accepted when they arrive through main; no such rebase diff was present here.

## Commands and outputs

Toolchain shells source /workspace/adamic-tools/env.sh; TMPDIR=/tmp/adamic-gate.
Earlier setup passed ready=0s, warm-cache=116s, total=116s; nproc remains 5.

```sh
ADAMIC_WAVE27_THIRD_ARTIFACTS=/workspace/wave-27-scratch/f801-third \
ADAMIC_WAVE27_REPOSITORY_MANIFEST=/workspace/wave-27-scratch/repository.manifest \
ADAMIC_WAVE27_COMPILER_MANIFEST=/workspace/wave-27-scratch/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-27-scratch/typescript \
python3 stage1/cohere/typeaware/wave_27_third/validate.py > /workspace/wave-27-resume-regex-validation.log 2>&1
```

The runner rebuilds stage 0, normal and Go-ASan archives, native and --sanitize
binaries, and the independent production Go oracle. It prints PASS. Of 639
candidates, 605 are Go-parseable and 34 exclusions remain recorded, not passed.
Controls match 225707 bytes and 437 findings. The 287-file repository corpus
matches 18485 bytes, zero findings; the 77-file compiler corpus matches 5934 bytes,
zero findings. Full ranges, messages, fixes and ordered suggestions/edits match.
Normal and sanitizer comparisons agree; sanitizer stderr is empty.

The new regex mutant changes Effect to EffectNever. It compiles and exits 0 with
empty stderr; only the Go byte comparison catches byte 28395. The existing rest,
Hook message and regex arity mutants are caught at bytes 54, 4962 and 112016 with
the same normally-exiting behavior. Released declaration-ancestry and binding
queries panic 70; retained-registry mutants exit 0 and fail the required refusal.
The mutant runner now inserts replacement source literally: Python replacement
processing had interpreted the JS regex backslashes and refused the first run.
The successful complete rerun is the retained validation result.

A separate exact-literal native probe matches Go's four boolean results for
useEffect, useLayoutEffect, useEffectEvent and useEffect followed by LF. Those
booleans alone do not expand the default recognized-hook set.

Three alternating whole-process samples per engine retain identical output.
Repository median native 561.254ms versus Go
224.730ms (2.497x);
compiler native 4440.112ms versus Go
518.373ms (8.565x).
These include loading, parsing, judgments, rendering and teardown, not just
regex execution. They establish no speedup.

## Not covered

Runtime additionalHooks options cannot be compiled and remain blocked. Other
nondefault options, shared registry integration, emitted-JavaScript comparison
and the full repository gate were not run. The other nine default ports remain
byte-unchanged on the unchanged main base and retain their preceding landing
results. React HIR/single-assignment/capture claims remain parked. The previous
nine-rule numeric sidecars still await the shared handed-node migration. No
protected compiler, shared parser, harness or registration-generator file changed.
