# Array second pass reproduction

Run on Linux. Source /workspace/adamic-tools/env.sh. Install stock TypeScript
6.0.3 and put its tsc directory first on PATH. Use the pinned test262 submodule.
Build the runner from a separate worktree at af12899, not this branch:

```sh
git worktree add --detach /workspace/array2-validity af12899
# Use the same cohere submodule checkout in that worktree.
cd /workspace/array2-validity
go build -buildvcs=false -o /workspace/array2-runner ./cmd/adamic-test262
cd /workspace/adamic
PATH=/workspace/array2-tsc/node_modules/.bin:$PATH /workspace/array2-runner -adapt -json -root /workspace/adamic -test262 /workspace/test262 -work /tmp/array2-final-measure built-ins/Array > /tmp/array2-final-measure.json 2> /tmp/array2-final-measure.log
```

The runner builds the compiler from its -root, so the independent runner always
measures this checkout. For the baseline compiler use the merge commit 8fc9808
in another checkout; the exact baseline result is archived. All fixture and mutant
commands, Linux results, and known cohere CLI limitation are in
second-pass-report.md. Never pipe test output; capture it to a log file.
