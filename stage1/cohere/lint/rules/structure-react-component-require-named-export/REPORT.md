# structure/react-component-require-named-export

Ported the pinned Go rule, using the existing FileContextFor, IsLikelyComponentName,
IsExportedByName, IsLikelyReactComponent and JSX-return helpers. Rule-local collection
preserves source order, duplicate-name upgrades and export-specifier aliases. The
message catalog slice comes from helpers/testdata/catalog.json, exported from Go's
resolved policy catalog. No options, fixes or suggestions are supplied upstream.

## Observations and limits

Go only considers top-level declarations. Hooks alone and returns inside an if do
not identify a component; a first parameter named props or properties does. An
export list preceding its declaration does not mark that later component exported.
The order.tsx witness pins this behavior. Any property call named forwardRef counts,
regardless of its object. These observations are matched, without widening detection.
Agreement covers captured upstream cases, witnesses and inherited corpus; it is not
an exhaustive proof over all TypeScript input or arbitrary parser recovery.

## Validation

Cohere pin: 7945d102a6c18dd36adf9114a758ce646e8b2359.
Clean TypeScript source pin: 050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3).
WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot.
Both ADAMIC_LINT_PROFILE_DIR and ADAMIC_LINT_PROFILE_SNAPSHOTS used one fresh
mktemp directory; ADAMIC_LINT_BENCH=1; ADAMIC_LINT_RULES unset.

Commands, with stdout/stderr redirected directly to logs:

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
go run ./cmd/lint-registry
go test -overlay=/tmp/s13-named-export-overlay.json ./stage1/cohere/lint -run '^TestRequireNamedExportSelected$|^TestMutants$/^default-specifier-counts-as-named$' -count=1 -v -timeout=30m
go test -overlay=/tmp/s13-named-export-overlay.json ./stage1/cohere/lint -count=1 -json -timeout=30m
```

The overlay maps stage1/cohere/lint/require_named_export_selected_test.go to this
rule's testdata/selected_test.go; the test calls t.Parallel. Selected comparison:
32 captured upstream cases, five witnesses (selected and all), inherited generated
all-rule rows; Go, source Node, emitted JavaScript and ASan/UBSan native identical,
533987 bytes. Selected tests passed in 106.068s.

Mutant default-specifier-counts-as-named removes the default-export exemption from
export-list marking. It compiled and ran; Node and emitted JavaScript disagreed
with Go on case 88 (default-specifier.tsx). See evidence/selected.log:11 and :23.
The full gate also caught it. This mutant is not claimed as a native semantic run;
the unmutated rule's native comparison passed, and the full harness runs its native
canary independently.

Whole package: exit 0, 160 passed, 0 failed, 1 skipped (counts include subtests).
Only skip: TestCheckerBridgeRefusalPending, awaiting codex/tsgo-errors-as-values
and TSGoError from the C error buffer. No optional input-dependent skips.
Wall time 1259.954s; nproc 5; load before 0.89/0.77/1.31,
after 4.86/5.69/4.44. Compiler/stage1 corpus identical: 31884100 bytes.
Setup total 0.987s; go 0.022s, node 0.020s, submodules 0.056s,
markdown dependencies 0.066s, clang 0.151s, build cache warm 0.959s.
Full event log: evidence/whole.jsonl. Setup and registry logs are alongside it.

No shared files, upstream bodies, compiler files or area branches were changed.
