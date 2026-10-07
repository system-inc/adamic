Built three .a rule ports: structure/tailwind-no-physical-direction, @next/next/no-assign-module-variable, @typescript-eslint/default-param-last.
Commits: pushed claim 0a9a5ae6; Next b83fd28c; TypeScript 411d5137; Tailwind 5bcb5848.
Checks: final overlay suite PASS 258.271s, 840 supported comparisons and 37,132,151 identical canonical bytes on Go, source Node, emitted JavaScript and sanitized native; registry, vet and a filtered uncached oracle pass.
Mutants: wrong module name, ignored optional parameter and removed rtl/ltr exemption all compile/run cleanly and are caught only by comparison on all three Adamic executions.
Not covered: one Tailwind JSX fixture, default registration/package integration, pinned CLI self-lint of .a, the full repository gate and arbitrary parser recovery.

Existing branch work was clean and already pushed before this continuation. All
origin heads were fetched without submodules. Main was ef3d907e. The first 45
helper-ready positions were named by other claims or documented as earlier ports;
position 46 was the first available. After it, the first two available entries in
inventory.json's `syntax ready for AST/API adaptation` wave were selected. The
selection checked every claims Markdown artifact on every fetched origin branch,
27 distinct blobs, and found no mention of these candidates. The actual claim was
committed and pushed before any rule source was written. See evidence/selection.json
for the ref and claim-blob snapshot. No new foundation merge, rebase or PR occurred.
The requested docs/parallel-work.md is still absent; docs/lint-registration.md
supplies the directory registration contract. Only the claim and owned rule
directories are changed. Every newly authored Adamic module is .a.

The Next rule inspects every declaration in a VariableStatement, accepts only
plain identifier names, reports the whole statement once, and preserves the
upstream exclusions for imports, parameters, destructuring and loop headers.
The TypeScript rule finds the last plain parameter, reports preceding defaulted
or optional parameters, skips bodyless declarations and preserves method,
constructor, accessor, modifier, destructuring and rest behavior. Scanner tokens
rather than an equals-sign text search distinguish initializers from comments
and type contents. Neither rule has a fix or suggestion upstream.

Tailwind preserves all twenty physical/logical mappings, exact and negative
family distinctions, Go Unicode Fields whitespace, the anchored class grammar,
variant prefixes, rtl/ltr exemptions, decoded literal text, string/template
listeners and the original filename gate (.ts and .tsx only). Static template
pieces belong to expression templates, not interpolated template types. Its
policy placeholders substitute in Go's sorted order, including recursive-looking
placeholder text supplied by a class. It reports whole literal pieces once per
violating token and offers no fix. Own witnesses include Unicode escapes,
Unicode whitespace, punctuation, template types and message placeholders.

The final corpus contains all 77 TypeScript src/compiler files at
050880ce59e30b356b686bd3144efe24f875ebc8 and all 141 stage1 .ts/.a files on this
branch, including the generated registry. Each of the three rules runs over
all 218 files. Cohere remains pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db.
Its own rule bodies and tests are unchanged. Capturing its real tests preserves
filename extensions and deduplicates by filename, source, options and public name.
The comparison oracle chooses TS/TSX/JS from that filename; forcing TS for JSX was
an initial harness failure, retained in continuation-parity-1.log and corrected
only in the scratch harness. All independently filtered Go family tests exit 0.
The combined capture has 472 unique captured inherited/new combinations; only
these three new rules' cases enter the final rule comparisons.

| Rule | Captured own cases | Supported comparison rows | Identical bytes |
| --- | ---: | ---: | ---: |
| structure/tailwind-no-physical-direction | 54, one JSX gap | 272 | 12,374,429 |
| @next/next/no-assign-module-variable | 13 | 232 | 12,356,510 |
| @typescript-eslint/default-param-last | 117 | 336 | 12,401,212 |

Each row count includes the 218 compiler/stage1 files and one expanded owned
witness. Findings, full descriptions, IDs, ranges, repair fields and full fixed
sources agree byte for byte. Successful executions require exit 0 and empty
stderr; native correctness builds enable ASan/UBSan and leak checks. Emitted
JavaScript runs independently through oracle/node.mjs.

The Tailwind JSX fixture is kept in gaps/jsx-attribute.ts.txt. Its unchanged Go
rule test passes and the Go oracle accepts it. Source Node, emitted JavaScript
and sanitized native all exit 70 with the identical parser panic at position 22,
expected GreaterThanToken and got Identifier. It is explicitly excluded from
successful parity, not counted as a clean result. There are no parser gaps in
the other two rules' captured cases or the 218-file corpus. A prior return-void
JSX gap also remains an explicit inherited capture exclusion, outside this unit.

Every test wrote directly to a log. Reproduction uses the complete Go source
snapshots under evidence as scratch overlays and never modifies shared sources:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/evidence/reproduce.py \
  --compiler /workspace/scratch/typescript-6.0.3
```

Exact final commands, with logs in /tmp/lint-wave1-13 and copied unchanged to
evidence/continuation-*.log:

```sh
GOFLAGS=-overlay=/tmp/lint-wave1-13/overlay.json bash cloud/setup.sh > /tmp/lint-wave1-13/continuation-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test -overlay=/tmp/lint-wave1-13/overlay.json ./stage1/cohere/lint -run '^TestWave13Next' -count=1 -v -timeout=20m > /tmp/lint-wave1-13/continuation-final.log 2>&1
go test -overlay=/tmp/lint-wave1-13/overlay.json ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=10m > /tmp/lint-wave1-13/continuation-owned.log 2>&1
go test -overlay=/tmp/lint-wave1-13/overlay.json ./stage1/cohere/lint/registry -count=1 -v > /tmp/lint-wave1-13/continuation-registry.log 2>&1
go vet -overlay=/tmp/lint-wave1-13/overlay.json ./... > /tmp/lint-wave1-13/continuation-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /tmp/lint-wave1-13/continuation-oracle.log 2>&1
```

All six commands exit 0. Setup timing lines and nproc:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (20s)
setup: done in 20s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
nproc: 5
```

Go 1.27.1, clang 20.1.8 and Node 24.19.0. Expanded owned witnesses also match
selected and all-rule Go/Node/sanitized-native outputs, 30,593 bytes, PASS 15.182s.
The final suite passes corpus/backends in 213.92s and mutants in 44.34s, total
258.271s. Registry PASS 0.021s includes rejection controls. Vet exits 0 with an
empty log. The filtered uncached oracle PASS 0.325s shows zero native/Node cache
hits and one miss each. The earlier continuation-parity-2.log was green before
filename deduplication, expanded witnesses and shared immutable mapping tables;
only continuation-final.log is credited as the final corpus/mutant result.

| Compiling semantic mutant | Independent comparison that catches it |
| --- | --- |
| next-no-assign-module-variable-wrong-answer | Compare against moduleName instead of module: missing module statement findings on source Node, emitted JavaScript and sanitized native |
| typescript-default-param-last-wrong-answer | Ignore optional parameters: missing optional-before-required finding on the same three executions |
| tailwind-direction-aware-exemption-removed | Remove rtl/ltr exemptions: extra rtl:mr-4 finding on the same three executions |

All nine mutant executions exit 0 without stderr before their outputs are
compared. No compiler error, panic, warning or sanitizer result is credited as
a mutant kill. The uncached external one-byte oracle control separately proves
its comparator can fail.

Best of three interleaved count-only rounds on the supported mixed corpus and
owned witness. Counts match across Go, native and Node in every round. Native
timing uses an unsanitized release build; correctness uses sanitized native.
Startup, reading and parsing are included, builds/fixing/output formatting excluded.
These are measured workload rates, not isolated visitor or production guarantees.

| Rule | Findings | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: |
| structure/tailwind-no-physical-direction | 42 | 32.84 | 47.66 | 192.79 |
| @next/next/no-assign-module-variable | 9 | 7.06 | 9.88 | 42.07 |
| @typescript-eslint/default-param-last | 105 | 81.82 | 112.65 | 499.51 |

Best elapsed native/Node/Go seconds: Tailwind 1.278881/0.881162/0.217849;
Next 1.274893/0.910605/0.213909; TypeScript 1.283347/0.932064/0.210208.

Default integration remains blocked and is not called green. The unmodified
registry looks for next-no-assign-module-variable/rule.ts and exits 1. The default
lint package cannot compile profile_test.go:32 because it ranges over the old
portFiles value, now a function. Both logs are retained. evidence/required-infrastructure.patch
is the previously proposed .a discovery/copy/profiling repair, unapplied because
shared files are outside this unit's territory. It does not add JSX support or
the scratch harness's filename-sensitive capture/parser improvements. The final
comparisons use all three scratch Go overlays; no shared dispatcher or corpus
list is hand-edited. This remains a validated candidate, not merge-ready.

Pinned cohere CLI self-lint of all six .a modules, with --no-fix --lint --no-cache,
exits 1: nothing to check, .a is not a TypeScript or JavaScript file. That log is
retained and no clean self-lint is claimed. The Adamic loader/typechecker and both
backends do accept these modules, as their builds show. No .ts source copy is used.
No full repository gate, arbitrary malformed-source recovery, complete config/
suppression frontend or nondefault policy-catalog validation is claimed.
