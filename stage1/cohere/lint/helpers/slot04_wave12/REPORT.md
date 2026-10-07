# Slot 04 wave 12

Built `isQuotedLiteral`, `replaceValueNode`, and `unionNodeSets`, one `.a` module each. Claim f44e112 was pushed before implementation. All prior slot 04 work was landing-ready on origin/main f8013f0baac41ddc340d76f83bddde38536a8f07, with the landing verification recorded in 623987c. A final fetch left that main revision unchanged. All 18 origin codex/lint-helpers* claim trees were checked again; no competing claim for these three was found.

## Observations

Actual pinned Go cohere helpers are invoked through a temporary overlay. No shared harness, generator, compiler, or rule files were edited. The value adapter uses Go ParseValue to produce node views; the Adamic replacement callback uses the already verified slot04_wave8 writeValueCss module. Replacement changes kind before invoking the callback, then clears children and their nil-presence flag. Union preserves the existing map object on either empty-input fast path and promotes every key to true only on its fresh-map path, including false-valued keys and nil-pointer keys represented by identity IDs.

`python3 testdata/capture.py` captured 112 distinct asserted source fixtures from all four consumers. The actual Go tailwind suite passed in 0.392s. Instrumented helper entry points produced four calls, three distinct input records, for quote/replacement; union had no recorded calls. Replacement capture serializes values through Go ValueToCss and reparses them, so it does not preserve arbitrary original node kinds or pointer alias geometry. Separate self/child alias controls preserve those relationships explicitly. Union controls cover nil and allocated empty maps, shared maps, overlapping keys, false values, nil keys, and result mutation to observe map aliasing.

`go test -count=1 -v ./stage1/cohere/lint/helpers/slot04_wave12` passed in 11.559s. Actual Go, source Node, sanitized native, and emitted JavaScript agree on 154 observations from 113 controls, 224 observations from consumer sources, six observations from captured records, and all 65,793 raw byte strings of length zero through two for the quote predicate. The consumer sources are parsed as value input here, rather than being full Adamic rule executions. Longer quote inputs include Unicode, invalid bytes, NUL, mismatched quotes, and a 2048-byte control. `go vet ./stage1/cohere/lint/helpers/slot04_wave12` and `git diff --check` passed. Logs are under evidence/; an initial oracle-driver boolean logging type error was fixed before the successful run.

Six semantic mutants compiled and ran successfully in all three Adamic execution modes, then differed from Go: quote minimum length changed to one; double quotes rejected; replacement kind changed to function; replacement retained children; empty-left union returned left; fresh union retained false values from left. Four omission mutants, one per consumer, failed the coverage check.

Inherited setup: Go 0s, clang 1s, Node 1s, submodules 2s, warm 165s, total 165s; nproc 5. The tool environment is /workspace/adamic-tools/env.sh.

## Readiness inference

Each helper removes one prerequisite for each of:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

That is twelve prerequisite removals across four distinct rules, not four completed rules. The frozen ledger lists 264, 201, 221, and 204 remaining helpers respectively before this batch. No rule reaches zero remaining dependencies from this batch alone.

Not covered: complete Adamic lint diagnostics/fixes/suggestions, a complete native Go pointer model, arbitrary cyclic value graphs, nil replacement targets, or exhaustive longer value-parser inputs. Union runtime reachability in those existing Go consumer fixtures was not demonstrated; its helper semantics are exercised directly. The full repository gate was not rerun; prior helper packages were already green on this unchanged main base.
