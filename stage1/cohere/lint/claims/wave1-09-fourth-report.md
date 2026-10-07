Built: pushed three new claims and reproducible prerequisite evidence; no new rule implementations.
Commits: prior work 4348b268 fully pushed; claim 76a01059 pushed before code; evidence commit follows.
Commands and outputs: setup exit 0 in 21s, nproc 5; 320 refs and 39 claim documents audited; 15 upstream tests passed; four parser inputs matched all four backends.
Mutants: inherited native one-byte output mutant caught by external comparison, uncached; no per-rule semantic mutants.
Not covered: new rule ports, complete findings/fixes parity, rule mutants, findings-per-second rates and full gate; .a registration and suggestion serialization remain blocked.

## Selection and publication

Continued on codex/lint-wave1-09. First push printed Everything up-to-date,
then fetched every origin head without submodule history recursion:

```
git fetch --no-recurse-submodules origin '+refs/heads/*:refs/remotes/origin/*'
```

Main remains ef3d907e; inventory remains 73ac2eb0963e1a4166eaa0fbd160203f11dcdbdf.
Audited 320 origin refs and 39 Markdown claim documents. The original 46-rule
helper handoff linked by helpers/REPORT.md is completely claimed. The first
three available syntax inventory rows, retaining JSON order, are:

1. @typescript-eslint/no-non-null-asserted-optional-chain
2. @typescript-eslint/no-non-null-assertion
3. @typescript-eslint/no-this-alias

Syntax-only excludes needs_type_information and binding_only. Existing claims
from this worker count as occupied. Main implementation and registration source
are searched outside inventory/helper/testdata artifacts. The exclusive claim
update was pushed as 76a01059 before implementation. The inventory is not exhausted.

The owned audit.py and selection.json preserve queue decisions and branch SHAs.
The replay pins this worker's own ref to its pre-claim commit 4348b268. Other
refs follow the fetch, so another fetch may change replay results.

## Executed prerequisites

The installed directory registry requires rule.ts, emits rule.ts imports and
rejects mutant files without a .ts suffix. New Adamic source must be .a under
the user's instruction. The real generator accepts all five baseline rules,
but renaming just no-debugger's identical module bytes to .a makes generation
exit 1 with rule.ts: no such file or directory. registration.log retains the
fresh run of the preceding probe-registration.py. This is not a semantic mutant
and fails before rule compilation. Shared generator/test copier changes are
outside the rule-directory ownership contract; none was made here.

The inherited lint Go oracle also cannot serialize the non-null suggestions.
probe-contract.py builds it through a scratch overlay selecting the three
unmodified upstream rules. It changes only the parser's synthetic FileName to
remove the raw witness's .txt suffix, so filename-sensitive rules see .ts or .js
as intended; the rule bodies, diagnostic collection and serializer are unchanged.
Count mode returns exactly one finding for each rule witness. Normal mode:

| Witness | Go finding count | Inherited serializer |
|---|---:|---|
| foo?.bar! | 1 | exit 2, unexpected suggestion shape |
| foo!.bar | 1 | exit 2, unexpected suggestion shape |
| const self = this, .ts | 1 | exit 0 |

The optional-chain rule removes only the assertion operator, a narrower range
than its finding. The assertion rule's property access suggestion contains two
edits, removing ! and replacing the dot. The inherited adapter requires exactly
one suggestion containing exactly one edit at the diagnostic range. Omitting
these suggestions would lose requested parity. Stack traces are retained in the
individual go-rule.log files. No failed serializer run is counted as a rule pass.

The same this-alias source with a .js filename returns zero findings, as upstream
requires. That control is retained separately. The shared parser exposes its
path, so filename access itself is not asserted as an infrastructure blocker.
NoThisAlias has no repair serializer blocker on the executed witness; its
remaining common prerequisite is .a registration.

## Parser and upstream checks

The four raw .ts.txt inputs are an optional-chain assertion, an assertion before
a property access, a this alias and an ordinary declaration control. All exit 0
and print byte-identical complete trees on Node source, emitted JavaScript,
sanitized native and the independent Go parser. probe-parser.py asserts those
comparisons. The parser binary and JavaScript were built in the prior allocation
by the committed build_probe.go; only claim/evidence files changed since.
The native artifact uses ASan and UBSan. Rebuild and reproduce from the repo root:

```
source /workspace/adamic-tools/env.sh
go run stage1/cohere/lint/claims/wave1-09-evidence/build_probe.go > /tmp/wave109-build.log 2>&1
python3 stage1/cohere/lint/claims/wave1-09-fourth-evidence/probe-parser.py > /tmp/wave109-parser.log 2>&1
python3 stage1/cohere/lint/claims/wave1-09-fourth-evidence/probe-contract.py > /tmp/wave109-contract.log 2>&1
python3 stage1/cohere/lint/claims/wave1-09-next-evidence/probe-registration.py > /tmp/wave109-registration.log 2>&1
```

bash cloud/setup.sh exited 0. Exact timing lines, with nproc 5:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (21s)
setup: done in 21s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

From cohere, ran:

```
go test ./internal/lint/rules/typescript -run '^(TestNoNonNullAssertedOptionalChain|TestNoNonNullAssertion|TestNoThisAlias)' -count=1 -v -timeout 15m > /tmp/lint-wave1-09-fourth-upstream.log 2>&1
```

Exit 0, 15 top-level tests passed; package time 0.013s. This verifies upstream,
not an Adamic implementation. Full output is upstream.log. From the repo root:

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout 10m > /tmp/lint-wave1-09-fourth-mutant.log 2>&1
```

Exit 0, package time 0.562s; native and Node each report zero cache hits and one
miss. The native one-byte output mutant executes and external output comparison
catches it. No registration failure, parser probe or serializer panic receives
per-rule mutant credit. There are no new rule implementations to mutate.

## Coverage limits

No new rule directory or authored .ts module was added. All rule behavior was
observed through the independent Go implementation. Complete findings/fixes over
TypeScript compiler sources, stage1 sources and upstream rule cases were not
compared. Findings per second for native, Node and Go are unavailable, not zero.
No full gate or repository-wide vet was run for claim/evidence-only changes.
The next prerequisites are shared .a registration and test-copy support, and
finding/serializer support for separate suggestion spans and multiple edits.
No shared compiler or dispatch file was changed and no PR was opened.
