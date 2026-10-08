# JSX helpers parallel subtests, wave one, #g68rb4h

Base: origin/area/stage1-lint, 9156bf5c579a44d687c9955d13e44f9ad8bbb6f8.
Owner/reviewer: @system_cohere_lint_rules. Only TestJsxHelpersAndMutants changed;
no other test was run directly or edited. The target's existing capture invokes
its pinned upstream next/react/structure rule tests as before.

Six existing mutant subtests now call t.Parallel before setup. The full-corpus
baseline stays in the parent; no corpus sharding or output merging changed.

## Shared state and native builds

The parent scratch directory holds the completed cases.json and want.txt; children
only read cases.json and the immutable want byte slice. Parent TempDir cleanup
waits for children. The mutation table and checked-in sources are read-only.
Each child has its own source directory, native binary, emitted JavaScript, IR,
loader program, and run() output files and stderr buffer. No loop-owned map or
slice is written by children, and no manifest/output path is reused for writes.

No package-wide native-build semaphore, channel, or counter exists in the helpers
package; none was invented. TestTopLevelTestsAreParallel is absent. The separate
parent lint package does have nativeBuilds in checker_build_test.go, capped at
min(runtime.NumCPU(), 4); it is private to that test package, is not used by these
helpers, and remains untouched.
The native runtime cache's sync.Map/per-key mutex and atomic directory publication
remain intact, as do the split-build per-unit locks and Jobs limit. This test uses
native.Build with Sanitize: true and no Split/Jobs option; it has no process pool.
Go's -test.parallel bounds the mutant subtests.

## Measurements

Linux amd64, Go 1.27.1, 5 available CPUs (nproc), identical pinned dependencies.
All measurements ran serially. Commands used the tools environment from
/workspace/adamic-tools/env.sh and only this -run selector:

```sh
go test -v ./stage1/cohere/lint/helpers -run '^TestJsxHelpersAndMutants$' -count=1 -timeout=30m -parallel=2
go test -v ./stage1/cohere/lint/helpers -run '^TestJsxHelpersAndMutants$' -count=1 -timeout=30m -parallel=8
# After: same commands with -json in addition to -v.
```

Before: the target's own verbose PASS duration includes its serial children.
After: JSON run-to-pass timestamps include all parallel children; the printed
parent duration excludes waiting for them, so it is shown separately. Compilation
and package startup are excluded. Single measurements, not a statistical benchmark;
the first baseline had colder caches. An interrupted p8 baseline after an environment
restart was discarded and rerun. An initial after-p2 validation passed (package
218.021 s, parent 62.58 s); it was repeated with JSON for precise test timestamps.

| -test.parallel | Before test s | After test wall s | After parent active s | Reduction |
| --- | ---: | ---: | ---: | ---: |
| 2 | 379.23 | 203.09 | 50.87 | 46.4% |
| 8 | 348.43 | 172.76 | 63.17 | 50.4% |

After run/pass events:
```json
{"Time": "2026-10-08T17:55:05.665956378Z", "Action": "run", "Package": "github.com/system-inc/adamic/stage1/cohere/lint/helpers", "Test": "TestJsxHelpersAndMutants"}
{"Time": "2026-10-08T17:58:28.757816085Z", "Action": "pass", "Package": "github.com/system-inc/adamic/stage1/cohere/lint/helpers", "Test": "TestJsxHelpersAndMutants", "Elapsed": 50.87}
{"Time": "2026-10-08T17:58:44.925165977Z", "Action": "run", "Package": "github.com/system-inc/adamic/stage1/cohere/lint/helpers", "Test": "TestJsxHelpersAndMutants"}
{"Time": "2026-10-08T18:01:37.687265116Z", "Action": "pass", "Package": "github.com/system-inc/adamic/stage1/cohere/lint/helpers", "Test": "TestJsxHelpersAndMutants", "Elapsed": 63.17}
```

## Verdict comparison

An exact comparison of case/backend verdict tuples and count lines passed across
before-p2, before-p8, after-p2, and after-p8. Each has six mutants caught on all
three backends (18 verdicts). Unmutated output remains byte-identical to all
514736 Go rows on all three backends via the unchanged bytes.Equal comparison.
The unchanged oracle still appends the planted JSX source, seeds Unicode matcher
controls, and appends the nil query. The unchanged capture asserts 1284 inputs;
its 27 per-consumer count lines also remain identical in the full logs.

The following are the actual verdict/count messages, grouped by case for comparison;
source-line prefixes and concurrent execution order are omitted. Raw verbose logs
and after JSON events are beside this report.

Parallelism 2, before:
```text
captured 1284 distinct upstream inputs from 27 JSX consumers
1285 sources, 23909 nodes (AttributeName/ElementParts), 31640 matcher pairs, 459187 intrinsic/attribute queries
514736 Go output rows byte-identical on Node, emitted JavaScript and sanitized native
TestJsxHelpersAndMutants/attribute_name.a: output mutant caught on Node
TestJsxHelpersAndMutants/attribute_name.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/attribute_name.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/element_parts.a: output mutant caught on Node
TestJsxHelpersAndMutants/element_parts.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/element_parts.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/has_attribute_named.a: output mutant caught on Node
TestJsxHelpersAndMutants/has_attribute_named.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/has_attribute_named.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/intrinsic_element_named.a: output mutant caught on Node
TestJsxHelpersAndMutants/intrinsic_element_named.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/intrinsic_element_named.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/match_exactly.a: output mutant caught on Node
TestJsxHelpersAndMutants/match_exactly.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/match_exactly.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/match_ignoring_case.a: output mutant caught on Node
TestJsxHelpersAndMutants/match_ignoring_case.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/match_ignoring_case.a: output mutant caught on sanitized native
```

Parallelism 2, after:
```text
captured 1284 distinct upstream inputs from 27 JSX consumers
1285 sources, 23909 nodes (AttributeName/ElementParts), 31640 matcher pairs, 459187 intrinsic/attribute queries
514736 Go output rows byte-identical on Node, emitted JavaScript and sanitized native
TestJsxHelpersAndMutants/attribute_name.a: output mutant caught on Node
TestJsxHelpersAndMutants/attribute_name.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/attribute_name.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/element_parts.a: output mutant caught on Node
TestJsxHelpersAndMutants/element_parts.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/element_parts.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/has_attribute_named.a: output mutant caught on Node
TestJsxHelpersAndMutants/has_attribute_named.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/has_attribute_named.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/intrinsic_element_named.a: output mutant caught on Node
TestJsxHelpersAndMutants/intrinsic_element_named.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/intrinsic_element_named.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/match_exactly.a: output mutant caught on Node
TestJsxHelpersAndMutants/match_exactly.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/match_exactly.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/match_ignoring_case.a: output mutant caught on Node
TestJsxHelpersAndMutants/match_ignoring_case.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/match_ignoring_case.a: output mutant caught on sanitized native
```

Parallelism 8, before:
```text
captured 1284 distinct upstream inputs from 27 JSX consumers
1285 sources, 23909 nodes (AttributeName/ElementParts), 31640 matcher pairs, 459187 intrinsic/attribute queries
514736 Go output rows byte-identical on Node, emitted JavaScript and sanitized native
TestJsxHelpersAndMutants/attribute_name.a: output mutant caught on Node
TestJsxHelpersAndMutants/attribute_name.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/attribute_name.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/element_parts.a: output mutant caught on Node
TestJsxHelpersAndMutants/element_parts.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/element_parts.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/has_attribute_named.a: output mutant caught on Node
TestJsxHelpersAndMutants/has_attribute_named.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/has_attribute_named.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/intrinsic_element_named.a: output mutant caught on Node
TestJsxHelpersAndMutants/intrinsic_element_named.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/intrinsic_element_named.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/match_exactly.a: output mutant caught on Node
TestJsxHelpersAndMutants/match_exactly.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/match_exactly.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/match_ignoring_case.a: output mutant caught on Node
TestJsxHelpersAndMutants/match_ignoring_case.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/match_ignoring_case.a: output mutant caught on sanitized native
```

Parallelism 8, after:
```text
captured 1284 distinct upstream inputs from 27 JSX consumers
1285 sources, 23909 nodes (AttributeName/ElementParts), 31640 matcher pairs, 459187 intrinsic/attribute queries
514736 Go output rows byte-identical on Node, emitted JavaScript and sanitized native
TestJsxHelpersAndMutants/attribute_name.a: output mutant caught on Node
TestJsxHelpersAndMutants/attribute_name.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/attribute_name.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/element_parts.a: output mutant caught on Node
TestJsxHelpersAndMutants/element_parts.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/element_parts.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/has_attribute_named.a: output mutant caught on Node
TestJsxHelpersAndMutants/has_attribute_named.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/has_attribute_named.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/intrinsic_element_named.a: output mutant caught on Node
TestJsxHelpersAndMutants/intrinsic_element_named.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/intrinsic_element_named.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/match_exactly.a: output mutant caught on Node
TestJsxHelpersAndMutants/match_exactly.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/match_exactly.a: output mutant caught on sanitized native
TestJsxHelpersAndMutants/match_ignoring_case.a: output mutant caught on Node
TestJsxHelpersAndMutants/match_ignoring_case.a: output mutant caught on emitted JavaScript
TestJsxHelpersAndMutants/match_ignoring_case.a: output mutant caught on sanitized native
```
