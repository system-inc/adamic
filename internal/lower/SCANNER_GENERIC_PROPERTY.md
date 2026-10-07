SHAs: implementation 9754646c037afb6cb0158b7adc6dfb4c74df4713 reflection guard 17c16941a7b192d65e485f7727ac5189ce73d242 and cast guard 4ff0d88d9eca54fe8002fc8d35efdf4c662aadc2, based on origin/main 48c05d09.
Built: generic arrows called through returned object properties, with concrete bodies and counted captures; exact probe goes from refusal to Node agreement.
Commands: focused lower/oracle pass, complete counts update passes, full lowering/oracle and vet/format results recorded below.
Mutants: thirteen source overlays and one wrong-result IR mutant caught; the IR mutant is caught independently in both backends by Node stdout.
Not covered: full native scanner acceptance, every generic function-value shape, recursive nested generic helpers and the full repository test suite.

Branch: codex/scanner-generic-property. The input is
origin/codex/stage3-scanner-proof dc9f8482,
stage3/drivers/scanner/probes/generic-function-property.a, discovered at slice
scanner.ts:555. Its reduced-source comment names upstream scanner.ts:1112.
BLOCKERS.md says this was discovered behind explicit discovery-only stubs;
passing the probe does not prove the validated scanner slice reaches this stop.
No scanner slice, adaptation or proof branch was edited.

The exact probe is copied byte-for-byte into oracle/testdata and registered for
source Node, emitted JavaScript, release native, sanitized native and counts.
On this actual main its initial stop is the nested scanRange declaration at
3:5: `a function inside a function (a closure)`. Source Node prints `x`.
The unit therefore supports the nested generic helper as well as the generic
arrow, instead of replacing either source construct.

A generic arrow is held as a private object containing an ordinary counted
closure per inferred concrete call signature. Its identity is the containing
object's identity; aliases retain that object, and replacement of a property
replaces the object. At a call, the receiver and property are evaluated once,
the concrete closure slot is selected, then arguments are evaluated. This uses
existing IR and both existing backends; no emitter or runtime file changes.
Each body has its own type mapper, substitution and parameter locals, while
captured locals remain the enclosing invocation's counted cells.

Slots use the checker's full identities for the resolved parameters and result,
not merely a native tag, type-parameter spelling or order. Family discovery
collects compatible concrete calls before a runtime function value is made.
A generic call cannot introduce a slot afterward: calls without a source seed
return NotYet. Conversions to ordinary function values or noncallable views,
mixed function unions, optional calls, erased casts, spreads and property reflection remain
diagnostics. typeof still prints function, not the private representation's
object tag. Equivalent generic signatures can share a family; independently
written interface views can still hit the existing generic variance refusal.

Nested generic declarations are registered before lowering their enclosing
body, including declarations following its return. Direct calls instantiate
ordinary closures with that invocation's captured cells. A helper value read
as a first-class value still says NotYet, and recursive nested helpers hit the
finite instantiation limit. Generic arrows made inside another generic
instantiation remain explicitly NotYet. General higher-order family discovery
and structural generic method vtables are not claimed supported by this unit.

The expanded fixture generic_function_properties.a covers number, string and
object returns, original object identity, runtime-built strings, independent
scanner state, inferred aliases, one receiver/callback evaluation and runtime
property replacement. The helper declaration follows return and captures its
factory's label and counter. Node prints:

```text
firstfirst:1
secondsecond:1
0 7
firstfirst:2
wordword
firstfirst:3
true
true false function
firstfirst:4
9 1 1
thirdthird:1
11 true
```

Counts columns allocations/frees/retains/releases/peak/regions:
exact probe 5/5/5/8/5/0; expanded fixture 63/63/58/80/26/0.
Both free every allocation. This implementation pays for a containing object
and a closure per concrete signature, plus a helper closure at each direct
nested call. No performance improvement is claimed.

Mutation evidence on the final code:

| Mutant | What catches it |
| --- | --- |
| Remove both the instance mapper and its substitution entries | Exact probe's Lower diagnostic |
| Allow generic-to-ordinary value conversion, keeping the mutation nil-safe | Negative conversion assertion accepts an unsafe view and fails |
| Permit an unseeded call | Negative higher-order call assertion fails |
| Permit arrow creation under an outer instantiation | Negative outer-instantiation assertion fails |
| Permit mixed generic-function unions | Negative union assertion fails |
| Permit optional generic calls | Negative optional-call assertion fails |
| Omit hoisted generic registration | Expanded fixture's Lower diagnostic at the helper read |
| Collapse all concrete signature slot keys | Number/string/object concrete-body count assertion fails |
| Permit spreading a generic function value | Negative spread assertion fails |
| Copy only global locals into closure instantiations | Expanded fixture's Lower diagnostic at a captured local read |
| Permit observing generic function properties | Negative hasOwnProperty assertion fails |
| Permit a cast erasing the generic function representation | Negative typeof-as-unknown assertion fails |
| Lose typeof's function result | Expanded source-Node oracle catches stdout disagreement |
| Add one to the number property wrapper's returned value | Native and JavaScript finish cleanly; each independently disagrees with source Node stdout |

The source overlays do not edit repository files. Every overlay exits 1 at the
listed checker/test assertion or observation, not clang -Werror. The permanent
wrong-result test requires exit 0 and empty stderr, with stdout as the only
disagreement. The runner is committed at
internal/lower/testdata/run-generic-property-mutants.py. Complete run:
/tmp/generic-property-mutants-land-final.log and per-mutant logs with the same prefix.
The first naive conversion mutation dereferenced nil and was rejected as proof;
the revised mutation removes the guard without causing that unrelated failure.
The first outer-instantiation witness was masked by an independent signature
inference diagnostic; an internal concrete call isolates the tested guard.
Only the final successful catches above are claimed.

Setup: bash cloud/setup.sh completed successfully. Node ready 0.020s, Go ready
0.022s, Markdown dependencies ready 0.070s (validated bytes already installed),
clang ready 0.155s, submodules ready 4.125s, go build ready 183.269s, test binaries
deferred 183.396s, cache warm 183.397s, done 183.427s. nproc 5;
cpu.max 400000 100000. Go 1.27.1, clang 20.1.8, Node 24.19.0. Every build/test
shell sourced /workspace/adamic-tools/env.sh. Log:
/tmp/generic-property-setup.log. An initial recursive fetch stalled in the
TypeScript submodule; its owned fetch processes were terminated, and fetching
with fetch.recurseSubmodules=false succeeded. Setup then updated the pinned
submodules normally. No setup failure or toolchain workaround was needed.

Commands, with test output redirected directly to log files:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run 'TestGenericProperty|TestGenericFunctionPropertyMutant|TestNativeAgreesWithNode/internal/oracle/testdata/(generic-function-property|generic_function_properties)' -count=1 -v -timeout 30m > /tmp/generic-property-focus-complete.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/generic-property-counts-complete.log 2>&1
ADAMIC_GATE_UNCACHED=1 python3 internal/lower/testdata/run-generic-property-mutants.py > /tmp/generic-property-mutants-land-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -count=1 -timeout 30m > /tmp/generic-property-gate-land.log 2>&1
go vet ./... > /tmp/generic-property-vet-land.log 2>&1
gofmt -l cmd internal > /tmp/generic-property-format-land.log
```

Focused lowering passes 0.926s, oracle 0.613s; counts update passes 27.345s.
Original expanded Node output is /tmp/generic-property-source-node.log.
Intermediate gates are retained in /tmp/generic-property-gate*.log. A negative
witness initially expected refusal for a call that was legitimately seeded by
a compatible top-level generic call; it was changed to isolate an unseeded
family. A gate run spanning the later fixture/hoisting change is not evidence
for the final artifact. Final gate results follow below.

No full scanner binary or fresh census recount was attempted. No code was
copied from cohere, no protected orchestration/oracle file was edited, no main
or area branch was changed, and no pull request was opened.

Final artifact gate exits 0: internal/lower 29.000s and the complete uncached
internal/oracle suite 167.144s. Vet and formatting logs are empty; git diff
--check passes. The final source-overlay runner exits 0 with all thirteen
independent catches, including erased-cast and typeof. The erased-cast overlay
was adjusted to keep its temporary variable used, so a Go build failure is not
mistaken for a successful mutation check. The permanent wrong-result test also
passes in the complete oracle, independently checking native and JavaScript
stdout against Node with clean exits and empty sanitizer stderr.

The full repository go test ./... gate was not run under the worker exception;
the complete touched lowering package and complete uncached oracle were run.
The latest main fetch remains the branch base 48c05d09. Final report and the
reviewable code are pushed only to codex/scanner-generic-property.
