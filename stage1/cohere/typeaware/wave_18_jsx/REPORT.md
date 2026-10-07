Built: jsx-fragments, jsx-no-undef and no-adjacent-inline-elements in native Adamic, with numeric listeners and a kind-indexed private driver.
Commits: claims b899b5007 and eligibility correction 12019f4ae; native code 937dc881, based on main f8013f0b.
Checks: 130 valid controls, 110 default findings; option counts 96/107/93; both corpora, sanitizers, checker, Go production tests, vet and filtered uncached Node pass.
Mutants: fragment member byte 446, undef guard byte 8288, adjacency byte 27811, numeric fact byte 43; all byte-only; stale registry violates required exit 70.
Not covered: parked HIR/capture rules, shared parser/harness integration, full root gate, arbitrary JSX projects or migration to an unspecified Diagnostic SHA.

## Selection, parking and ownership

The branch was landing-ready at 8cf6bf224 before continuation: it contains
current main f8013f0b, and all six previous suites were re-green and pushed.
529 fetched origin refs and 33 distinct claim Markdown blobs were inspected.
Ranking combines VOLUME_REPORT.md's compiler-all and repository-all counts,
descending total then lexical name. Five higher-volume Adamic/Nexus candidates
are already baseline ports in COVERAGE_REPORT.md and were skipped. None of
the selected three was claimed or had a named port on an origin branch.

The earlier three React HIR rules are explicitly parked. The initial sixth-batch
reservation also included jsx-no-constructed-context-values. Its separate
stability module requires callback/callee return and capture/escape analysis,
so the reservation was corrected and that rule parked before implementation.
No-adjacent-inline-elements is the next eligible ranked replacement. Both
claim commits were pushed before code. No capture-dependent port is advertised
as complete. Analysis modules are being ported on #dnv6f2c and JSX integration
is landing on area/stage1-lint. Only codex/typeaware-wave-18 is pushed.

All new Adamic modules are .a inside this owned directory. The only shared
source edit is the single jsx-syntax-facts registration line in facts.go.
The Go question and Adamic decoder are separate named files. No shared harness,
registration generator, parser or protected compiler file was changed.
The existing Diagnostic is used; no batch-8 migration SHA was supplied.

## Numeric listeners and facts

Each rule has rule.json kinds matching its syntaxKinds declaration:

| Rule | Numeric SyntaxKind listeners |
| --- | --- |
| jsx-fragments | 289 JsxFragment, 285 JsxElement, 286 JsxSelfClosingElement |
| jsx-no-undef | 287 JsxOpeningElement, 286 JsxSelfClosingElement |
| no-adjacent-inline-elements | 285 JsxElement, 214 CallExpression |

The private driver fetches each primary node once, uses a map indexed by numeric
kind, and passes that node to relevant callbacks. Rules do not read source kinds
as strings or refetch their primary node. Child/declaration traversal uses
query-local numeric identities. The incoming shared driver is not edited.

The parser for this increment is typescript-go inside the existing C bridge.
One root question per file exports raw numeric syntax for the listener kinds,
the required structural links/children, byte token ranges, decoded text units,
symbol presence and complete declaration identities. It makes no lint judgment
and emits no findings, repairs or suggestions. Symbol metadata is requested for
JSX name references and createElement identifiers. UTF-16 unit frames protect
embedded NUL text from the C string ABI. A Go test independently checks binding
resolution, source spans, identities and NUL transport, and rejects malformed
questions. This does not port JSX parsing into Adamic's native parser; it is
explicit use of the bridge parser's raw numeric facts.

Native code owns fragment import/initializer recognition, prop exclusions and
mode decisions; intrinsic/custom/member-tag resolution and module-global rules;
inline-name classification, Go Unicode edge-whitespace behavior, createElement
import gates and first-adjacent-pair reporting. Fragment fixes remain withheld
exactly as in pinned Go. All three produce zero fixes and zero suggestions,
verified as complete fields rather than assumed from counts.

## Independent oracle and results

The oracle is an overlay under testdata inside CoHere. It imports no bridge
code, loads its own program, dispatches unchanged production registry rules
and serializes complete findings/fixes/suggestions. Its parse filter retains
130 of 132 extracted/independent candidate source strings; the two invalid
strings and both manifests are preserved. Controls include Unicode/CRLF,
shadowed and merged bindings, fragment aliases/imports/destructuring, generic
JSX tags, global bindings, intrinsic/custom/namespaced/this tags, sibling gaps,
foreign inner calls, empty strings, exotic whitespace and embedded NUL.

The fresh reproduction builds from an empty scratch directory and passes:

| Population/options | Findings |
| --- | ---: |
| Controls/default | 110 |
| Controls/fragment element mode | 96 |
| Controls/allowGlobals | 107 |
| Controls/both options | 93 |
| Compiler, frozen 77 sources | 0 |
| Repository, frozen 287 sources | 0 |

Every row matches normally and under ASan/UBSan/LeakSanitizer, with empty
native and sanitizer stderr. Canonical byte lengths depend on artifact header
paths: the first default run matched 43013 bytes; fresh-run lengths are retained
in reproduce.log. Compiler matches 5318 bytes and repository 18485 bytes.
All source manifests and hashes are archived without truncation.

Each rule mutant compiles, exits 0 and has empty stderr; only the independent
Go byte comparison catches it. Fragment member equality is reversed (byte 446),
undef's resolved-symbol guard is reversed (8288), and adjacency membership is
reversed (27811). The first run's shorter header paths yield bytes 426/8158/27131.
A raw-question mutant adds one to numeric kinds and likewise compiles/runs
normally, caught at byte 43. A new-question request after program release exits
70 with invalid or released checker handle. Retaining the released registry
entry makes the same request succeed with exit 0 and empty stderr; the required
exit-70 contract catches it. Mutants are not counted if compilation fails.

## Commands and time

    bash cloud/setup.sh > /tmp/wave18-jsx-setup.log 2>&1
    source /workspace/adamic-tools/env.sh
    export ADAMIC_WAVE18_COMPILER_CONFIG=/workspace/wave-18-typescript/src/compiler/tsconfig.json
    export ADAMIC_WAVE18_COMPILER_MANIFEST=/tmp/wave-18-compiler.manifest
    export ADAMIC_WAVE18_REPOSITORY_CONFIG=/workspace/adamic/tsconfig.json
    export ADAMIC_WAVE18_REPOSITORY_MANIFEST=/tmp/wave-18-repository.manifest
    python3 stage1/cohere/typeaware/wave_18_jsx/validate.py --scratch /workspace/wave18-jsx-reproduce > /tmp/wave18-jsx-reproduce.log 2>&1
    python3 stage1/cohere/typeaware/wave_18_jsx/validate_bridge.py --scratch /workspace/wave18-jsx > /tmp/wave18-jsx-bridge-validation.log 2>&1
    go test ./bridge/tsgo/checker -count=1 -timeout=10m
    go -C cohere test ./internal/lint/rules/react -run '^(TestJsxFragments|TestJsxNoUndef|TestNoAdjacentInlineElements)' -count=1 -timeout=10m -v
    go vet ./...

The bridge mutation command uses the artifacts produced by validate.py; either
validated scratch directory can be supplied. Setup prints Go 0s, clang 0s,
Node 0s, submodules 0s, cache warm 34s, done 34s; nproc 5, CPU quota 4.
Checker PASS 0.129s, Go production tests PASS 0.076s, vet exit 0. Filtered
uncached Node PASS 1.173s, native misses 19 and Node misses 13, including
closures/method closures/generics/regions and the one-byte oracle mutant.
All test output went directly to logs. Native files were formatted through
CoHere's private native formatter; Go files are gofmt-clean.

Initial build probes exposed unsupported chained optional access, an inferred
never array, a constructor used directly as a statement, and numeric console
output; owned source/probe syntax was adapted before successful validation.
The first checker compile selector matched no tests and is not reported as
coverage; the final whole checker package was actually run. No compiler or
shared infrastructure fix was made.

Three alternating backend rounds measure whole process time over each corpus:

| Corpus | Native median seconds | Go median seconds |
| --- | ---: | ---: |
| Compiler | 1.776647 | 0.327015 |
| Repository | 0.278631 | 0.130370 |

Native remains slower; no speedup claim is made. The original run measured
1.710237/0.292275 and 0.273420/0.135200 respectively. These are complete source
pipelines, including checker loading and raw fact serialization/decoding.
No full root gate or full JSX application population was run. Suppression and
edit application are outside this private runner, as in previous wave suites.
Numeric facts and the standalone listener driver are implemented; integration
with the incoming shared driver and native JSX parser remains integration work.
