# Integer counters in fields: measurement stop

Measured the requested scanner hotspot; no field proof or emission change was built.
Base: b8fb957aa839a9e8cb0b54279dd9864fa317bd30, branch codex/integer-fields.
Callgrind: 10,804,405 pos conversion/add instructions, 3.047110% of 354,578,754 template self instructions.
Mutant: increasing the fresh profile summary by one fails self-cost accounting with exit 1.
Not covered: field proof, fractional/subclass mutants, backend oracle gates, parse counts or before/after optimization counts; the requested under-5% stop applies.

## Observation

The archived `origin/area/runtime` profile has `positions: line`, not instruction
addresses. `callgrind_annotate --auto=no --inclusive=no --threshold=100` reports
177,988,434 template self instructions in generated `main.c` and 176,590,320 in
inlined `adamic.h`, totaling 354,578,754. Line costs cannot isolate opcodes.

Reconstructed source commit c230f01 in a detached scratch worktree, referencing
the existing cohere submodule (715ba94f3608a6500086b1076ce5cb7e51b836db) by symlink.
No cohere source was copied. Generated C from `adamic c` and `TestProfileArtifacts`
compares byte for byte. The pinned TypeScript corpus is v6.0.3 at
050880ce59e30b356b686bd3144efe24f875ebc8, 77 compiler files, 434,790 tokens.

A fresh profile with `--dump-instr=yes --collect-jumps=yes` records
`positions: instr line`. Its total is 2,429,543,863 versus the archive's
2,429,614,991 (71,128 fewer process instructions). **Every template source-line
self count, including the inlined header and line zero, matches the archive
exactly.** Template self is identical. This establishes the attribution bridge
without pretending the archive itself contains instruction addresses.

`objdump -d --disassemble=adamic_function_19_Scanner_template` and `addr2line`
identify these executed instructions. Generated C was also annotated with
`callgrind_annotate`. Addresses below are from the reconstructed binary.

| Operation | Addresses | Executed instructions |
| --- | --- | ---: |
| pos to integer, first charCodeAt | 200d3, 200eb | 1,754 |
| pos to integer, loop charCodeAt | 20306, 2031a | 5,196,904 |
| pos + 1 to integer, dollar lookahead | 2055e, 20572 | 160 |
| pos to integer, carriage-return lookahead | 20a17, 20a2b | 112,102 |
| Double-to-integer total | | 5,310,920 |
| pos adds, initial increment | 201a0 | 2,093 |
| pos adds, ordinary increment | 20218 | 5,278,095 |
| pos adds, dollar lookahead | 20514 | 111 |
| pos adds, carriage return increment | 2093d | 105,560 |
| pos adds, newline increment | 20af5 | 105,560 |
| pos adds, closing delimiter | 20da0 | 1,983 |
| pos adds, interpolation opening | 20ea2 | 83 |
| Double-add total | | 5,493,485 |
| Combined | | 10,804,405 |

Both conversions generated for each unsigned size_t cast are counted, even
though the valid small position makes the second conversion's contribution to
the result unnecessary. Counts are executed opcodes, not source casts.

Five other adds, at 200b6, 202e5, 20451, 2053d and 209f6, belong to unsigned
string-length-to-double conversion. Their 10,879,750 instructions do not add
pos and are excluded. Counting every addsd in the method would give a misleading
6.115469% combined figure. The actual pos numerator is 3.047110%, below the
17,728,937.7-instruction cutoff. This is the requested measurement criterion,
not a prediction of the total gain of a different representation or runtime API.
Called string-runtime costs are outside the requested template self denominator.

## Reproduction and evidence

Toolchain: Go 1.27.1, clang 20.1.8, Node 24.19.0, Valgrind 3.24.0.
`bash cloud/setup.sh > /tmp/integer-fields-setup.log 2>&1` passed: Go ready 0s,
clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 123s,
done 123s. Environment file `/workspace/adamic-tools/env.sh`. `nproc`: 5;
cgroup `cpu.max`: 400000 100000; memory 17.6 GB.

Historical native flags: `-std=c11 -Wall -Wextra -Werror -pedantic
-Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function
-Wno-unused-parameter -Wno-self-assign -ffp-contract=off
-fno-optimize-sibling-calls -O2 -g`, linked with `-lm`.
No sanitizers or RC instrumentation in the profiling binary.

Commands, executed in the historical scratch worktree unless stated:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/integer-fields-typescript ADAMIC_SCANNER_PROFILE_DIR=/tmp/integer-fields-profile go test ./stage1/typescript/scanner -run '^TestProfileArtifacts$' -count=1 -timeout 30m > /tmp/integer-fields-artifacts.log 2>&1
VALGRIND_LIB=/tmp/integer-fields-valgrind/usr/libexec/valgrind /tmp/integer-fields-valgrind/usr/bin/valgrind --tool=callgrind --dump-instr=yes --collect-jumps=yes --callgrind-out-file=/tmp/integer-fields-instructions.callgrind /tmp/integer-fields-profile/profiled --manifest /tmp/integer-fields-profile/compiler.txt --count > /tmp/integer-fields-instructions.stdout 2> /tmp/integer-fields-instructions.stderr
/tmp/integer-fields-valgrind/usr/bin/callgrind_annotate --auto=no --inclusive=no --threshold=100 /tmp/integer-fields.callgrind > /tmp/integer-fields-annotate.log
/tmp/integer-fields-valgrind/usr/bin/callgrind_annotate --auto=no --inclusive=no --threshold=100 /tmp/integer-fields-instructions.callgrind /tmp/integer-fields-profile/main.c > /tmp/integer-fields-generated-C.log
PYTHONDONTWRITEBYTECODE=1 python3 /tmp/integer-fields-count.py > /tmp/integer-fields-count.log
```

Artifact test passed in 9.792s. Count output: 434790. Accounting sums every
non-call-edge instruction self cost to the whole-program summary; then compares
all template source-line counts with the archive. The summary-plus-one mutant
exits 1 with `AssertionError: (2429543863, 2429543864)` in
`/tmp/integer-fields-accounting-mutant.log`. No field-decision checks were added,
so the 0.5 and subclass proof mutants were not run.

Scratch evidence includes the generated C, profiled binary and manifest under
`/tmp/integer-fields-profile`, raw profiles `/tmp/integer-fields.callgrind` and
`/tmp/integer-fields-instructions.callgrind`, disassembly
`/tmp/integer-fields-template.asm`, attribution script
`/tmp/integer-fields-count.py`, and its output `/tmp/integer-fields-count.log`.
The branch contains only this report. No emitter worker paths were changed.

Representation hook, for a future field unit: `member` in
`internal/native/emit_values.go` selects `adamic_value.number` for numeric field
storage; `cType` selects double for numeric C values. Object fields currently
use uniform `adamic_value` slots, so choosing integer field storage would need
consistent slot access and widening. Neither hook was changed here.
