Built: the owned React correctness check now fails on the known native RegExp blocker; its skip was removed.
Base: existing pushed 98b9e70f4 remains on lint area d65a8f931 and main 39638d9e2; fetched bases were unchanged.
Command/output: go test ./stage1/cohere/typeaware -run '^TestWave03ReactAgreementAndMutants$' -count=1 -v fails, Go exit 1, 7.690s.
Mutant: an overlay-only Fatalf-to-Skipf mutant compiles and exits zero; the required-failure output check rejects its SKIP result.
Uncovered: full React oracle, whole repository gate and the 17 external-input correctness checks; no new claims.

The current instruction requires real findings to remain failures with reproducible evidence. The owned React constructor preflight previously used t.Skipf for a documented native lowering gap. It now uses t.Fatalf. There are no Skip/Skipf calls in the owned wave_03 test files. No shared checks, input requirements or compiler implementation were relaxed or edited. Historical reports recording the earlier skip remain evidence of their earlier runs; they do not certify today's full package gate.

The failure is reproducible after sourcing /workspace/adamic-tools/env.sh:

```sh
go test ./stage1/cohere/typeaware -run '^TestWave03ReactAgreementAndMutants$' -count=1 -v > /tmp/wave03-react.log 2>&1
```

It prints `--- FAIL: TestWave03ReactAgreementAndMutants` and `BLOCKED: required new RegExp(pattern, 'u') is not supported by native lowering`, followed by `stage 0 can't lower RegExp with a nonconstant pattern yet` at stage1/cohere/typeaware/wave_03_react/gaps/general_regex.a:3:24. Go exits 1. This is an actual failed check, not a skip or a passing expected-error oracle.

The minimal owned reproducer is gaps/general_regex.a:

```typescript
import { programArguments } from 'adamic';
const args = programArguments();
console.log(new RegExp(args[0] ?? '^is[A-Z]', 'u').test(args[1] ?? 'enabled') ? 'match' : 'mismatch');
```

Build with `go build -o /tmp/adamic ./cmd/adamic`, then `/tmp/adamic build stage1/cohere/typeaware/wave_03_react/gaps/general_regex.a -o /tmp/regex-option`. It exits 1 with the same refusal. The rule's option helper requires this runtime constructor under the user regex contract; no finite matcher fallback is permitted. Native lowering currently only compiles constant patterns. Closing this requires compiler/runtime support outside this unit's rule territory, so the blocker is named rather than patched in shared files.

The intentional mutant is kept only in a Go overlay under /workspace/wave-03. It replaces the exact owned t.Fatalf call with t.Skipf. The resulting Go test compiles, reports SKIP and exits zero. A strict independent observation check requires the normal failure marker, the exact constructor refusal and no SKIP; it accepts the normal failure and rejects the mutant's suppressed result. Production remains t.Fatalf. Exact logs and the overlay are retained as gzip evidence with hashes; no mutant source is active in the repository.

The twelve active rules' full 631.669s runtime-area validation, their comparison mutants, sanitizer and released-handle checks remain same-baseline evidence because their implementation and runtime did not change this turn. The interpolation literal still has its separate fifteen-control Go/native/JavaScript/sanitizer proof. Neither certifies the blocked full React driver. The branch is not landing-ready as a fully green oracle branch, and no further rules are claimed. This unit stops at the exact named native constructor failure. The whole repository gate and the 17 announced external-input checks were not run; no pass is claimed for them.
