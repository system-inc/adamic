# Next stops on main f44b5c27: stopped before measurement

Target: `f44b5c272a39bcbd7cd2fb278939ac1c1832ce2e`.
Baseline branch: `typescript/next-stops-3007955b`, fetched at `933c8981edfa0e046893b1a1f1570c166a9e8e2c`; measured baseline compiler: `3007955b6db96ae5f65cedb2e680d77ad1b88664`.

Stopped under the instruction to stop if the tooling no longer fits, rather than change anything beyond the pin, baseline and paths. No apply, build or census children were run. No first stops on the target were observed, including sys.ts and the three root-order artifact roots. The expected toISOString and __filename stops remain unmeasured.

## Exact tooling mismatch

The fetched baseline directory does not contain `finalize.py`, `classify-artifacts.py`, or build scripts. Its workflow instead uses `assemble.py`, `write-report.py`, and `validate.py`, with build commands in `measure.py`.

More decisively, baseline `validate.py` unconditionally reads `build-before-setup-status.json` and asserts:

```python
failed=json.loads((out/'build-before-setup-status.json').read_text())
assert failed['exit_code']==1 and failed['seconds']<1800 and failed['process_group_cleaned']
```

It also reports `documented_initial_build_exit=1`. This is historical evidence of a failed build from the previous run, not a reusable measurement requirement. A fresh run does not have that artifact. Copying it would misrepresent this run; manufacturing that failure or changing validation logic is outside the requested pin/baseline/path-only edits. `write-report.py` likewise hardcodes that previous failed-build narrative.

The ordinary census interface at the target was inspected and still accepts a file root and output path, emitting individual and whole-program observations. No incompatibility of the compiler or apply pipeline was established.

## Retained evidence and scope

`previous-REPORT.md`, `previous-grouped-first-stops.json`, and `previous-comparison.json` retain the actual baseline evidence. Python tooling is copied for inspection with target directory and workspace paths adapted; it has not been executed or certified for this target. Report generation still requires baseline narrative updates and the historical-failure validation dependency needs resolution before running this workflow.

Only this new directory is included in this commit. No measurement checkpoint exists because no group ran. No grouped target table or per-root comparison can be reported without observations.
