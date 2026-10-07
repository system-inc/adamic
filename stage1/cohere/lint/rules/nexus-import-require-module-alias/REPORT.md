Built: nexus/import-require-module-alias in .a, validated through an isolated registry overlay; production registration is blocked.
Commits: registration merge 0d59239, helpers merge 661b9e0, pushed claim 5098bbd, implementation 08b34c1.
Commands and outputs: owned suite PASS 105.109s; overlay vet PASS; registry tests PASS 0.013s; filtered oracle PASS 10.500s; overlay setup PASS 20s, nproc 5.
Mutants: suppress the namespace-style finding, caught only by Go output comparison on all three backends; remove duplicate-key refusal, caught by the explicit refusal check on all three.
Not covered: shared production integration, one upstream JSX source, Go duplicate-key merge behavior, the full repository gate, or revalidation of the two skipped ports.

## Assignment and claim

The report supplied by the task does not itself enumerate the rules. It links
`../../HELPERS.md`, whose 30 option-ready entries followed by 16 policy-ready
entries define positions 40 through 42:

| Position | Rule | Disposition |
| --- | --- | --- |
| 40 | nexus/consistency-no-stuttering-name | Skip: existing port on origin/codex/stage1-lint-batch4 |
| 41 | nexus/consistency-no-utils-folder | Skip: existing port on origin/codex/stage1-lint-batch4 |
| 42 | nexus/import-require-module-alias | This directory's implementation |

The skipped branch was inspected at
`d486b03a15f3202acc317dc81c40cc21818e21f7`. Its files are
`nexus_consistency_no_stuttering_name.ts` and
`nexus_consistency_no_utils_folder.ts`. Their presence and dispatch were observed;
this unit makes no new parity claim for them. All fetched origin branches were
searched for the remaining rule; no Adamic implementation was found.

Branch `codex/lint-wave1-14` starts at current main
`d090af531216ddd3c25a0dede6b82d7c0a6edf76`. Claim
`stage1/cohere/lint/claims/wave1-14.md` was committed and pushed before any rule
code was written. No pull request is opened.

Registration merged cleanly. Contrary to the task's clean-merge premise, the
helper branch conflicted in six shared lint files: README.md, lint.ts,
lint_test.go, main.ts, settings.ts and testdata/oracle.go. Resolution retains
the directory-registration versions of these files and the helper additions.
`docs/parallel-work.md` was absent from main and both foundations. Its committed
version at `7f958de` on origin/codex/no-shared-lists was read, along with
`docs/lint-registration.md` and the existing rules.

## What runs

`rule.a` handles ImportDeclaration, default and namespace bindings, type-only
imports, default-plus-named bindings, exact package matching, configured package
additions and overrides, empty names, omitted and null styles, null names,
unknown styles, and the upstream style-before-name reporting order. Named-only,
side-effect and unconfigured imports are untouched. The directory owns its
factory, descriptor, messages, options descriptor, Go adapter, witness and
semantic mutant. New descriptors omit order. No shared dispatch, corpus, oracle
selection or copied-file list was edited for the rule.

Messages and strict target data are extracted from the helper's Go-generated
resolved catalog and target descriptors, pinned to cohere
`715ba94f3608a6500086b1076ce5cb7e51b836db`. PolicyMessage supplies Go's sorted
placeholder substitution, including values that themselves contain placeholders.
StrictOptions and OptionsJson validate and decode the configured map. Go scalar
null values retain the zero value. Configured entries with empty names leave
the built-in entry intact. Strict validation runs in the factory, outside the
constructor, to avoid the observed nested-constructor compiler refusal. Disabled
factories do not validate another selected rule's options.

The rule proposes no fixes or suggestions, exactly as Go cohere does. The
comparison still uses Go's real report writer and converging edit engine and
compares complete fixed-source output, finding ranges, IDs, descriptions and
repair fields. The independent adapter executes the unmodified upstream rule.
No cohere, compiler, runtime or parser file was changed.

## Production integration is blocked

The merged registry hardcodes rule.ts in discovery and generated imports and
rejects .a mutant modules. The copied-module harness also only recognizes .ts.
Ordinary lint package compilation additionally fails because inherited
profile_test.go ranges over portFiles, which registration changed to a function.
The unmodified generator's observed failure is in
[evidence/registry-blocked.log](evidence/registry-blocked.log).

These shared infrastructure files are outside this unit's directory ownership.
[Integration.patch](integration.patch) is a concrete, unapplied compatibility
patch for the registry, .a copying/corpus discovery, and the profiling compilation
error. `git apply --check` accepts it. Its registry variant passed deterministic
regeneration and descriptor-rejection tests. It is proposed work, not a landed
infrastructure fix. This branch is not ready for an ordinary main merge until
that shared support is integrated.

[validate.py](validate.py) creates temporary Go overlays containing that support
and injects the owned `suite.go.txt` as a temporary test source. It changes no
shared worktree file. Its additional JSX exclusion is test-only, explicit and
logged; it is not included in integration.patch. The generated registry remains
ignored. All new Adamic files in this directory are .a.

## Commands and observations

After `source /workspace/adamic-tools/env.sh`, from the repository root:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-14-typescript-pinned \
  python3 stage1/cohere/lint/rules/nexus-import-require-module-alias/validate.py
```

The runner executes `go test -overlay=<printed temporary overlay>
./stage1/cohere/lint -run '^TestWave14' -count=1 -v -timeout=20m`, with stdout and
stderr written directly to [evidence/validation.log](evidence/validation.log).
Final result: PASS, 105.109s.

- TestWave14Rules: 27 unique upstream module-alias source/options cases, eight
  added option configurations, the owned positive witness, and an eqeqeq
  selection/option-isolation control. 42 total findings, of which 41 belong to
  this rule. Go, Node source, emitted JavaScript and sanitized native agree on
  17,816 canonical bytes.
- TestWave14Corpus: all 77 TypeScript v6.0.3 src/compiler files at
  `050880ce59e30b356b686bd3144efe24f875ebc8` and all 135 stage1 .ts/.a sources
  present during validation. All four agree on 12,334,992 canonical bytes.
  The combined compared output is 12,352,808 bytes per implementation.
- TestWave14Mutant and TestWave14DuplicateOptions: PASS, as described below.
- `go vet -overlay=<temporary overlay> ./... > /tmp/lint-wave1-14-vet.log 2>&1`:
  exit 0, empty log.
- `go test -overlay=<temporary overlay> ./stage1/cohere/lint/registry -count=1
  -v > /tmp/lint-wave1-14-registry-tests.log 2>&1`: PASS, 0.013s.
- `go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1
  -timeout=10m > /tmp/lint-wave1-14-oracle.log 2>&1`: PASS, 10.500s.
- `gofmt -l` on the owned Go adapter and staged `git diff --check`: empty output.

Each successful native comparison must exit 0 with no stderr under ASan/UBSan
and Linux leak checking. A failed build, runtime failure or sanitizer finding is
not credited as a semantic mutant catch. The emitted JavaScript is built through
Adamic's lowering/backend; source Node strips types from the original .a files.

The original `bash cloud/setup.sh` failed, exit 1, during cache warming at
profile_test.go:32: `cannot range over portFiles (value of type func(t *testing.T)
[]string)`. It printed go ready at 0s and clang, Node and submodules ready at 1s.
The workaround was:

```sh
GOFLAGS=-overlay=<temporary compatibility overlay> bash cloud/setup.sh \
  > /tmp/lint-wave1-14-setup-overlay.log 2>&1
```

It completed successfully and printed:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (20s)
setup: done in 20s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc` printed 5. Go 1.27.1, clang 20.1.8, Node 24.19.0; Linux x86_64,
AMD EPYC 9V74, four-core CPU quota. Original and recovered setup logs, vet,
registry tests, filtered oracle and machine observations are in evidence/.

Superseded attempt logs are retained. The first stopped comparison rejected Go's
successful dependency-download notices on stderr; it also started before the
compiler checkout was ready. The duplicate-guard attempt initially called the
parser's nonexistent parse method; it was corrected to file. Strict validation
inside the constructor produced the refusal recorded in the nested-constructor
attempt; moving it into the factory resolved it. None of these failed attempts
is reported as a green check. The final validation.log is the release evidence.

## Every new mutant

| Mutation | Observation and check that caught it |
| --- | --- |
| Change `if(!expectsNamespace)` to additionally require an empty source | The scratch variant compiles and exits successfully without stderr on source Node, emitted JavaScript and sanitized native. It omits the first witness's requireDefaultStyle finding. Only the canonical comparison against Go catches it. |
| Remove uniqueOptions(raw) | The scratch variant compiles and prints accepted on all three backends. Baseline instead produces exactly `adamic: panic: NotYet: repeated module alias option keys require Go merge semantics` and exit 70. The explicit refusal assertion catches the mutant. This is a gap-control mutant, not a claim of Go duplicate-key parity. |

The registry's inherited rejection mutants were also rerun under the proposed
compatibility overlay. Their log is evidence/registry-tests.log; they are not
new per-rule semantic mutants.

## Throughput

The natural compiler plus stage1 corpus has zero findings for this rule. Reporting
zero findings/s as positive throughput evidence would be misleading. The timing
workload therefore adds 500 copies of the four-finding owned witness, for 2,000
findings across 213 source files. This is a compiler/stage1-plus-positive-probe
workload, not a natural-corpus findings-rate claim.

Five interleaved rounds, best elapsed sample per implementation; every sample
checks the count against Go. Native timing is an unsanitized -O2 build; correctness
and mutants use sanitized native. Count mode includes startup, file reads,
scanning, parsing and visitors, excluding finding formatting and fixing.

| Implementation | Best seconds | Findings per second |
| --- | ---: | ---: |
| Native | 1.295240 | 1,544.12 |
| Node source | 0.886630 | 2,255.73 |
| Go | 0.220259 | 9,080.21 |

These are shared-machine observations, not performance guarantees. Raw rounds
and checked counts are retained in validation.log.

## Limits

- Production registration and the shared profiling compilation repair are
  unapplied. Ordinary unmodified generation still fails; overlay results must
  not be described as a passing production gate.
- One captured upstream source contains `<Reakt.Thing ... />`. The original Go
  test runs and passes during capture. Its source is explicitly excluded from
  comparison because the Adamic parser has no JSX support. Thus full upstream
  fixture parity is not claimed.
- Duplicate option keys can require encoding/json's partial struct/map merges.
  OptionsJson keeps only the last member, so the port explicitly refuses that
  shape rather than silently using different configuration. Arbitrary invalid
  option diagnostic prose, arbitrary configurations and Go duplicate-key
  merge behavior are not held byte for byte here.
- No parser recovery, full configuration/suppression integration or type/binding
  queries are claimed. The complete repository test gate was not run; the named
  owned suite, registry tests, repository-wide overlay vet, filtered external
  oracle and overlay cache-warm gate are the commands actually run.
- Rules 40 and 41 were skipped under the task's already-ported instruction;
  their code and existing evidence were not changed or revalidated.
