# Verification: f95deb36, with replacement focused tests at 3ac38c67

Original head: `f95deb3618b7f5e3ad508ea29b43a8c5d6f88fb1`.
Replacement A–C head: `3ac38c67f2745703f82644d9116240a03ec48e64`.
Admission base: `ff41896cee93f832a4a8fa68c4c1822ddef5885e`.

The original A–C command had started before the user supplied 3ac38c67. It was allowed to finish and restore exactly, then A–C were rerun at the replacement. D/E stayed at the originally requested f95deb36, as stated in commentary before running them. The only change from f95deb36 to 3ac38c67 is three added lines in generic_method_view_test.go; compiler implementation is identical. [Replacement A–C report](../3ac38c67/README.md).

| Run | Exit | Result |
|---|---:|---|
| Original A, gofmt / vet | 0 / 0 | Formatting output empty; vet passes. |
| Original B, combined pins | 1 | Both chdir rows fail at load with TS2591, Cannot find name process. NodeWatch passes. |
| Original C, remove source-generic guard | 1 | Same two rows red with the same load error; not a discriminating mutant. |
| Original restoration / Git diff | 0 / 0 | Restored exactly. |
| Replacement A, gofmt / vet | 0 / 0 | Both pass. |
| Replacement B, combined pins | 1 | Chdir and exit each fail for .a/.ts with TS2591; NodeWatch passes. |
| Replacement C mutant | 1 | Same four rows red; baseline already red, so guard not proven. |
| Replacement restoration / Git diff | 0 / 0 | Restored exactly. |
| D whole lower at f95deb36, timeout 40m | 1 | Completed in 637.561s without timeout; only the two chdir rows fail. |
| E manifest | 0 | Generated at exact f95deb36. |
| E admission vs ff41896c | 0 | Complete-corpus PASS: 4,957 programs, zero omissions, zero newly admitted, **zero new refusals**. |

## Commands and environment

Go 1.27.1, clang 20.1.8 and Node 24.19.0 are selected with `/workspace/adamic-tools/env.sh`. The isolated source worktree is /tmp/adamic-b3; the evidence worktree stays on evidence/compiler/generic-method-view.

```sh
gofmt -l internal/lower/
go vet ./internal/lower/
go test -count=1 -run 'TestAGenericMethod|TestNodeWatch' -v ./internal/lower/
# Remove exactly " && l.generic(source)":
go test -count=1 -run 'TestAGenericMethod' -v ./internal/lower/
# Restore exactly. Repeat A–C at replacement; restore; switch to f95deb36 for D/E:
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-typescript-source go test -count=1 -timeout 40m -skip '^TestProduct_' ./internal/lower/
python3 cloud/admission-corpus/manifest.py --sha HEAD > /workspace/adamic/evidence/f95deb36/manifest.json
go run ./cmd/adamic-admission-delta --base ff41896cee93f832a4a8fa68c4c1822ddef5885e --head f95deb3618b7f5e3ad508ea29b43a8c5d6f88fb1 --manifest /workspace/adamic/evidence/f95deb36/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --json
```

All numbered outputs/exits and exact mutant patches are retained. The TypeScript source is the verified upstream 050880ce59e30b356b686bd3144efe24f875ebc8 pin used by the preceding passing census checks. No compiler/test code was changed except the requested exactly restored mutants. No test repair or load option change was made.

## Failure details

Full original baseline chdir failures:

```text
    generic_method_view_test.go:69: Load: /tmp/adamic-gate/TestAGenericMethodSeenThroughAnotherViewIsNotYeta_host_method_a3404346504/001/main.a:4:59: error TS2591: Cannot find name 'process'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig.
    generic_method_view_test.go:69: Load: /tmp/adamic-gate/TestAGenericMethodSeenThroughAnotherViewIsNotYeta_host_method_a919081966/001/main.ts:4:59: error TS2591: Cannot find name 'process'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig.
```

The full lower log records only the two chdir rows, plus their parent, as failing. Every `--- FAIL` is indexed in lower-failure-index.log, with full messages in lower-failures.log. No timeout occurred.

## Admission

[New refusals](new-refusals.md): none. Complete, non-void pass, 4,957 programs, zero omitted and zero newly admitted. Runtime results: 13 Node agreements, 50 accepted negative witnesses and 37 ruled runtime stops. The manifest does not turn inline Go-test snippets into standalone program entries; the admission pass therefore does not cure the focused proof's load failure.

npm provisioning used the prior proven temporary ~/.npmrc cafile=/etc/ssl/certs/ca-certificates.crt, preserving TLS verification. The original configuration was restored and checked. Only evidence files are committed.
