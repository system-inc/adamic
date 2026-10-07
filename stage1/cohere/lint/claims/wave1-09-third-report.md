Built: pushed three more claims and reproducible prerequisite evidence; no new rule ports.
Commits: prior work e37fe824 fully pushed; pre-code claim ad1519c2; evidence commit follows.
Commands and outputs: setup exit 0 in 17s, nproc 5; 311 refs and 33 claim documents audited; 14 upstream tests, registry tests and uncached byte oracle passed.
Mutants: native one-byte output mutant caught by external comparison; eleven inherited descriptor controls rejected; no new per-rule semantic mutants.
Not covered: rule implementations, complete findings/fixes parity, three rule mutants, corpus throughput and full gate; .a registration and repair serialization remain blocked.

## Allocation

Continued on codex/lint-wave1-09. The initial push printed Everything up-to-date.
Fetched all origin heads with:

```
git fetch --no-recurse-submodules origin '+refs/heads/*:refs/remotes/origin/*'
```

Audited 311 refs and 33 Markdown claim documents. Main remains ef3d907e;
inventory remains 73ac2eb0963e1a4166eaa0fbd160203f11dcdbdf. The original
46-rule helper list linked by helpers/REPORT.md is exhausted. The first three
available syntax inventory rows, preserving inventory order, are:

1. @typescript-eslint/no-dupe-class-members
2. @typescript-eslint/no-empty-object-type
3. @typescript-eslint/no-import-type-side-effects

Syntax-only excludes needs_type_information and binding_only. Existing claims
from this worker count as occupied. Main implementation/registration sources
are scanned outside inventory, helper and testdata artifacts. The claim update
was committed and pushed as ad1519c2 before any rule code. This is not an
exhaustion result for the inventory: other unclaimed syntax rules remain.

wave1-09-third-evidence/audit.py and selection.json retain the allocation,
branch SHAs, queue and exclusions. To replay the pre-claim state, the script
pins this worker's own branch to e37fe824; all other refs come from the fetch.
Running it after another fetch can change those other refs and the selection.

## Observed prerequisites

The shared directory registry still reads rule.ts, emits imports ending in
rule.ts and rejects mutant paths without a .ts suffix. The user requires new
Adamic modules to be .a. The executed registration control copies the five
baseline rules, successfully generates them, then renames only no-debugger's
module to .a without changing its bytes. Generation exits 1 with:

```
open .../rules/no-debugger/rule.ts: no such file or directory
```

No candidate rule can register as .a through this foundation. The unchanged
probe-registration.py in the preceding evidence directory reproduces this
control; its new output is retained in registration.log here. This is a
prerequisite probe, not a semantic-mutant pass. The binding rule-directory
ownership instructions prohibit editing shared dispatch, oracle and copied-file
lists. No shared registration, parser or compiler file was changed.

There are also executed serializer blockers for two of the three rules.
probe-contract.py builds the existing lint/testdata/oracle.go byte for byte
through a scratch Go overlay, adding only an isolated selection of the three
unmodified upstream rules. Count-only mode finds exactly one diagnostic for
each of the four rule witnesses. Normal mode produces these results:

| Witness | Go finding count | Inherited formatter result |
|---|---:|---|
| class A with two foo methods | 1 | exit 0 |
| empty interface | 1 | exit 2, unexpected suggestion shape |
| empty type alias | 1 | exit 2, unexpected suggestion shape |
| inline type-only import with two names | 1 | exit 2, unexpected fix shape |

NoEmptyObjectType offers two suggestions, and interface rewrites do not share
the diagnostic's name-only range. NoImportTypeSideEffects produces multiple
edits: remove each inline qualifier and insert the top-level qualifier. The
inherited adapter asserts one repair at exactly the diagnostic range. Silently
dropping repairs would fail the requested byte-for-byte contract. The rule
bodies and upstream tests were not changed. Full stack traces and the successful
duplicate-member output are retained in the individual go-rule.log files.

## Parser evidence

Unlike the preceding JSX allocation, these syntax witnesses are supported.
The five raw .ts.txt inputs are duplicate methods, an empty interface, an empty
type alias, an inline-type import and an ordinary declaration control. A fresh
sanitized parser binary and emitted JavaScript were built using the already
committed build_probe.go. Every input exits 0 on Node source, emitted JavaScript,
sanitized native and the independent Go parser. Each complete tree output is
byte-identical across all four backends on these five inputs. This does not
prove corpus-wide rule parity. No parser failure is claimed as a blocker for
these particular inputs.

Reproduce from the repository root with the toolchain environment sourced:

```
go run stage1/cohere/lint/claims/wave1-09-evidence/build_probe.go > /tmp/wave109-build.log 2>&1
python3 stage1/cohere/lint/claims/wave1-09-third-evidence/probe-parser.py > /tmp/wave109-parser.log 2>&1
python3 stage1/cohere/lint/claims/wave1-09-third-evidence/probe-contract.py > /tmp/wave109-contract.log 2>&1
python3 stage1/cohere/lint/claims/wave1-09-next-evidence/probe-registration.py > /tmp/wave109-registration.log 2>&1
```

probe-parser.py preserves each backend output separately; the native builder
uses ASan and UBSan. Raw witness files are .ts.txt, outside the module graph.
No new authored .ts module was added.

## Checks and mutants

bash cloud/setup.sh wrote setup.log and exited 0. Environment file:
/workspace/adamic-tools/env.sh. nproc printed 5. Exact timings:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (17s)
setup: done in 17s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

From cohere, ran:

```
go test ./internal/lint/rules/typescript -run '^(TestNoDupeClassMembers|TestNoEmptyObjectType|TestNoImportTypeSideEffects)' -count=1 -v -timeout 15m > /tmp/lint-wave1-09-third-upstream.log 2>&1
```

Exit 0; 14 top-level tests passed, including the empty-object upstream corpus;
package time 0.049s. This is upstream verification, not Adamic rule verification.

From the repository root, ran:

```
go test ./stage1/cohere/lint/registry -count=1 -v > registry-tests.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout 10m > oracle-byte-mutant.log 2>&1
```

Both exit 0, package times 0.009s and 0.283s respectively. The byte oracle
reports native and Node zero cache hits, one miss each. Its native stdout
one-byte mutant runs and is rejected by the external comparator. Eleven
inherited descriptor controls fail validation: duplicate name, unknown field,
missing hook, invalid kind, missing oracle exports, no listener, missing factory,
missing class, missing finish hook, unsafe name and duplicate oracle adapter.
Those descriptor rejections are not semantic mutants for these three rules.
Neither the serializer panics nor the registration failure receives mutant
credit. No per-rule implementation exists to mutate.

## Limits

No new rule is registered or marked ported. Complete findings and fixes over
TypeScript compiler sources, stage1 sources and upstream rule cases were not
compared. Native, Node and Go findings-per-second rates are unavailable; no
rate is fabricated from parser probes or upstream test durations. No full gate
or repository-wide vet was run for these claim/evidence-only edits. No PR was
opened. The concrete next prerequisites are shared .a registration/test-copy
support and a serializer/finding model preserving multiple repairs and
suggestions. Reservations remain on this branch for continuation once those
foundations are available.
