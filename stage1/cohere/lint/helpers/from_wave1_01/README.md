# CFG enter helper

Claim d6fbc1351 was pushed before code on codex/lint-helpers-from-lint-wave1-01, based on origin/codex/lint-helpers. The parked rule branch is 850fcdd0f.

enter.a exports enter(node, state), Block and State. The caller hands a valid non-null block; incoming records its incoming-edge flag. The helper assigns reachable from incoming and assigns state.current to that exact block object. Block has no owned graph edges, so the current-block reference is acyclic. Caller graph adapters supply all other block/event/edge fields. Existing caller aliases observe the mutation and the current-block identity is preserved; a previously reachable block can become unreachable.

Actual Go's private Builder.enter is observed through a temporary source overlay while all matching upstream tests for each consumer run. Instrumentation records pre-call state and actual post-call state; no expected output is passed to Adamic. Temporary files and capture logs live in test directories. Independent controls cover all incoming/reachable booleans, five handed block indices and six prior-current choices (nil, same identity, different identity), 120 observations. This boundary does not certify invalid Go nil-pointer panic prose, arbitrary CFG building, rule findings or the unported adjacent CFG helpers.

Run with source /workspace/adamic-tools/env.sh, then redirect all output:

    go test ./stage1/cohere/lint/helpers/from_wave1_01 -count=1 -timeout 15m -v > /tmp/enter-tests.log 2>&1

One helper per .a file. No finding.ts, context.ts, main.ts, registry generator, shared oracle or shared comparison changes.
