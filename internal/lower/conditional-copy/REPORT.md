Built: only the conditional-copy decision from the evidence branch, using cohere's invariant rule at conditional-expression sites.
Commits: based on main 71d7e491b3c9724f7a0e2ee754592149e7f9790b; evidence branch remains 66ad9bbabf0032981465cb61253636cc34274351.
Commands and outputs: lower, load, filtered uncached oracle and vet pass; 576 stage sources, 299 baseline accepts, zero newly refused.
Mutant: removing RunRule makes the conditional refusal assertion fail with got nil; unguarded native/JavaScript/Node and sanitizer comparison passes.
Not covered: the full repository oracle gate; rows 2 to 4 remain main's local proofs, with no fixes or rule migration here.

| Case | Program | Observed result | Decision |
| --- | --- | --- | --- |
| Conditional fresh copies | `internal/oracle/testdata/type_rules_conditional_copy.a` | Source prints `1\n`. Main accepts. This branch refuses at line 4, column 27 through cohere `adamic/invariant-mutable`. With the invocation removed, the ordinary oracle agrees across release native, ASan/UBSan native, JavaScript, and Node, with no leak finding. | Adopt cohere's conservative refusal, as ruled by @system_adamic. |

The newly refused existing-program list is empty: no stage1/ or stage3/fixtures file/line entries. This is a fresh comparison on this branch, not reused evidence. Both compiler binaries were built here: the before binary from main at 71d7e491 with its original pins, and the after binary from the conditional-only implementation. All 576 executable `.a` and `.ts` sources were compiled with both; declaration-only `.d.ts` files were excluded. [Per-file statuses](evidence/inventory.tsv), [summary](evidence/inventory.txt).

Reproduce with:

```sh
python3 internal/lower/conditional-copy/inventory.py \
  /path/to/main-adamic /path/to/conditional-adamic /path/to/log-directory \
  > /path/to/inventory.log 2>&1
```

The local conditional branch recursion in `freshOrWidened` is replaced by `refuseConditionalCopies`, which invokes cohere and consumes only findings whose source site is a conditional expression, including parentheses. Main's other invariant entry paths remain. The read-only String.raw argument exemption remains. There is no nominal rule invocation. `widened`, including its source-constraint handling, is byte-identical to main, as are `class_inheritance.go`, `proven_relations.go`, and `predicates.go`. The existing parameter-field, Weak-method, and union-callback probes were read from 66ad9bb into scratch only: both compiler binaries refused each with identical diagnostics and locations. [Checks](evidence/unchanged-guards.txt), [scratch probe output](evidence/main-guard-probes.txt). No row 2 to 4 source fixture or guard change was brought over.

Main's cohere pin predates the public RunRule API. The necessary dependency update references cohere 7945d102a6c18dd36adf9114a758ce646e8b2359 and its TypeScript pin d92d9bfee114c80be2c375d72edae966176e3a4f. That checker uses rooted file paths, requiring mechanical API adaptations in the loader, checker bridge, meter, fuzz parser, and one regexp file-name check. No shared SSA changes, copied cohere implementation, stage source edits, prohibited emitter edits, or main oracle-file edits are included.

Validation output was saved directly to logs:

- `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`: pass. `nproc` is 5, cgroup quota 4 CPUs. Node 24.19.0, Go 1.27.1, clang 20.1.8. Timing lines: node 0.056s, Go 0.055s, submodules 0.143s, markdown dependency validation 0.016s, markdown ready 0.154s, clang 0.694s, build 15.953s, deferred tests 16.195s, cache warm 16.198s, total 16.271s. [Setup](evidence/setup.txt).
- `go test ./internal/lower ./internal/load -count=1 -timeout 30m`: pass, lower 65.972s and load 2.036s. The former accepted conditional-copy expectation was removed; other main expectations remain. [Output](evidence/packages.txt).
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestConditionalCopyRefusedByCohere|TestNativeAgreesWithNode/internal/oracle/testdata/(invariance_readonly|library_array_copy|fresh_writes)[.]a$' -v -count=1 -timeout 30m`: pass, 3.189s. Tests the new refusal and existing sound readonly/copy fixtures through the ordinary runtime harness. [Output](evidence/oracle.txt).
- `go vet ./...`: pass. [Output](evidence/vet.txt). `gofmt -l cmd internal` and `git diff --check`: no findings.
- `go test ./bridge/tsgo/checker ./cmd/adamic-meter ./internal/fuzz -count=1 -timeout 30m`: pass, 0.671s, 13.523s, 40.045s. These cover the required shim API adaptations. [Output](evidence/api-tests.txt).
- `python3 internal/lower/conditional-copy/run-mutant.py`: pass. The sole mutant changes the loader wrapper's RunRule invocation to return no findings. `TestConditionalCopyRefusedByCohere` fails with `want cohere conditional refusal at 4:27, got <nil>`; compilation succeeds, so this is not a build-warning catcher. The runner also registers the probe with the normal oracle via `ADAMIC_CONDITIONAL_COPY_ORACLE=1`, removes oracle caches, and runs `TestNativeAgreesWithNode/internal/oracle/testdata/type_rules_conditional_copy.a`: pass, 1.707s. It restores the invocation in a finally block. [Summary](evidence/mutant-summary.txt), [catcher](evidence/mutant-refusal.txt), [unguarded runtime oracle](evidence/unguarded-oracle.txt).

The observation is that this particular conditional copy runs correctly; the decision is to adopt cohere's refusal anyway. The final compiler refuses it before emission, so it has no new accepted-fixture count row. Only codex/shared-ssa-conditional-copy is pushed; 66ad9bb is unchanged and nothing is merged elsewhere.
