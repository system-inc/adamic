Built three .a helpers: jsx.MatchExactly, imports.NormalizedFileName and react.IsHookName, with seven consumers each.
Commit: feadff6; claims 470ddd3, bec68c8 and b7786fb were pushed before code; prior retained work was already pushed at bf07786.
Checks: final package PASS in 19.437s; 1,126,029 Go verdicts matched Node source and sanitized native; vet, types, formatting and six filtered oracle fixtures passed.
Mutants: unwanted case folding, first-backslash-only replacement and ASCII-only uppercase matching all compiled, ran cleanly and were caught by Go comparison.
Not covered: whole-rule findings/fixes/suggestions, emitted JavaScript, production parser integration, dynamic/external fixtures, invalid-UTF-8/lone-surrogate representation checks and the full repository gate.

## Branch, authorization and claims

Branch codex/lint-helpers-05. This unit claims helpers, not individual rules. All earlier retained helpers were tested and pushed before new work; the latest request authorized the next three after that condition. Every reservation explicitly fetched every origin codex/lint-helpers* branch and inspected all six remote claims directories. The older comment bundle in HELPERS.md was respected too. Strict option decoding/schema validation is the existing common helper bundle, not another unclaimed public Go symbol.

All larger-count public symbols were reserved. Each selected helper had seven consumers, tied for the largest remaining unclaimed count. MatchExactly was claimed at 470ddd3, NormalizedFileName at bec68c8 and IsHookName at b7786fb, all pushed before their code. Later slot 03 claimed MatchExactly in be16b65 at 19:11:33 UTC-06:00, or 01:11:33 UTC. This slot's 470ddd3 at 01:09:43 UTC precedes it; slot 05 retains the helper. Ownership resolution was pushed in 0e12a44. The other two helpers had no competing claim at final inspection.

All implementation, oracle and evidence files stay under slot05/batch3. Only this slot's claims and README handoff are also updated. Shared registration, shared rule harness, compiler files, other worker directories and the Go cohere worktree are unchanged. All new Adamic sources are .a. No PR, force push or branch deletion was used.

## Behavior and consumers

| Helper | Consumers | Final listed blocker removed by this batch alone |
|---|---:|---:|
| jsx.MatchExactly | 7 | 0 |
| imports.NormalizedFileName | 7 | 1 |
| react.IsHookName | 7 | 0 |

[CONSUMERS.md](CONSUMERS.md) lists all 21 rules by helper. [readiness.json](readiness.json) retains their residual dependencies and conservative cumulative accounting. This batch removes 21 dependency occurrences across 21 distinct rules.

Inference from the frozen ledger: NormalizedFileName removes the last listed blocker for base/consistency-no-bare-throw. IsHookName combines with this slot's earlier imports.BindingsOf to remove the last listed blocker for structure/import-require-react-namespace. Helper readiness does not mean whole-rule diagnostics or fixes are implemented.

Across eight retained slot 05 helpers: 97 dependency removals, 64 distinct consumers, four final helper blockers removed. The four newly helper-ready rules are @next/next/no-location-assign-relative-destination, @typescript-eslint/no-import-type-side-effects, base/consistency-no-bare-throw and structure/import-require-react-namespace. The conservative ready total rises from 46 to 50, assuming the common AST adapter. Other workers' implementations are not assumed integrated.

Exact matching preserves case, NULs, empty strings, Unicode normalization distinctions and supplementary characters. Filename normalization returns empty on nil and replaces every backslash; roots, repeated slashes, dot segments and case are not cleaned. Its immutable arena projection supplies actual SourceFile.FileName, with -1 as nil.

Hook naming requires exact use followed by a Go Unicode uppercase rune, without imposing a condition on the remainder. Digits and titlecase-only letters decline; non-ASCII and supplementary capitals match. A sorted/strided Unicode 17.0.0 table generated from Go unicode.Upper lives inside the same helper file. Binary search and stride membership preserve its classification. Regeneration must reproduce the checked-in file byte for byte.

## Independent oracle

Actual Go cohere at 715ba94f3608a6500086b1076ce5cb7e51b836db decides all results. Pin drift, missing consumer evidence and Unicode regeneration drift fail tests. The generator reads every nonempty Go string literal in every inventory-listed test file for each consumer. Exact matching and hook naming also parse strings as TSX and collect decoded identifier and literal texts. Prose/options are included: these are not complete-fixture counts. Dynamic source expressions and external corpus files are not reconstructed.

| Helper | Test-file literals | Derived/control coverage | Go verdicts |
|---|---:|---|---:|
| Exact match | 1,188 | 2,085 source/parsed/control texts, six pair variants each | 12,510 |
| Filename | 527 | 545 filename/nil cases | 545 |
| Hook name | 498 | 910 source/parsed/control names plus every 1,112,064 Unicode scalar | 1,112,974 |
| Total | overlapping strings | 21 consumers | 1,126,029 |

Every verdict is compared with Node source through oracle/node.mjs and sanitized native. Successful baselines and semantic mutants must exit 0 without stderr. ASan/UBSan and Linux leak checking are enabled. Per-consumer counts and deterministic corpus hashes are committed in evidence/.

Observation: ordinary Go parser construction requires normalized absolute filenames and refuses Windows-style/empty names before this helper is reached. An oracle-only constructor sets the actual private field read by FileName, without changing FileName or NormalizedFileName, allowing direct checks of the promised backslash behavior. The overlay leaves the submodule unchanged.

The hook sweep constructs use plus each Unicode scalar with a compact JSON instruction; only surrogate code points, which are not scalars, are skipped. Actual Go IsHookName supplies expected output. JavaScript case conversion never decides the baseline.

## Commands and outputs

From repository root after sourcing /workspace/adamic-tools/env.sh. Test output went directly to logs, never through a pipe.

    ADAMIC_SLOT05_BATCH3_EVIDENCE=/workspace/adamic/stage1/cohere/lint/helpers/slot05/batch3/evidence ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch3 -count=1 -v -timeout=10m > /tmp/lint05-batch3-final.log 2>&1

PASS 19.437s: matcher 6.36s, filenames 4.82s, hook names 8.25s. All baselines and compiling mutants passed. The earlier complete package also passed in 18.640s. See evidence/package-final.log.

    go vet ./stage1/cohere/lint/helpers/slot05/batch3 > /tmp/lint05-batch3-vet.log 2>&1
    go run ./cmd/adamic types stage1/cohere/lint/helpers/slot05/batch3/main.a > /tmp/lint05-batch3-types.log 2>&1
    gofmt -l stage1/cohere/lint/helpers/slot05/batch3 > /tmp/lint05-batch3-format.log
    git diff --cached --check

All exit 0. Vet and formatting logs are empty; types printed checked entry-point variables, parameters and functions. Staged whitespace checks passed.

    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > /tmp/lint05-batch3-oracle.log 2>&1

PASS 0.965s, all six fixtures, zero cache hits and six probe misses. See evidence/filtered-oracle.log. The final touched package and filtered uncached oracle were run; the full repository gate was not.

Initial runner corrections: missing os/exec import was restored before matcher comparisons. The filename oracle's parser assertion was resolved by the test-only constructor. The common driver initially read a nonexistent optional right field for filename cases, and native lowering refused an inferred never[] in the nil-file branch. Optional fields and explicit SourceFileName[] fixed both inside this owned runner. These failures are not mutant kills. Both subsequent complete packages passed cleanly.

## Every mutant

| Helper | Compiling mutation | Go witness |
|---|---|---|
| Exact match | Compare lowercased strings | Verdict 15: mutant true, Go false |
| Filename | Replace only the first backslash | Line 125: later backslashes remain; Go replaces every separator |
| Hook name | Require scalar below 128 | Verdict 887: mutant false, Go true |

All three mutants compiled, exited 0 without stderr and differed from actual Go. The matcher also passed standalone in 7.872s. Filename and hook mutants first passed in the corrected complete package and again in the final package. Compiler/sanitizer failures are never credited.

## Toolchain and limits

The successful original setup was reused. Go 1.27.1, clang 20.1.8, Node 24.19.0. nproc again printed 5. /opt/adamic-tools/env.sh is absent here; setup installed /workspace/adamic-tools/env.sh. Original timing lines:

    setup: go ready (0s)
    setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
    setup: node ready (1s)
    setup: submodules ready (1s)
    setup: build cache warm (88s)
    setup: done in 88s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB

Not covered: full rule findings/fixes/suggestions; emitted-JavaScript comparison; production parser/linter/shared-rule-harness integration; dynamic fixture expressions; runtime-loaded external corpora; arbitrary malformed file arenas; raw invalid-UTF-8/lone-surrogate representation checks; full repository gate. Source and sanitized-native .a paths work, so no .ts fallback is needed.
