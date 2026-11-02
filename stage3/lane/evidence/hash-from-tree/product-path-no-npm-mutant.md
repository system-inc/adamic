# Product path without npm

Base origin/main: `3ea374e1d564b5989907510fca976af3493aa89a`.

Reports use the fixed identity `stage3-adapted-tree`, the product name in products_test.go. The supplied path ends in `/tree`, not a product key. The product build and planner read set own invalidation. Shared artifact preparation and verification keep artifact.input_key unchanged.

The test uses a PATH containing working node and python3 links, confirms npm is missing, and checks both check_unit_inputs and the identity reported by run_unit. Its runtime is computed from the current harness so a mutant reaches the input step.

Caller audit across stage3:

- artifact.py prepare: compute and recheck the shared build key, outside unit execution.
- shard.py shared_artifact: shared path only; check_unit_inputs: shared path only after this fix; run_unit delegates identity.
- merge_shards.py: shared-artifact merge against same-artifact run.sh whole-run results. Go wrappers only invoke its mocked shared fixtures through test_shards.py, never product-path merge. Left unchanged.
- prove_budget.py: manual shared-artifact cold-ledger and whole-run proof, including merge; not a gate unit.
- test_artifact.py: isolated shared-build invalidation/reuse fixtures.
- test_shards.py: shared adaptation-edit and mocked merge fixtures; plan-invariance fixture wraps input_key to detect unwanted calls.
- prove_hash_mutants.py: retained manual mutant inserts input_key into the plan builder.

No other callers found. Scratch copy: `/workspace/scratch/A2-no-npm-mutant`. Command from its lane directory:

```sh
python3 -B -m unittest test_shards.ShardTests.test_product_input_identity_without_npm -v
```

Positive passed. Mutant: replace the product/shared branch in check_unit_inputs with unconditional `key, _ = input_key()`. Result, exit 1:

```text
test_product_input_identity_without_npm (test_shards.ShardTests.test_product_input_identity_without_npm) ... ERROR

======================================================================
ERROR: test_product_input_identity_without_npm (test_shards.ShardTests.test_product_input_identity_without_npm)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/workspace/scratch/A2-no-npm-mutant/stage3/lane/test_shards.py", line 161, in test_product_input_identity_without_npm
    key, inputs_hash = shard.check_unit_inputs(unit, plan)
                       ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
  File "/workspace/scratch/A2-no-npm-mutant/stage3/lane/shard.py", line 138, in check_unit_inputs
    key, _ = input_key()
             ^^^^^^^^^^^
  File "/workspace/scratch/A2-no-npm-mutant/stage3/lane/artifact.py", line 34, in input_key
    tools = {name: subprocess.check_output([name, '--version'], text=True).strip()
                   ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
  File "/opt/codex/runtimes/codex-primary-runtime/dependencies/python/lib/python3.12/subprocess.py", line 466, in check_output
    return run(*popenargs, stdout=PIPE, timeout=timeout, check=True,
           ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
  File "/opt/codex/runtimes/codex-primary-runtime/dependencies/python/lib/python3.12/subprocess.py", line 548, in run
    with Popen(*popenargs, **kwargs) as process:
         ^^^^^^^^^^^^^^^^^^^^^^^^^^^
  File "/opt/codex/runtimes/codex-primary-runtime/dependencies/python/lib/python3.12/subprocess.py", line 1026, in __init__
    self._execute_child(args, executable, preexec_fn, close_fds,
  File "/opt/codex/runtimes/codex-primary-runtime/dependencies/python/lib/python3.12/subprocess.py", line 1955, in _execute_child
    raise child_exception_type(errno_num, err_msg, err_filename)
FileNotFoundError: [Errno 2] No such file or directory: 'npm'

----------------------------------------------------------------------
Ran 1 test in 0.058s

FAILED (errors=1)
```

Restored scratch shard.py byte for byte and reran: passed. Production artifact.py and merge_shards.py were unchanged by the fix.

## Real Go product-path reproduction

Apply, generate check, product preparation and the complete ordinary package check passed. With no other preparation or tests running, the following units were run serially through go test.

PATH: `/workspace/adamic-tools/bin:/opt/codex/runtimes/codex-primary-runtime/dependencies/python/bin:/usr/local/bin:/usr/bin:/bin`. npm was absent there and in both toolchain-prefixed directories. Node v24.19.0 and Python 3.12.14 were available.

Command for each unit: `go test -json -count=1 -run ^TEST_NAME$ -timeout=5m ./stage3/lane/`, using the absolute Go executable, with GOFLAGS `-parallel=1 -buildvcs=false -trimpath`.

- `TestShard_Stage3LaneFile002`: passed in 3.57 seconds; complete command 3.948 seconds.
- `TestShard_Stage3LaneFile003`: passed in 6.50 seconds; complete command 6.909 seconds.
- `TestShard_Stage3LaneFile324`: passed in 4.78 seconds; complete command 5.119 seconds.

Logs: `/workspace/scratch/A2-no-npm-go.log` and the individual JSON logs it names.
