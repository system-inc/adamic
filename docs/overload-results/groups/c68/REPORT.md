Measured all seven assigned intervals and recorded their current stops; no compiler acceptance changed.
Compiler c68bf26cb4225e1f24a84f1a412f10e7eaa3da9a; source pin 388096e6a83a4e9d287fb827f793c599ba1bf0ad.
Scoped census and count calculation succeeded; four exact replays matched, two selectors mismatched.
Wrong reason, wrong byte count, shifted endpoint and false reveal mutants are rejected.
No new revealed bytes; callback and structural contract implementations remain unfinished toward step 16.

All 82 adapted source hashes match the pinned stock. The seven intervals remain
fully hidden: 13,625; 6,899; 6,078; 5,748; 11,417; 7,289; 7,102 bytes.
These total 58,158 bytes. Each independent declaration overlapping an interval
was attempted, with all project declarations registered. The census retains its
checker-rejected measurement label and no-output guard. It is not a compilation
of the corpus and excludes unrelated intervals.

The replay worker was built before applying census-scope.patch.gz. Thus its
normal selector and exact diagnostic matcher are unchanged. Four selectors
reproduce the visitNode, visitNodes, Block and EvaluatorResult refusals.
Hidden-01 instead reports an indirect overload value at declarations.ts:612:90;
the optional-callback selector reports visitNode's earlier parameter refusal.
Their exit 1 records signature mismatches, not successful compilation.

The exact commands and exits are in replays.json, each complete record and log
is beside it. The scoped census is census.jsonl.gz. Regenerate an overlay with
stage3/census/latent/make_overlay.py on compiler c68bf26c, build the replay worker,
then decompress and apply census-scope.patch.gz to internal_lower_latent_units.go
in the overlay before building the census tool. The measurement command was:

```sh
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/overload-groups-c68-census \
  /tmp/overload-results-adapted/src/compiler /tmp/overload-groups-c68.jsonl \
  > /tmp/overload-groups-c68-census.log 2>&1
python3 docs/overload-results/measure-regions.py \
  /workspace/overload-results-sourcepin /tmp/overload-results-adapted/src/compiler \
  /workspace/overload-results-sourcepin/stage3/census/hidden/evidence/full.jsonl.gz \
  /tmp/overload-groups-c68.jsonl docs/overload-results/groups/c68/regions.json \
  --all --compiler c68bf26cb4225e1f24a84f1a412f10e7eaa3da9a \
  > /tmp/overload-groups-regions.log 2>&1
python3 docs/overload-results/groups/c68/validate.py \
  > docs/overload-results/groups/c68/validation.log.txt 2>&1
```

The measurement helper now derives each census's original root separately;
the pinned source and the new run have different absolute paths. Historical
three-region use remains available; --all selects the seven assigned intervals.
No worker branch was merged. Hidden-06's separate 8,048-byte constraint-storage
reveal is excluded because its implementation is absent from c68bf26c.

Next work by assigned bytes is hidden-01's callback contract (20,524), hidden-05
(11,826), hidden-06 (11,417), hidden-13 (7,289), then hidden-14 (7,102).
An input-domain proof and a closure boundary are still needed for callback
specialization. Checking only an AST kind does not establish Block's complete
contract, and narrowing a shared EvaluatorResult could let a wider alias violate
its promised field type. This group introduces no relaxation for those cases.

Setup timings already recorded for this compiler: Go 0.024s, Node 0.025s,
submodules 0.069s, markdown 0.082s, clang 0.166s, build 47.079s, cache 47.260s,
done 47.284s. nproc 5, Go 1.27.1, Node 24.19.0, clang 20.1.8.
An initial overlay regeneration omitted the toolchain env and failed with
Go: Unknown option: run. Sourcing /workspace/adamic-tools/env.sh fixed it.
No production compiler file or fixture changed in this measurement group;
no package-wide tests, full gate, or counts refresh was needed.
