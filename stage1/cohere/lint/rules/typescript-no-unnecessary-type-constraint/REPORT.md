Built: six owned .a rule ports, including the original slot's three and the three repair-heavy claims; shared integration and seven JSX profiles remain explicit gaps.
Commits: original pre-code claim db16927 and repair-heavy pre-code claim e0842246; each rule is committed separately on codex/lint-wave1-03.
Commands and outputs: 496 comparable fixture/profile cases match Go on Node, emitted JavaScript and ASan/UBSan native; compiler/stage1 corpora match; registry/vet pass through the scratch .a overlay; setup 19s, nproc 5.
Mutants: six compile and run, then fail only output comparison against unchanged Go on all three backends; names and evidence below.
Not covered: installed .a registry support, installed full-repair serialization, seven upstream JSX profiles, prior Tailwind JSX gap, full repository gate and arbitrary invalid option inputs.

## Owned implementation

All original slot claims now have rule directories: no-unused-expressions,
unified-signatures and class-methods-use-this. Their syntax decisions, decoded
option defaults, messages, finding ranges, witnesses, upstream adapters and
mutants are implemented. No shared source was edited. Their findings have no
fixes or suggestions and fit the existing finding model.

The next three claims also have complete listeners and typed diagnostic records.
No-unnecessary-type-constraint keeps the finding on the parameter name and the
suggestion's independent edit range, including filename-dependent generic-arrow
comma insertion. Prefer-as-const compares cooked strings/canonical numbers and
preserves both annotation edits; its as-expression arm has one safe fix.
Prefer-enum-initializers preserves all three ordered suggestions, their IDs,
messages, positional values and the upstream quoted-name behavior.

Owned records.a holds every fix and suggestion. Owned drivers expose them for
validation while the shared contract remains unchanged. The registered finish
hooks of repair-heavy rules explicitly refuse unsupported repair shapes. They do
not silently truncate a repair or report a partial success. The single-range
as-expression fix can use the existing context directly.

The announced codex/lint-harness-dot-a branch was absent from the fetched origin
heads. Default registry code still requires rule.ts, and the installed Finding
and Go serializer still cannot carry independent edit ranges, multiple edits or
multiple suggestions. Those changes belong to the harness worker. The earlier
compatibility proposal is used only through scratch Go overlays for setup,
registry validation and vet. This work does not install it or edit shared files.

## Independent parity

validate.py builds an owned driver and a Go oracle through scratch overlays.
Cohere's parser, rule bodies, messages, formatter and converging edit engine stay
unchanged. Only the scratch serialization exposes every independently computed
fix/suggestion. The compared output contains formatted findings, byte ranges,
IDs, complete edit and suggestion records, and converged fixed source.

validate-original.py uses the same structure for the original three rules. It
preserves source, filename and decoded JSON options in its capture key and
manifest. Options are loaded in the source driver and the independent upstream
Go adapter. A first default-only comparison was superseded by this complete
profile comparison, because identical source under distinct settings must not
be deduplicated.

| Rule | Captured upstream profiles | Explicit JSX gaps |
| --- | ---: | ---: |
| no-unused-expressions | 191 | 7 |
| unified-signatures | 119 | 0 |
| class-methods-use-this | 48 | 0 |
| no-unnecessary-type-constraint | 43 | 0 |
| prefer-as-const | 69 | 0 |
| prefer-enum-initializers | 21 | 0 |

Seven unused-expression JSX rows have independently successful Go verdicts and
exit 70 with parser slice refusals on Node, emitted JavaScript and sanitized
native. These are logged individually and listed in evidence/original-gaps.json.
They are excluded openly from comparable parity, not counted as matching output.
The JSX listener decision is implemented, but the shared parser cannot supply
those nodes. Tailwind's previously proven JSX attribute gap remains unchanged.

Original profiles plus owned witnesses: 354 comparable cases, 116,619 identical
bytes on all three paths. Repair-heavy profiles plus witnesses and extra numeric,
Unicode and .tsx/.mts/.cts corner cases: 142 cases, 50,531 identical bytes.
The finalized record constructor was then rebuilt and the repair-heavy cases
rechecked: 51,035 identical bytes; different scratch path lengths explain the
byte-total difference.

Pinned TypeScript compiler 050880ce59e30b356b686bd3144efe24f875ebc8:
repair-heavy rules match over 162 compiler/stage1 files per rule, 36,995,234 bytes;
original rules match over 169 files per rule, 36,701,026 bytes. These are two
source snapshots during implementation, not 331 distinct files. Each includes
all 77 compiler files and then-present stage1 .ts/.a programs, excluding
.generated and deliberately invalid gaps. Full output SHA-256 values and sizes
are in evidence/index.json; compressed canonical Go output is retained and every
matching path's independent hash is recorded. Other raw logs are retained.

## Commands

Source /workspace/adamic-tools/env.sh. Go cohere is pinned at
715ba94f3608a6500086b1076ce5cb7e51b836db. Go 1.27.1,
Node 24.19.0, clang 20.1.8. All test output goes to files, never pipes.

```
GOFLAGS=-overlay=/tmp/lint-wave1-03-corpora/overlay.json bash cloud/setup.sh
```

Exit 0: Go ready 0s; clang ready 1s; Node ready 1s; submodules ready 1s;
build cache warm 19s; done 19s on 5 processors, quota 400000/100000, 17.6 GB.

From the repository root:

```
python3 stage1/cohere/lint/rules/typescript-no-unnecessary-type-constraint/validate.py \
  --scratch /tmp/lint-wave1-03-full-final --typescript /tmp/lint-wave1-03-typescript --mutants --benchmark
python3 stage1/cohere/lint/rules/typescript-no-unnecessary-type-constraint/validate-original.py \
  --scratch /tmp/lint-wave1-03-original-options --mutants
python3 stage1/cohere/lint/rules/typescript-no-unnecessary-type-constraint/validate-original.py \
  --scratch /tmp/lint-wave1-03-original-final --typescript /tmp/lint-wave1-03-typescript --benchmark
python3 stage1/cohere/lint/rules/typescript-no-unnecessary-type-constraint/validate.py \
  --scratch /tmp/lint-wave1-03-repair-finalized
```

All PASS with the explicit JSX exclusions above. Native builds use ASan/UBSan;
every successful compared process must exit 0 with empty stderr. Release native
is used only for timing. The builders use load/lower/native/javascript through
an owned explicit-file Go overlay; no compiler or test-harness source is changed.

```
go test -overlay=/tmp/lint-wave1-03-corpora/overlay.json ./stage1/cohere/lint/registry -count=1 -v
go vet -overlay=/tmp/lint-wave1-03-corpora/overlay.json ./...
```

Registry PASS 0.032s, including existing metadata rejection probes. Vet exits 0
with an empty log. The complete repository gate was not run. The normal shared
lint parity gate remains blocked by its module/repair contract; these independent
profile drivers prove rule kernels, not installed integration.

Two local compiler refusals were corrected without shared edits: a postfix
increment used as a value became an explicit statement, and numeric || in a
comparator became a conditional. An empty overload array received an explicit
type. Initial capture instrumentation used uppercase JSON names, then was
corrected to the capture's lowercase keys. None counts as a semantic mutant.

## Mutants and findings per second

| Rule | Compiling semantic mutant | Native/s | Node/s | Go/s | Findings |
| --- | --- | ---: | ---: | ---: | ---: |
| no-unused-expressions | identifier_expression_ignored | 893.29 | 1104.78 | 5537.45 | 1003 |
| unified-signatures | parameter_difference_silenced | 758.99 | 1098.54 | 5282.23 | 1043 |
| class-methods-use-this | empty_method_silenced | 822.37 | 979.43 | 5150.84 | 1000 |
| no-unnecessary-type-constraint | any_constraint_ignored | 895.40 | 1119.38 | 5480.24 | 1000 |
| prefer-as-const | annotation_fix_changes_type | 872.16 | 1080.02 | 5399.82 | 1000 |
| prefer-enum-initializers | second_suggestion_wrong | 1570.28 | 1861.98 | 9086.87 | 1680 |

Each mutant compiles and runs successfully on all three paths before only the
Go output comparison catches it. In particular, the enum mutant changes only
the second suggestion, and the annotation mutant changes only the inserted fix,
so findings-only checks would miss both. Raw mutant output is in evidence.

Timing uses the best of three interleaved rounds with startup and parsing
included, 77 compiler files plus 1,000 explicitly planted rule violations.
Native/Node/Go counts agree in every round. These are observed worker rates,
not compiler-only rates or a claim native beats Go.
