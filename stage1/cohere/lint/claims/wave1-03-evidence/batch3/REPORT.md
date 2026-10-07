Built: claimed three syntax rules and reproducible independent Go repair-shape proofs; no new ports completed.
Commits: pre-code claim e084224600ef7f8fc3ef39a827f7651d16ac2fbd; all evidence is committed on codex/lint-wave1-03.
Commands and outputs: probe PASS; 16 upstream Go test functions PASS; three shared oracle refusals exit 2; overlay setup PASS in 12s, nproc 5.
Mutants: none for this batch because no port was implemented; no failed build or serializer panic is credited as a caught semantic mutant.
Not covered: Node, emitted JavaScript, sanitized native parity, compiler/stage1 corpus parity, semantic mutants, findings/s and full gate for these three rules.

## Selection and claim

Pushed all earlier branch work, then fetched every origin head. Main remains
ef3d907ecdc4c771b016f7d9c52372def057a340. All names in the original 46-rule
helper-ready list occur in claim Markdown on origin branches. The inventory's
syntax ready for AST/API adaptation queue, in its published order, then yields:

1. @typescript-eslint/no-unnecessary-type-constraint
2. @typescript-eslint/prefer-as-const
3. @typescript-eslint/prefer-enum-initializers

selection.json preserves the fetched refs and exclusion ledger. It searches
39 distinct claim Markdown blobs and checks main's lint modules/descriptors.
Our origin branch is evaluated at its pre-claim c1e9724b tip, so the later claim
does not exclude itself retroactively. The claim update was pushed at e0842246
before creating these adapters or witnesses. No other rules were taken.

## Observations and implication

The shared Finding has one finding range, repair kind, replacement and suggestion
string. RuleContext.report constructs exactly that record. The independent Go
oracle explicitly refuses multiple fixes, multiple suggestions, or repair ranges
that differ from the finding range. Three minimal valid programs demonstrate
that each newly selected rule needs a different unsupported shape:

| Rule | Finding | Actual Go repair | Legacy oracle |
| --- | --- | --- | --- |
| no-unnecessary-type-constraint | 14..15 | Suggestion deleting 15..27 | unexpected suggestion shape |
| prefer-as-const | 9..14 | Fix deleting 7..14 and inserting at 22 | unexpected fix shape |
| prefer-enum-initializers | 17..19 | Three separate suggestions | unexpected suggestion shape |

Each legacy process exits 2. A separate scratch serializer prints every edit,
suggestion ID and description, and runs the same Go converging fixer. It exits 0
and confirms all ranges and complete repair sets. The annotation fix produces
let foo = 'bar' as const; with its original newline. Suggestions leave fixed
source unchanged. Neither harness changes cohere's actual rule implementations.

Inference: a byte-for-byte rule-local port cannot send these complete diagnostics
through the installed finding contract. Discarding suggestions, moving findings,
or collapsing the two edits would violate the requested parity even if final
fixed text looked similar. The shared model, driver serialization/fix application,
and oracle serialization need an agreed extension, in addition to the earlier
.a registry compatibility change. The upstream source also requires preserving
filename-dependent generic-arrow suggestions for .tsx, .mts and .cts.

CLAUDE.md explicitly says: "Never edit a dispatch, oracle, corpus or copied-file
list." Its rule ownership instruction limits implementation to each owned rule
directory. The earlier narrow exception request remains unanswered. No shared
source was changed and no permission is inferred from silence. These are reserved,
unported rules, with concrete evidence ready for the shared contract owner.

## Commands and output

Source /workspace/adamic-tools/env.sh. Go cohere pin:
715ba94f3608a6500086b1076ce5cb7e51b836db. Go 1.27.1,
Node 24.19.0, clang 20.1.8. nproc prints 5, quota 400000/100000.

Default bash cloud/setup.sh exits 1 during cache warming because the previous
.a candidate has no rule.ts. Timing lines: Go ready 0s, clang ready 0s, Node
ready 0s, submodules ready 0s; no final done line. The earlier scratch overlay
is the workaround. An intermediate retry exposed the registry's requirement
that every rules subdirectory contain a descriptor. Moving unregistered evidence
under this owned claim directory removes that additional failure. Final command:

```
GOFLAGS=-overlay=/tmp/lint-wave1-03-corpora/overlay.json bash cloud/setup.sh
```

Exit 0: Go ready 0s; clang ready 1s; Node ready 1s; submodules ready 1s;
build cache warm 12s; done 12s on 5 processors, 17.6 GB.

From the repository root:

```
python3 stage1/cohere/lint/claims/wave1-03-evidence/batch3/typescript-no-unnecessary-type-constraint/probe.py \
  --scratch /tmp/lint-wave1-03-batch3-published
```

The runner builds two Go oracle binaries through scratch overlays, checks the
three legacy refusals, verifies complete independent repair output, and runs:

```
go test ./internal/lint/rules/typescript -count=1 -v -timeout=5m -run '^TestNoUnnecessaryTypeConstraint'
go test ./internal/lint/rules/typescript -count=1 -v -timeout=5m -run '^TestPreferAsConst'
go test ./internal/lint/rules/typescript -count=1 -v -timeout=5m -run '^TestPreferEnumInitializers'
```

All output goes directly to log files. Final results: 4 test functions PASS in
0.010s, 7 PASS in 0.013s, and 5 PASS in 0.007s respectively, plus their subtests.
The first probe attempt failed its hand-counted expected insertion offset of 21;
Go observed the correct end offset 22. The corrected assertion passes. That
failed expectation is preserved and is not a semantic rule mutant.

These Go checks establish the blockers, not completed port coverage. Native,
Node and Go findings/s comparisons are unavailable for this batch because no
valid comparable Adamic port exists. Previous batch measurements remain in its
own report. No .ts Adamic sources, shared edits or pull request were created.
