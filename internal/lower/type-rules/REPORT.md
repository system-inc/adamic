Built: cohere RunRule integration, thirteen behavior probes, one sound control, and retained guards for witnessed failures.
Commits: implementation 8ee0c458d35295638edc4572b5997ec76efac89c; main merge 5bcbb05fb2667221f6040e654e30bfdbb02cca2d.
Commands and outputs: affected-package tests and go vet pass; 576 stage sources compared, 299 baseline accepts, zero newly refused; setup 69.165s, nproc 5.
Mutants: dropping either cohere invocation fails refusal assertions; dropping constrained-source protection triggers ASan; dropping Weak or union protection changes stdout.
Not covered: a full repository uncached gate or a general proof of cohere's rules; complete deletion of local pair proofs would leave observed failures unguarded.

| Case | Programs in internal/oracle/testdata | Result with ordinary cohere checks and local guards removed | Decision |
| --- | --- | --- | --- |
| 1. Conditional fresh copies | `type_rules_conditional_copy.a`; all executable sources in stage1/ and stage3/fixtures | Synthetic safe copy changes Accepted to Refused at line 4, column 27. No existing program becomes refused. | Adopt cohere `adamic/invariant-mutable`; conservatively refuse this conditional copy. |
| 2. Constrained parameter widening | `type_rules_parameter_min.a`, `type_rules_parameter_array.a`, `type_rules_parameter_map.a`, `type_rules_parameter_field.a` | Minimal declaration accepts and prints `accepted`. Each store/read variant prints `undefined` on Node through a declared string read; native ASan reports heap-buffer-overflow in all three. | Retain Adamic `adamic/invariant-constraint-view`; newly accepted case is unsound. |
| 3. Structural Weak<A> | `type_rules_weak_min.a`, `type_rules_weak_field.a`, `type_rules_weak_method.a`, `type_rules_weak_callback.a` | Minimal and field variant agree with Node. Method: Node 9, native 3. Callback replacement: Node 8, native 2. JavaScript backend exits 70 on both method variants. | Retain Adamic `adamic/native-nominal-view`; structural objects cannot be treated as native class instances. |
| 4. Structural A \| B | `type_rules_union_min.a`, `type_rules_union_field.a`, `type_rules_union_method.a`, `type_rules_union_callback.a` | Minimal and field variant agree with Node. Method: Node 9, native 4. Callback replacement: Node 4, native 2. JavaScript backend exits 70 on both method variants. | Retain Adamic `adamic/native-nominal-view`; structural union values do not establish native class identity. |

Observed outputs come from the normal differential oracle, including release native, sanitizers, JavaScript, and Node: [unguarded experiment](evidence/unguarded-oracle.txt). The inference is that cases 2, 3, and 4 cannot safely be accepted by this compiler. Passing field-only variants do not establish that class method dispatch is sound. Final compilation refuses every negative probe with the scoped rule listed above; Node outputs and final diagnostics are recorded in [focused oracle evidence](evidence/oracle-focused.txt).

The constrained-source case means `Pack extends Dog[]` viewed as `Animal[]`, or equivalent Map/object constraints. It does not mean widening to another unconstrained destination parameter: cohere already refuses that different relation. Union probes construct the structural literal directly into the union, then introduce a structural alias; a pre-existing structural variable assigned into the union is already refused by cohere.

The existing-program refusal list for case 1 is empty, so there are no file/line entries. The inventory covers 576 `.a` and `.ts` sources under both requested directories, excluding declaration-only `.d.ts` files. Each was compiled before and after the draft switch: [complete draft inventory](evidence/inventory-draft.tsv). All 299 baseline accepts were then recompiled with the final guards: [final inventory](evidence/inventory-final.tsv), [summary](evidence/inventory-final.txt). Baseline was the main-merged shared branch, not an unrelated main build.

Ordinary relation checks now use `rule_runner.RunRule` through `internal/load/rules.go`, referencing the cohere submodule at `7945d102a6c18dd36adf9114a758ce646e8b2359`. No cohere code was copied. Local assignment/argument mirror paths were removed. Existing read-only `String.raw` behavior remains exempt at its intrinsic site. Local pair helpers remain for the demonstrated constrained and native nominal holes, inferred views (`adamic/invariant-inferred-view`), explicit proven relations (`adamic/proven-relation`), instantiated generic checks, and override ABI checks. RunRule exposes file checks, not the pair-proof API those consumers require. Deleting these helpers completely is not sound within this unit; the conservative assumption is documented in the implementation commit.

The merge includes main `71d7e491b3c9724f7a0e2ee754592149e7f9790b` on shared base `bafac299b7c1c1c4b1bb2ea0106c187269ad056a`. The fuzz source parser needed the same rooted-path shim API adjustment as the merged parser fixtures. Local module replacements resolve cohere and its pinned TypeScript modules; the initial tidy failure on nonexistent `tsc/v0.0.0` was resolved by the local underlying-module replacement.

Suggested cohere lint tests follow. These are the actual oracle programs, not hypothetical examples.

`type_rules_parameter_field.a`: the assignment to `animals` must preserve the source constraint's writable slots. Node prints `undefined` from `dogs.pet.bark`, declared string; unguarded native fails ASan.

```a
interface Animal { readonly name: string; }
interface Dog extends Animal { readonly bark: string; }
function store<Pack extends { pet: Dog }>(dogs: Pack): void {
 const animals: { pet: Animal } = dogs;
 animals.pet = { name: 'cat' };
 console.log(dogs.pet.bark);
}
store({ pet: { name: 'old', bark: 'woof' } });
```

`type_rules_weak_method.a`: structural-to-Weak assignment loses the distinction between an own callback and a nominal class method. Node prints 9, native prints 3; JavaScript fails while looking up class methods.

```a
import type { Weak } from 'adamic';
class A { value: number = 1; read(): number { return this.value; } }
const object: { value: number; read: () => number } = { value: 2, read: () => 9 };
const weak: Weak<A> = object;
object.value = 3;
function read(value: Weak<A>): number { return value === undefined ? -1 : value.read(); }
console.log(`${read(weak)}`);
```

`type_rules_union_callback.a`: an aliased callback update remains visible in Node, while native dispatches the B method. Node prints 4, native prints 2; JavaScript fails while looking up class methods.

```a
class A { read(): number { return 1; } }
class B { read(): number { return 2; } }
const view: A | B = { read: () => 3 };
const object: { read: () => number } = view;
object.read = () => 4;
function read(value: A | B): number {
 if (value instanceof A) { return value.read(); }
 return value.read();
}
console.log(`${read(view)}`);
```

Validation commands, with output saved directly to logs:

- `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`: success. Node 24.19.0, Go 1.27.1, clang 20.1.8; `nproc` 5, cgroup quota 4 CPUs. Setup timing lines: node 0.073s, Go 0.102s, submodules 0.129s, clang 0.579s, markdown install 1.042s, markdown ready 1.255s, build 68.941s, tests deferred 69.053s, cache warm 69.056s, total 69.165s. [Full setup output](evidence/setup.txt).
- Unguarded draft: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/type_rules_' -v -count=1 -timeout 30m`. All new probes were temporarily registered with the ordinary oracle; failures are the observations in the table. The final test registration uses `ADAMIC_TYPE_RULE_PROBES=1` to reproduce individual unguarded experiments through mutants.
- Final affected-package gate: `go test ./internal/lower ./internal/load ./internal/fuzz ./internal/oracle -count=1 -timeout 30m`. All pass: lower 59.428s, load 3.103s, fuzz 56.443s, oracle 74.033s. This uses the normal worker build cache. [Output](evidence/packages.txt).
- Uncached focused oracle tests for the new refusal probes, sound control, and arrow-block refusal site pass (11.693s): [output](evidence/oracle-focused.txt). An earlier uncached package attempt completed ordinary comparisons but failed an obsolete arrow-block diagnostic assertion; the assertion now checks cohere's rule/message and retains the exact line 5, column 42 expectation. [Earlier attempt](evidence/uncached-packages-attempt.txt).
- A later package attempt under competing inventory work failed the timing-sensitive interrupt signal probe. Its isolated uncached rerun passes (7.000s), and the subsequent complete affected-package gate passes. [Attempt](evidence/loaded-packages-attempt.txt), [isolated signal rerun](evidence/signals.txt).
- `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts` passes (38.784s). Only the new control row was added; existing counts are unchanged. The control prints `Rex\n2 1\n3\n4\n` on every backend and passes sanitizer/leak checks. Counts: allocations 14, frees 14, retains 21, releases 34, peak live 9, region allocations 0. [Output](evidence/counts.txt).
- `go vet ./...` passes; `gofmt -l cmd internal` and `git diff --check` produce no findings.

Run the five independent mutants with `python3 internal/oracle/testdata/run-type-rule-mutants.py`; it restores every mutated file in a finally block. All five were rerun after the final nominal-target caching change and caught, without build failures: [summary](evidence/mutants.txt).

- Drop the cohere invariant invocation: `TestTypeRuleProbesStayRefused/conditional_copy` reports wanted `adamic/invariant-mutable`, got nil. [Log](evidence/mutant-cohere-invariant.txt).
- Drop the cohere nominal invocation: `TestCohereNominalInvocation` reports wanted cohere nominal refusal, got nil. [Log](evidence/mutant-cohere-nominal.txt).
- Disable the constrained-source guard: ordinary oracle parameter store/read tests report ASan heap-buffer-overflow. [Log](evidence/mutant-constraint-source.txt).
- Disable the retained Weak nominal guard: ordinary oracle method/callback tests report stdout differences and JavaScript failure. [Log](evidence/mutant-weak-nominal.txt).
- Disable the retained class-union nominal guard: ordinary oracle method/callback tests report stdout differences and JavaScript failure. [Log](evidence/mutant-union-nominal.txt).

The scope does not include repairing native structural class dispatch, changing cohere, or proving every retained pair helper with a new independent mutant. It includes mutants for both new cohere invocations and all three acceptance guards exercised by this unit. No prohibited emitter or main oracle files were edited. Only the unit branch is to be pushed; no PR or merge into another branch.
