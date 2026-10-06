# Enum coverage

Base: origin/codex/flag-enums, 28d1db9, on origin/codex/enums, 7127080.
Read CLAUDE.md, README.md, docs/0.1.md, docs/memory.md, the two commit
messages, the main-to-feature diff and its changed files. Read the separate
origin/cloud/grok-enums-coverage diff before selecting programs.

16 new entry programs, plus one imported support file. No output disagreements
on the restored branch. Every entry passed source Node, generated JavaScript,
native release, ASan/UBSan and LeakSanitizer, and a separate public CLI build/run.
No compiler implementation changes are retained. Counts gained exactly 16 rows;
all existing rows are unchanged. All new allocations equal frees.

In the tables, paths are relative to internal/oracle/testdata. New paths start
with enums_coverage_. Grok paths refer to its separate branch, not files copied
into this branch. Internal nil, symbol-kind and representation dispatch guards
are grouped with the source-language cases that reach them. Rejection paths
cannot produce a successful three-way runtime fixture.

| Code and cases | Existing use | New use or limitation |
| --- | --- | --- |
| enumConstant: finite numeric, implicit numbering, explicit numbering, unary negative, fractional, string, member references | enums.a, enums_const.a | constants.a adds other-enum constants, implicit successor of 16, unary plus, subtraction, division, remainder, right/unsigned shifts, AND, XOR, exponentiation; const negative/fraction/string member |
| enumFields: insert forward, insert reverse only for numbers, replace duplicate reverse names | enums.a; Grok enums_values.a | constants.a tests aliases produced by different operators |
| enumFields: zero normalization; numeric key spelling, integer order vs other keys | enums.a | numeric_edges.a adds negative exponent and smaller/larger exponential boundaries |
| enumFields: arbitrary ordinary string member names | none | string_names.a: whitespace, Unicode, embedded newline |
| enumFields: string NaN/Infinity names allowed; empty enum | none | reflection.a |
| enumLocal/enumDeclaration/declareModule/statement: regular runtime object; const erased declaration | enums.a, enums_const.a | const_flags.a: const flag combinations and Set |
| enumAssignable/enumIdentity: same declaration, compatible member, unions; member aliases | enums.a, enums_modules | generic.a, views.a: full numeric/string/flag types, readonly widening, validation returning optional enum |
| enumAssignable: runtime enum object and typeof aliases retain exact identity | enums.a | lookup.a stores actual enum objects in readonly array; reflection.a passes alias to reflection and Object.assign |
| enumObjectView/invariance: actual enum object alias, nested exact enum objects | enums.a | lookup.a |
| enumMember/enumExpression: property, string literal bracket, const inline, parentheses | enums.a, enums_const.a | lookup.a: literal bracket and template-literal bracket |
| enumExpression: dynamic number reverse lookup, missing key, string member-name union, side-effect key | enums.a | lookup.a: receiver and key both have side effects, hit and miss; verifies both counts |
| enumExpression: optional property, present and absent object | enums.a | lookup.a adds numeric alias receiver and undefined fallback |
| enumExpression: dynamic lookup helper evaluates object and key once, scans compatible representation, undefined fallback | enums.a | lookup.a, modules/main.a |
| enumExpression: helper with proven forward-name union and unreachable panic fallback | enums.a | Existing successful programs cannot reach the panic without breaking the type proof |
| enumSwitch: numeric/string, aliases counted by value, narrowed member, exhaustive/no default | enums.a, enums_const.a; Grok enums_switch.a | switches.a adds literal cases and exhaustive void switch |
| enumSwitch: default covers remaining values, including default before cases | enums_flags.a | switches.a: numeric and string partial switches, default-first ordering |
| enumDefaultUnreachable: closed exhaustive never default omitted; flag defaults retained | enums_flags_never_default.a, enums_flags.a | const_flags.a and flag_slots.a observe reachable combination/zero defaults |
| library_object: enum is exact shape for keys; homogeneous string values and entries | enums_modules/main.a uses string values | reflection.a: Object.entries, alias reflection, same-type Object.assign, hasOwnProperty |
| library_object: numeric Object.values/entries rejected due to reverse strings | no successful fixture | Deliberate homogeneous reflection restriction |
| modules: named import, renamed import, type-only import, const import, dependency order | enums_modules, enums_flags_modules; Grok enums_linked | modules/main.a: imported generic function and flag keys/members with imported string values |
| generic functions: member-specific enum instantiation | enums_const.a | generic.a: explicit full enum instantiations, higher-order callback, generic array return, generic class; modules/main.a: imported generic |
| invariance/class_inheritance: safe readonly arrays, maps, fields, invariant class argument | no enum widening fixture | views.a: readonly number views and child of Base<Rank> |
| enumInitialization: function declarations skipped until call; const aliases created before enum initialization; initialized reads afterward; no regular enums fast path | existing const and module fixtures | initialization.a: direct defaulted call, immutable function alias, readonly callback property, array callback, instance field initializer |
| flagEnum: every initializer explicit, zero, 1 << literal 0..30, OR with same-enum member leaves; parentheses skipped | enums_flags.a | flag_proofs.a: hex/octal/binary literal spellings and nested parenthesized OR; flag_slots.a has no declared zero |
| flagEnum: implicit/nonzero literal/arithmetic/direct alias remain closed; no members remain closed | enums.a, enums_const.a; Grok enums_values.a | reflection.a adds empty enum; constants.a remains closed despite bitwise constants; shapes.a proves direct alias, bare literal and zero-leaf OR all remain closed with no-default switches |
| flagTarget: full enum only, nullable removed, member-specific slot not combination | enums_flags.a; rejection unit tests | flag_proofs.a exercises undefined union; null representation is unsupported |
| flagDomainSeen: AND either proof, OR/XOR both proofs, declared identity on typed reads/calls | enums_flags.a | flag_keys.a, generic.a, modules/main.a |
| flagDomainSeen: complement only under AND; nested operators | enums_flags.a | flag_slots.a removes highest bit from array slot |
| flagDomainSeen: both conditional arms proven | none | flag_proofs.a exercises both arms at runtime |
| flagDomainSeen: immutable inferred aliases recursively followed, cycle guard, parentheses | enums_flags.a has one alias | flag_proofs.a: three aliases and parenthesized XOR |
| flagUpdate: &= arbitrary number, |= and ^= proven operand, variable/field symbols | enums_flags.a has local updates | flag_slots.a: object and this fields; array element compound updates lack a symbol and are refused |
| flagDomainSeen: AND with arbitrary numbers and foreign enum, both operand orders | enums_flags.a uses -1 | mask_edges.a: NaN, infinities, fractional truncation, signed boundary, 32-bit wraparound, foreign enum mask |
| flags stored in fields, class fields, arrays, Map values, function parameters/results | enums_flags.a | flag_slots.a, generic.a |
| flags as Map keys/Set members, combined highest two bits and low bit, XOR high bit, duplicate keys, iteration/deletion | none; Grok covers closed numeric/string keys | flag_keys.a: FEATURES with bits 0,29,30; const_flags.a: const Set; modules/main.a: imported flags |
| flag reverse lookup: undeclared combinations missing, declared pair/high hit | enums_flags.a | flag_slots.a observes zero absent when enum has no zero member |
| flag initializer bound: literal 0..30 only, recurse through OR | enums_flags.a hits 0 and 30 | flag_proofs.a and flag_keys.a add 29; sign bit 31 is deliberately refused |

## Cases that cannot become successful runtime programs

The branch explicitly rejects these (its existing lowering tests cover them):

* Arbitrary numbers, bare literals or foreign enums written as closed enums;
  arithmetic/increment/decrement or compound updates on closed enum slots.
* Widening writable arrays, maps, fields, function/method views or generic class
  arguments; incompatible overrides or Object.assign; structural enum-object
  copies, including copies nested in arrays; writable views of enum slots.
* Complement, arithmetic, shifts, increments, bare numeric values, foreign-enum
  OR/XOR, unproven OR/XOR operands, two unproven AND operands, mutable inferred
  number aliases and member-specific combination slots in flag contexts.
* Exclusion narrowing of a full flag slot into a specific member; a flag switch
  without a default, even with all declared members; incomplete closed switches.
* Computed/non-finite members, ambient, merged, local/block declarations;
  unqualified members outside initializers; mixed-representation switches.
* Pre-initialization reads through direct calls, aliases, defaults or callbacks;
  unknown callback/indirect call and construction before initialization;
  multi-step optional enum chains; unproven indexes/forward member names.
* Prototype-changing or NUL-containing names, numeric NaN/Infinity/-Infinity
  names, numeric heterogeneous values/entries, and flag shifts outside 0..30.

Additional forms attempted during this work failed before native execution:

| Form | Observed refusal and responsible location |
| --- | --- |
| identity<Features>(Features.Low | Features.High), likewise imported generic | Type-argument invariance refuses inferred number before flag proof; internal/lower/generic.go:49 (instantiateFunction). Typed intermediate works and is covered |
| Features \| null parameter | Stage 0 cannot represent this type; successful fixture uses undefined union |
| value?.['Low'] on typeof Enum | Optional bracket access limited to tuples; internal/lower/collections.go:335. Optional dot access is covered |
| flagsTuple[0] \|= Enum.Member | enum_flags.go flagUpdate requires an updated symbol; tuple element has none. Explicit regular array writes with undefined fallback are covered |
| Explicit tuple element assignment | Stage 0 cannot lower assigning an element of a value; use regular array |
| case -2 | Unary negative case is not an IR constant under object.go's switch check; use enum member case or positive literal |
| Defaulted function as an alias value | Function values with optional/rest parameters unsupported; direct defaulted call plus required-parameter alias are covered |
| Static readonly field on class with no instance | Stage 0 cannot lower reading Holder; instance field initializer is covered |

TypeScript also rejected unchecked array compound reads and returning solely from
numeric literal cases without a trailing return; those were invalid input, not
compiler output disagreements. The console prelude takes strings, so numeric
reflection length was converted to a template string.

## Mutation evidence

Temporarily changed exactly one line in internal/lower/enums.go enumConstant:
`return ir.NumberConstant{Value: value}, nil` to
`return ir.NumberConstant{Value: value + 1}, nil`.

The flag_keys fixture compiled and finished at exit 0 with empty stderr on every
backend. Node printed keys 1610612737 and 536870913; native and generated JS
printed 1610612739 and 536870914. The oracle failed only on stdout comparisons.
Restored the line and reran the same oracle successfully. See mutant.log and
restored.log. No mutant compiler code is committed.

## Toolchain and commands

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (76s)
setup: done in 76s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc`: 5. Go 1.27.1, clang 20.1.8, Node v24.19.0.

From /workspace/adamic:

```sh
git fetch origin codex/flag-enums:refs/remotes/origin/codex/flag-enums codex/enums:refs/remotes/origin/codex/enums cloud/grok-enums-coverage:refs/remotes/origin/cloud/grok-enums-coverage
git fetch origin main
git switch -c coverage/enums origin/codex/flag-enums
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
gofmt -w internal/oracle/enums_coverage_test.go
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/enums_coverage' -count=1 -v -timeout 30m
python3 notes/enums/run_builds.py
# Expands to Node oracle/node.mjs <file>, then for each of 16 entries and their support file:
# go run ./cmd/adamic build <file> -o /tmp/enums-coverage-builds/<name>
# /tmp/enums-coverage-builds/<name>
# With the one-line mutant, then after restoring it:
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/enums_coverage_flag_keys.a' -count=1 -v -timeout 30m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
gofmt -l cmd internal
go vet ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...
git diff --check
```

Test output was redirected to /tmp/enums-coverage-*.log and read afterward.
The targeted oracle ran five times during fixture development; its first three
runs rejected unsupported/invalid probes, then the 14-entry and final 16-entry
runs passed. The full gate was launched with the first 14 entries. The final two
entries were added during that gate and received a fresh 16-entry uncached
oracle, CLI builds, and regenerated counts afterward.
The first CLI pass likewise recorded compile refusals for draft forms; the saved
runner was then run on all 16 final entries plus the support file and every comparison passed.

Final results: the full uncached repository gate exited 0 (Unicode properties
690.824s, whole oracle package 195.666s, native package 220.674s). This gate used
the initial 14-entry registry. The final 16-entry uncached oracle passed in
4.142s, the final counts update in 22.565s, all 17 CLI source files agreed, and
final formatting, vet and whitespace checks were clean. See gate.log, oracle.log,
counts.log and cli.log for the observed output.
