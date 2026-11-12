# Mapper readings — task 68x1t2y

Setter: `(*lowering).instantiateFunctionSignature` at `internal/lower/generic.go:148` installs `l.typeMapper = l.instanceMapper(declaration, concreteTypes)`; caller units are `flattenDestructuringBinding` (`destructuring.ts:244:1`) and `makeArrayBindingPattern` (`destructuring.ts:600:1`).

Same measured main **219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205**, same adapted tree (all 747 hashes reverified), same workspace/toolchain and existing instrumented overlay. This follow-up adds logging only. Both parameter readings report **`l.typeMapper != nil` = true**. The mapper pointer at each reading equals the directly logged pointer installed by the setter for that exact caller unit/call.

## assertNode overload 1

Caller unit: **flattenDestructuringBinding** at `/workspace/orm-adapted/src/compiler/transformers/destructuring.ts:244:1`. Call site: `/workspace/orm-adapted/src/compiler/transformers/destructuring.ts:319:9`.

`l.typeMapper != nil`: **true**. Mapper address in both setter and PARAMETER records: `0x218e9a3e5a00`.

Mapped `given`:

```ts
(node: ArrayBindingPattern | ArrayLiteralExpression | ElementAccessExpression | Identifier | ObjectBindingPattern | ObjectLiteralExpression | OmittedExpression | PropertyAccessExpression) => node is BindingName
```

Implementation `takes`:

```ts
((node: Node) => boolean) | undefined
```

Setter evidence (at the assignment, after installing the mapper):

```text
function=instantiateFunctionSignature original=generic.go:148 mapper=0x218e9a3e5a00
unit=/workspace/orm-adapted/src/compiler/transformers/destructuring.ts:244:1
call=/workspace/orm-adapted/src/compiler/transformers/destructuring.ts:319:9
generic_declaration=/workspace/orm-adapted/src/compiler/debug.ts:281:5
resolved_declaration=/workspace/orm-adapted/src/compiler/debug.ts:281:5
T -> ArrayBindingPattern | ArrayLiteralExpression | ElementAccessExpression | Identifier | ObjectBindingPattern | ObjectLiteralExpression | OmittedExpression | PropertyAccessExpression
U -> BindingName
```

Top **15 frames** of `runtime/debug.Stack()` called directly at the requested PARAMETER log point (including debug.Stack itself):

```text
goroutine 1 [running]:
runtime/debug.Stack()
	runtime/debug/stack.go:26 +0x5e
github.com/system-inc/adamic/internal/lower.(*lowering).censusOverload(0x218e98988308, 0x218e87ce5968, 0x218e87ce5808, 0x1)
	github.com/system-inc/adamic/internal/lower/census_small.go:236 +0x78a
github.com/system-inc/adamic/internal/lower.(*lowering).censusOverloads(0x218e98988308, 0x218e87ce5968)
	github.com/system-inc/adamic/internal/lower/census_small.go:154 +0x10f
github.com/system-inc/adamic/internal/lower.(*lowering).latentMainSignatureRaw(0x218e98988308, 0x1fa4, 0x218e87ce5968, 0xffffffffffffffff)
	github.com/system-inc/adamic/internal/lower/functions.go:83 +0x73
github.com/system-inc/adamic/internal/lower.(*lowering).latentSignatureReturnRaw(...)
	github.com/system-inc/adamic/internal/lower/functions.go:486
github.com/system-inc/adamic/internal/lower.(*lowering).latentSpecSignatureReturn(0x218e98988308, 0x1fa4, 0x218e87ce5968, 0xffffffffffffffff, 0x0)
	github.com/system-inc/adamic/internal/lower/latent_speculative.go:150 +0x168
github.com/system-inc/adamic/internal/lower.(*lowering).signatureReturn(...)
	github.com/system-inc/adamic/internal/lower/functions.go:490
github.com/system-inc/adamic/internal/lower.(*lowering).signature(...)
	github.com/system-inc/adamic/internal/lower/functions.go:482
github.com/system-inc/adamic/internal/lower.(*lowering).lowerFunction(0x218e98988308, 0x1fa4, 0x218e87ce5808, 0xffffffffffffffff)
	github.com/system-inc/adamic/internal/lower/functions.go:60 +0x2c5
github.com/system-inc/adamic/internal/lower.(*lowering).instantiateFunctionSignature(0x218e98988308, 0x218e8eb71f08, 0x218e87ce5808, 0x218e9614e800)
	github.com/system-inc/adamic/internal/lower/generic.go:176 +0x1938
github.com/system-inc/adamic/internal/lower.(*lowering).instantiateFunction(0x218e98988308, 0x218e8eb71f08, 0x218e87ce5808)
	github.com/system-inc/adamic/internal/lower/generic.go:27 +0x45
github.com/system-inc/adamic/internal/lower.(*lowering).call(0x218e98988308, 0x218e8eb71f08)
	github.com/system-inc/adamic/internal/lower/expression.go:1456 +0x413
github.com/system-inc/adamic/internal/lower.(*lowering).callOrMethod(0x218e98988308, 0x218e8eb71f08)
	github.com/system-inc/adamic/internal/lower/class.go:349 +0x596
github.com/system-inc/adamic/internal/lower.(*lowering).expressionStatement(0x218e98988308, 0x218e8eb71f08)
	github.com/system-inc/adamic/internal/lower/statements.go:251 +0xcdd
github.com/system-inc/adamic/internal/lower.(*lowering).latentStatementRaw(0x218e98988308, 0x218e8ebc4348)
	github.com/system-inc/adamic/internal/lower/statements.go:88 +0x396
```

## assertEachNode overload 4

Caller unit: **makeArrayBindingPattern** at `/workspace/orm-adapted/src/compiler/transformers/destructuring.ts:600:1`. Call site: `/workspace/orm-adapted/src/compiler/transformers/destructuring.ts:601:5`.

`l.typeMapper != nil`: **true**. Mapper address in both setter and PARAMETER records: `0x218e95d9abc0`.

Mapped `given`:

```ts
(node: BindingOrAssignmentElement) => node is ArrayBindingElement
```

Implementation `takes`:

```ts
((node: Node) => boolean) | undefined
```

Setter evidence (at the assignment, after installing the mapper):

```text
function=instantiateFunctionSignature original=generic.go:148 mapper=0x218e95d9abc0
unit=/workspace/orm-adapted/src/compiler/transformers/destructuring.ts:600:1
call=/workspace/orm-adapted/src/compiler/transformers/destructuring.ts:601:5
generic_declaration=/workspace/orm-adapted/src/compiler/debug.ts:259:5
resolved_declaration=/workspace/orm-adapted/src/compiler/debug.ts:257:5
T -> BindingOrAssignmentElement
U -> ArrayBindingElement
```

Top **15 frames** of `runtime/debug.Stack()` called directly at the requested PARAMETER log point (including debug.Stack itself):

```text
goroutine 1 [running]:
runtime/debug.Stack()
	runtime/debug/stack.go:26 +0x5e
github.com/system-inc/adamic/internal/lower.(*lowering).censusOverload(0x218e99de4e08, 0x218e87cdf708, 0x218e87cdf5a8, 0x4)
	github.com/system-inc/adamic/internal/lower/census_small.go:236 +0x78a
github.com/system-inc/adamic/internal/lower.(*lowering).censusOverloads(0x218e99de4e08, 0x218e87cdf708)
	github.com/system-inc/adamic/internal/lower/census_small.go:154 +0x10f
github.com/system-inc/adamic/internal/lower.(*lowering).latentMainSignatureRaw(0x218e99de4e08, 0x1fa1, 0x218e87cdf708, 0xffffffffffffffff)
	github.com/system-inc/adamic/internal/lower/functions.go:83 +0x73
github.com/system-inc/adamic/internal/lower.(*lowering).latentSignatureReturnRaw(...)
	github.com/system-inc/adamic/internal/lower/functions.go:486
github.com/system-inc/adamic/internal/lower.(*lowering).latentSpecSignatureReturn(0x218e99de4e08, 0x1fa1, 0x218e87cdf708, 0xffffffffffffffff, 0x0)
	github.com/system-inc/adamic/internal/lower/latent_speculative.go:150 +0x168
github.com/system-inc/adamic/internal/lower.(*lowering).signatureReturn(...)
	github.com/system-inc/adamic/internal/lower/functions.go:490
github.com/system-inc/adamic/internal/lower.(*lowering).signature(...)
	github.com/system-inc/adamic/internal/lower/functions.go:482
github.com/system-inc/adamic/internal/lower.(*lowering).lowerFunction(0x218e99de4e08, 0x1fa1, 0x218e87cdf5a8, 0xffffffffffffffff)
	github.com/system-inc/adamic/internal/lower/functions.go:60 +0x2c5
github.com/system-inc/adamic/internal/lower.(*lowering).instantiateFunctionSignature(0x218e99de4e08, 0x218e8ebe8368, 0x218e87cdf5a8, 0x218e9614e180)
	github.com/system-inc/adamic/internal/lower/generic.go:176 +0x1938
github.com/system-inc/adamic/internal/lower.(*lowering).instantiateFunction(0x218e99de4e08, 0x218e8ebe8368, 0x218e87cdf5a8)
	github.com/system-inc/adamic/internal/lower/generic.go:27 +0x45
github.com/system-inc/adamic/internal/lower.(*lowering).call(0x218e99de4e08, 0x218e8ebe8368)
	github.com/system-inc/adamic/internal/lower/expression.go:1456 +0x413
github.com/system-inc/adamic/internal/lower.(*lowering).callOrMethod(0x218e99de4e08, 0x218e8ebe8368)
	github.com/system-inc/adamic/internal/lower/class.go:349 +0x596
github.com/system-inc/adamic/internal/lower.(*lowering).expressionStatement(0x218e99de4e08, 0x218e8ebe8368)
	github.com/system-inc/adamic/internal/lower/statements.go:251 +0xcdd
github.com/system-inc/adamic/internal/lower.(*lowering).latentStatementRaw(0x218e99de4e08, 0x218e8eb74308)
	github.com/system-inc/adamic/internal/lower/statements.go:88 +0x396
```

## Road into the mapped check

The observed stack road, from the call toward the check, is:

`call` → `instantiateFunction` → `instantiateFunctionSignature` → `lowerFunction` → the measurement overlay’s signature/return wrappers (`signature`, `signatureReturn`, `latentSpecSignatureReturn`, `latentSignatureReturnRaw`, `latentMainSignatureRaw`) → `censusOverloads` → `censusOverload`.

The stack is captured while the installer frame is waiting in lowerFunction. Its displayed scratch `generic.go:176` is the call corresponding to original main `generic.go:165`; it is not the assignment line. The separate SETTER record observes the assignment itself at original **generic.go:148** (scratch **mapper-generic.go.txt:150**). Logging/import insertions explain the scratch line differences. The top-15 stack is retained verbatim rather than rewriting it into production line numbers.

In main, instantiateFunctionSignature builds an instance mapper from the call’s inferred concrete types, sets `l.typeMapper` at :148, and calls lowerFunction at :165. lowerFunction substitutes the overload set’s implementation (:49–51); signature validation then calls censusOverloads (original functions.go:68, reached through the retained measurement wrappers).

Both assertion implementations have **zero implementation type parameters**. censusOverload saves the incoming mapper (`census_small.go:163`) and only installs its own binder mapper inside `if len(servedTypes) > 0` (:165). That branch is skipped here, and there is no reset to nil on entry. Its parameter check therefore inherits the instance mapper installed for one particular caller. These readings confirm a call-specific, non-rigid judgment despite censusOverload’s rigid-signature comment.

For assertEachNode, the setter names generic declaration **debug.ts:259:5 (overload 4)** while its resolved declaration is **debug.ts:257:5 (overload 2)**. The retained latent registration writes generic declarations to `l.generics[symbol]` in declaration order, so the last generic declaration wins (latent_hook.go:166; nested registration repeats the policy in latent_units.go:59). Main expression.go:1456 passes that generic declaration to instantiateFunction. Consequently overload 4’s T/U are mapped from this call before validation walks the entire overload set; earlier overloads can still print their own rigid T. The direct setter reading records T=BindingOrAssignmentElement and U=ArrayBindingElement for the requested call.

Full mode also repeats the assertNode reading while lowering the nested `emitBindingOrAssignment` unit at :318:5, and records three further assertEachNode units (:605, :610, :615). All six parameter readings have non-nil mappers with matching setter observations; `mapper-readings.json` and `mapper-probe.log` retain them all.

## Reproduction and verification

The original REPORT.md, streams and evidence hashes are unchanged. `mapper-instrument.py` layers only the added PARAMETER/stack trace and assignment logging over the original /workspace/dao-probe overlay. Go evidence is `mapper-census_small.go.txt` and `mapper-generic.go.txt`; no production source or adaptation changes. The original probe binary remains intact; the additional logging binary is /workspace/dao-mapper/entry-census.

```sh
source /workspace/adamic-tools/env.sh
scripts=/workspace/op2-report/stage3/census/debug-assert-overloads
python3 -B "$scripts/mapper-instrument.py"
# From /workspace/orm-main:
python3 "$scripts/run.py" "$scripts/mapper-build.log" "$scripts/mapper-build.time.json" go build -buildvcs=false -overlay=/workspace/dao-mapper/overlay.json -o /workspace/dao-mapper/entry-census ./stage3/census/latent/tool
rc=$?; echo "mapper build child_exit=$rc"; exit "$rc"
```

Run `bash "$scripts/mapper-run.sh"` as the next independent command/session. It uses exactly `LATENT_FULL=1`, the adapted destructuring.ts file selector, `LATENT_ASSERT_NO_OUTPUT=1`, and the original adapted src/tsc/tsc.ts entry. It clears inherited mode selectors. Each process is bounded by run.py at 1,740 seconds plus at most ten seconds to kill; wrappers print and propagate their own child status. Then run mapper-report.py through run.py to regenerate this addendum and its validation.

The complete new census JSONL is **exactly equal** to both the original instrumented and original stock full-mode streams, including their relation trailers. There are zero checker diagnostics and the same 26 units. The build and census children exit 0 without timeout. All 15 frames in each of the two requested stacks and all six readings are checked by mapper-report.py.
