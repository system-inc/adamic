Built producer-specific callable-union result decoding and boxing for item 154, independently of views.
Commit: see the delivery SHA reported with this branch; base bfe05533, current origin/main at branch creation.
Validation: focused both-backend fixtures, full lowering, reader guard and counts passed; exact commands below.
Mutant: numeric .reference interpretation caught by Node stdout comparison; zero prints undefined rather than 0, six crashes with empty stdout rather than 6.
Not covered: the full native/oracle packages, full gate, every method/optional-result representation, or new view admission.

# Item 154

`callThrough` now decodes a union-valued invocation using the actual emitted producer's declared result representation. It identifies the code with the existing closure convention metadata (including counted code pointers), reads the correct adamic_value slot, and uses the existing union conversion. Numeric results get number boxes; owned string/reference results transfer ownership into the union. Already-boxed producer results remain boxed. The same boundary handles method thunks without changing the runtime ABI. A missing emitted producer reaches the existing unreachable guard rather than guessing a representation.

The fix is local to native call emission: three lines at the call boundary and a helper. No protected assembly/native.go/lower.go/oracle_test.go file changed. No code was copied from cohere. This lands the callable-union result representation correction toward integration item 154.

## Node evidence and tests

The four `.a` fixtures have source Node outputs `6`, `value:5`, `0`, and the mixed program's `6 / value:5 / 6 / 6`. The mixed fixture exercises both number and string producers through the same callable union, a named function adapter, a producer observing arguments.length, and a producer whose declared return is already number|string. Sanitized native, release native, and JavaScript match Node. Count rows were refreshed; there are four new rows and no changed existing rows, all allocations equal frees.

New top-level test leaf seconds from focused runs: TestCallableUnionResultNumber 0.66, TestCallableUnionResultString 0.60, TestCallableUnionResultZero 0.64, TestCallableUnionResultBoth 9.30, TestCallableUnionResultAgreement 3.19, TestCallableUnionResultReferenceMutant 2.68. Every test uses t.Parallel and is under 60 seconds.

Commands, output redirected to the correspondingly named evidence log:

```
timeout 180 go test ./internal/oracle -run '^TestCallableUnionResult' -count=1 -v -timeout 90s
timeout 180 go test ./internal/lower -run '^TestCallableUnionResult' -count=1 -v -timeout 90s
timeout 300 go test ./internal/lower -count=1 -timeout 240s
timeout 180 go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s
timeout 480 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 6m -args -update-counts
```

Outputs: focused oracle PASS 9.320s; focused lowering PASS 3.356s; full lowering PASS 98.512s; reader guard PASS 50.290s; counts PASS 129.204s. Tests do not read prohibited call-target fields. Lane checks run after commit and before push; their final output is reported with delivery.

## Mutants

The permanent lowering mutant changes emitted producer number boxing to a cast of the same slot's .reference. It changes no source input or runtime. Numeric zero makes the old ABI interpretation observable without invalid pointers: the executable builds successfully and exits 0, printing `undefined\n`, whereas source Node exits 0 printing `0\n`. recordNativeAgreementFailure catches precisely Native backend stdout. This proves the numeric fixture is load-bearing without relying on a sanitizer or clang rejection.

The separate six-valued witness applies the identical mutation to the number fixture: native builds successfully, prints nothing, and exits -11; Node prints `6\n` and exits 0. Generated fixed and mutant C are `.c.txt` evidence. `mutants.json` records commands and observations. An initial standalone clang attempt omitted the runtime archive and failed to link; it is retained explicitly as failed setup evidence and is not counted as a killed mutant. Both final mutants link the matching plain release runtime archive and run. `reference-result.diff` records removal of the new call boundary, which restores the former .reference read for union results.

## Setup and limits

The existing unit setup was refreshed with GOPROXY=https://proxy.golang.org|direct and timeout 600 bash cloud/setup.sh, then /workspace/adamic-tools/env.sh was sourced for Go/clang commands. Setup completed in 41.760s: Go 0.032s, Node 0.034s, submodules 0.095s, markdown 0.105s, clang 0.185s, build 41.313s, cache 41.578s. nproc=5 with a four-CPU quota. Every long command was bounded, and evidence is under review/compiler/fx7-callable-union-result/.

Assumption: this closed program's closure and method code pointers are generated from its IR functions, the same identity premise used by existing producer certification. This change addresses results; argument representation adapters and expanding callable admission remain separate work. No main or area branch is pushed and no unlanded views branch is imported.
