# Wave 01 continuation

Three isolated native rule ports: child-process error listeners, response status checks,
and independent awaits in loops. Each rule is in its own `.a` file. `suite.a` emits the
existing canonical diagnostic protocol, including the fix and suggestion counts.

The production Go oracle in `testdata/oracle.go` runs the unchanged pinned rules with
its own program loader. The bridge supplies declaration ancestry, raw compiler
reference flags, optional loop-header slots and a generic evaluation-order graph.
All producer, use, state, taint and path classifications run in native Adamic.

Run the complete gate without changing the shared harness:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_01_next/testdata/verify.py \
  --repository /workspace/adamic --typescript /workspace/wave-01-typescript \
  --artifacts /workspace/wave-01-next-final > /tmp/wave-01-next-final.log 2>&1
```

The TypeScript corpus must be at 050880ce59e30b356b686bd3144efe24f875ebc8.
The generated `.ts` and `.d.ts` files are oracle input fixtures, not Adamic rule code.
Committed results are under `validation/`; native binaries remain outside the repository.
See `REPORT.md` for measurements, mutants and integration limits.
