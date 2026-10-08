# React string refs port

The rule uses Go cohere at `7945d102a6c18dd36adf9114a758ce646e8b2359` as its oracle.
The duplicate registration check found no matching registration across 2,144 remote refs.

The selected gate matched 65 captured upstream source/rule/options combinations and
three owned witnesses on Go, Node, emitted JavaScript and ASan/UBSan native. Including
selected and all-rule witnesses and inherited inputs, the manifest had 159 rows and
produced 549,791 identical bytes. Findings include exact messages, byte ranges,
fixes, suggestions and fixed source, as serialized by the existing comparator.

The Go defect where `this[refs]` counts as the registry is preserved. Literal-keyed
accesses are silent. The nearest class masks enclosing component classes, while
the ES5 walk continues through function scopes. Both template literal kinds need
`noTemplateLiterals`; the refs-access half needs `checkThisRefs`. There are no fixes
or suggestions in this rule.

The helper adapters project existing parser facts. React class classification,
component bases, the narrow factory predicate and JSX attribute names use shared helpers. JSX projects only the
attribute and its name into an immutable two-node array. A mutable stored JSX
projection triggered a compiler cycle refusal in another consumer; the immutable
local form compiles without changing any helper or other rule.

## Commands and evidence

From the repository root:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh --wasi-sdk
source /workspace/adamic-tools/env.sh
go run ./cmd/lint-registry
bash stage1/cohere/lint/rules/react-no-string-refs/testdata/run-selected.sh
```

Setup timing lines and processor/load measurements are in `evidence/setup.log`
and `evidence/wasi-setup.log`. The successful selected comparison and the compiled
mutant caught on Node and emitted JavaScript are in `evidence/selected.log`.
The full package command and every external input are in `evidence/whole-inputs.json`.
Both profiling variables point to one fresh directory; TypeScript is a clean checkout
at `050880ce59e30b356b686bd3144efe24f875ebc8` and the WASI SDK is version 27.

The complete package run passed with 152 test passes (including subtests), zero
failures and one skip. `TestCheckerBridgeRefusalPending` explicitly awaits
`TSGoError` support in the shared checker bridge; no input-dependent test skipped.
Wall time was 1,118.825 seconds on 5 processors. Load before was 1.03/3.94/3.68;
load after was 5.14/5.48/4.65. See `evidence/whole-summary.json` for every completed
test, its duration, the skip reason and the complete command environment;
`evidence/whole.log` preserves the package output. The entire package ran once.
No rule helper or language blocker remains.
