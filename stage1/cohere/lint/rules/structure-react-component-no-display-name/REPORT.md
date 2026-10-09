# structure/react-component-no-display-name

Base: `origin/area/stage1-lint` at `6bf7bcec73a04b85900c5a69a257b3cf34b35a40`.
Only this directory changes. Required origin fetch and both descriptor-name searches
found no existing port; see `evidence/origin-check.txt` and `evidence/fetch.log.gz`.
Setup passed in 32.291 seconds, nproc 5, four-CPU quota; all timing lines are
preserved in `evidence/setup.log.gz`. Source `/workspace/adamic-tools/env.sh`.

The port uses the existing React-name/base/namespaced-member helpers and the
shared `isLikelyReactComponent` and JSX-return detectors, with projections of
parser nodes. It retains lazy whole-file component-name lookup, wrapper exemptions
in declaration order, top-level declarations only, and plain dot assignments only.
The Go adapter returns unmodified `structure.ReactComponentNoDisplayName`; this
rule has no options. The exact policy description is copied into `messages.a`.
`upstreamTest: Test` captures both dedicated tests and the absent-optional-node
shared guard, filtered by exact rule name by the existing capture harness.

Go cohere is the oracle. **56 captured upstream cases** and **five owned witnesses**
passed, with selected and all-rule witness runs, plus the inherited corner corpus.
Every witness was separately required to produce a Go finding. Go, source Node,
emitted JavaScript and ASan/UBSan native output matched **535,030 bytes**; process
wall 195.312 seconds. See `evidence/selected.log.gz`.

The `display-name-wrapper-exemption-lost` mutant removes the wrapper exemption.
It compiles and runs; Node and emitted JavaScript disagree with Go, and sanitized
native matches mutated Node's **519,170 bytes**, so native also disagrees with Go.
See `evidence/mutant.log.gz`; process wall 62.858 seconds.

`testdata/wrapper-order.tsx.txt` names the declaration-order boundary: an assignment
before a wrapper binding reports even though running that source would hit the
temporal dead zone. An assignment after the declaration is exempt. The port
matches Go rather than extending the exemption. The absent-node capture includes
the deliberately malformed uninitialized binding pattern; findings-only recovery
is transported by the existing harness, with the same Go output. No different
Go behavior or missing shared helper was found.

Commands, test output redirected to files:

```
go run ./cmd/lint-registry
python3 stage1/cohere/lint/rules/structure-react-component-no-display-name/testdata/check.py
python3 stage1/cohere/lint/rules/structure-react-component-no-display-name/testdata/check.py mutant
python3 stage1/cohere/lint/rules/structure-react-component-no-display-name/testdata/gate.py
```

The selected and native-canary tests use a rule-local Go overlay; shared test
source is unchanged. The added top-level selected test calls `t.Parallel()`.
Plain logs remain in `/tmp`; completed logs are preserved as deterministic gzip.
The full gate requires committed stage-1 sources, so the source checkpoint is
local until the unit's full evidence is ready. No partial unit is pushed.
