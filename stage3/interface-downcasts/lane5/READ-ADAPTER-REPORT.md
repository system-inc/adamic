Built: reached-read lowering adapter and both backend producer-signature emitters; JavaScript runtime inclusion is wired, shared source-read dispatch is still pending.
Commits: follows f13be3e2; this checkpoint adds the adapters and precise views-integration handoff.
Commands: focused lower/native/JavaScript/oracle tests pass; lower/native/JavaScript vet passes; logs retained.
Mutants: forged producer metadata caught by wrong-arity assertions in each backend; eager descendant traversal caught by unread unknown object-member control.
Not covered: zero certified pairs or reads; no integration tip is published, lazy admission and closure ABI reconciliation remain integrator dependencies.

The user changed coordination: no further individual-lane merges. git ls-remote
origin refs/heads/codex/views-integration returns no tip at this checkpoint.
Fetched lazy5002bfe0 is a planning commit, not lazy admission; closure4526aa46 is
not merged. Its code-pointer ABI differs from this branch's current three-argument
closure code, reinforcing the need for the integrator's reconciled tip.

prepareViewCallableRead is a concrete lower adapter: it removes optional undefined,
interns a callable placeholder if needed, completes its fixed signature at a reached
read and returns the contract id. Argument/result descriptors record representations
without recursively admitting object members. Generic/rest/optional-parameter,
overload, void and other unknown signatures retain conservative refusal/unknown
behavior. It is not invoked at a cast. The shared caller must propagate the error.

Each backend certificate emitter selects recorded metadata by generated closure
code identity using the final IR Function.Parameters/Returns and locals. Expected
metadata comes independently from the read's ViewContract. This needs no closure
layout change or extra calling convention. Unknown code is rejected, even when a
view supplies a plausible signature. Ordinary closures only are covered by the
adapter; Receiver closures, class method thunks and void returns remain unknown.
Native checks heap kind before accessing a possible code pointer. Operands must
already be evaluated; JS snapshots the value once. Native includes the helper
header automatically. Per-read signature table emission can be compacted later.

The adapters are exercised on valid calls, wrong arity, wrong result, optional
undefined and unrecorded code. Native release and sanitizer both check pinned
exit70 messages. Existing source Node/backend controls remain passing. These
adapter harnesses are not corpus certification or proof of lazy source integration.
No same-name scan or other soundness barrier has been removed. Hook instructions
for both shared lowering and backend field paths are in checked-views-plan.md.

Focused command after the adapter work:
go test ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
-run 'TestPrepareViewCallableRead|TestViewCallableProducerCertificate|TestViewCallableShape|TestViewCallablesNode|TestCheckedViewCallables|TestCheckedViewCallableContractControl|TestCheckedViewOpaqueSignature'
-count=1 -timeout10m. Logged result: lower.684s, native3.740s,
JavaScript.756s, oracle.610s, all pass. Vet on three touched packages passes.
The full gate is not rerun; the four proven baseline lower failures from the first
report remain outside these files.

run-read-adapter-mutants.py runs and restores three independent semantic mutations:
native-forge-producer replaces implementation metadata with declared metadata and
TestViewCallableProducerCertificateNative/wrong-arity catches it;
javascript-forge-producer does the same and
TestViewCallableProducerCertificateNode/wrong-arity catches it;
lower-eager-signature-descendants invokes full child contract admission and
TestPrepareViewCallableRead/interface_Result catches the unread unknown descendant
refusal. None is a build-only kill. The earlier 21 shape mutants remain recorded.

After this push: zero certified; all 308 lane2 pairs/1,503 reads and all 2,818
Unknown pairs/11,063 reads remain; 11,648 deduplicated explicit read witnesses.
These are conservative demand inventories, not exact reaching-view counts. Do not
subtract a shape merely because these adapters exist. Re-measure with lazy
admission's exact reachability when the integrator publishes it. Per the user's
coordination instruction, rest after this checkpoint pending that tip.
