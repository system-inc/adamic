Built: all prior retained rules complete and pushed; next three React hook rules claimed, with reproducible blockers rather than completed ports.
Commits: prior implementations 182d776e, evidence 2ccc0f30; next claim d8478862; blocker evidence committed separately on codex/typeaware-wave-30.
Commands and outputs: fresh all-origin fetch; 197 ranked, 25 base ports, 139 claims across 359 origin refs; prerequisite probe PASS 8.681s; vet PASS.
Mutants: prior exit/state, blocking/state, graph/fork and timer/range mutants caught by independent Go; no diagnostic mutants for the unported React rules.
Not covered: preserve-manual-memoization, purity and refs implementations, their byte-equivalence gates, sanitizers, released handles or native timings.

The next selected rules are react-hooks/preserve-manual-memoization,
react-hooks/purity, and react-hooks/refs. Their claim was pushed before writing
any probe. The snapshot records exact origin heads, claim blobs, token-boundary
matches and combined ranking. Every higher-ranked candidate is either ported on
the checked base branches or mentioned in an origin claim file. Mentions of
released claims are conservatively excluded too.

The existing native parser has no JSX syntax nodes. The private probe builds
through the native Adamic toolchain and calls the same stage1 Parser used by the
lint suites. The independent Go oracle loads the same three .tsx aliases, accepts
them without parse diagnostics and runs the unmodified production rule registry.
Each control produces one finding, with zero fixes and suggestions:

- purity: function Component(){return <div>{Math.random()}</div>;}.
  Go reports impureFunctionCall. Native parser exits 70 with expected
  CloseBraceToken, got DotToken at 38.
- refs: function Component(props){const value=props.ref.current;return <div>{value}</div>;}.
  Go reports refValueAccess. Native parser exits 0 but produces
  TypeAssertionExpression and no JSX node, so treating successful parsing as
  JSX support would be wrong.
- preserve-manual-memoization: a useMemo with non-optional body dependency and
  optional written dependency, returning <Foo data={data}/>.
  Go reports preserveManualMemoizationValueUnmemoized. Native parser exits 70
  with expected GreaterThanToken, got Identifier at 118.

TestWave30ReactPrerequisites is a passing reproduction of these missing
prerequisites, not a passing rule comparison. The initial probe incorrectly
expected every JSX input to panic; the observed refs type-assertion parse is
explicitly captured by the revised probe. Both logs are retained.

There is also no native React HIR lowering/SSA and analysis pipeline in this
branch. PreserveManualMemoization calls ForFunction, CloneFunction and
AnalyzePreservedManualMemoization, including OutlineFunctions, InferReactive,
DropManualMemoization, mutable range inference, reactive scope assignment,
alignment/merging, dependency collection and multiple scope-pruning passes.
Purity requires ForFunction/AsCompilationUnit and forward abstract-value transfer;
Refs requires ForFunctionWithoutManualMemoization, AsCompilationUnit and its
fixpoint sweep. The new native process graph has no SSA phi/value arena and does
not supply these analyses. Moving these judgments into a Go bridge question would
violate the requirement that native Adamic makes the rule decisions.

The shared .a module-loader work does not itself add JSX parsing or React HIR.
I stopped under Ahra's instruction to report other blockers rather than edit
shared files. The new three claims remain retained and unported; no further rules
were claimed. This report does not assert that standalone rule wrappers or empty
corpus results satisfy the port bar. A native JSX substrate and native HIR
pipeline are the prerequisites for completing these rules faithfully. No shared
parser, registration generator, test harness, or protected compiler file was
changed in this batch.

Commands, with output saved to logs: source /workspace/adamic-tools/env.sh;
ADAMIC_WAVE_30_REACT_ARTIFACTS=/workspace/wave-30-react-prerequisites-reviewed
go test ./stage1/cohere/typeaware -run '^TestWave30ReactPrerequisites$' -count=1 -v;
go vet ./stage1/cohere/typeaware. Source controls are authored .a and exposed to
Go through .tsx symlinks. Serialized Go findings, native parser trees/panics,
source controls and hashes are in validation-wave-30-fourth.

For the completed third-batch ports and measured native/Go times, see
[WAVE_30_PROCESS_REPORT.md](WAVE_30_PROCESS_REPORT.md).
