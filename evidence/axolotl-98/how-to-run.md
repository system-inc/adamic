# How to rerun

Built on main a23ba096388ceb3774b94d2db70009316d90b0cf, on a Linux cloud box (4 processors), Go 1.27.1, Node v24.22.2.
This follows stage3/meter/twice-daily.sh: the same upstream clone, the same `npm ci`, the same
`stage3/apply.sh`, and the same `stage3/census/tool` binary run on the adapted `src/compiler`
directory (stage 0's checker with its options and prelude unchanged). The own-file attribution in
`own_file.py` is report.py's: every diagnostic in every census record, assigned to its primary
`file:line:col`, deduplicated.

```sh
git fetch origin main && git checkout --detach a23ba096388ceb3774b94d2db70009316d90b0cf
git submodule update --init --recursive --depth 1
bash cloud/setup.sh --wasi-sdk > /tmp/setup.log 2>&1 && source /opt/adamic-tools/env.sh
mkdir -p /tmp/axo
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/axo/upstream > /tmp/axo/upstream.log 2>&1
git -C /tmp/axo/upstream rev-parse HEAD   # 050880ce59e30b356b686bd3144efe24f875ebc8
(cd stage3/api && npm ci --ignore-scripts --no-audit --no-fund) > /tmp/axo/packages.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/axo/upstream bash stage3/apply.sh /tmp/axo/tree > /tmp/axo/apply.log 2>&1
go build -o /tmp/axo/census ./stage3/census/tool > /tmp/axo/census-build.log 2>&1
/tmp/axo/census /tmp/axo/tree/src/compiler /tmp/axo/census.jsonl > /tmp/axo/census.log 2>&1   # 3m46s here
python3 -I evidence/axolotl-98/own_file.py /tmp/axo/tree /tmp/axo/census.jsonl /tmp/axo/diagnostics.txt \
  transformers/es2017.ts transformers/classFields.ts transformers/es2015.ts sys.ts \
  moduleNameResolver.ts resolutionCache.ts sourcemap.ts core.ts
```

The last command prints the count for each file and writes the diagnostics, with paths made
relative to the tree, to diagnostics.txt. Indented lines are the checker's elaborations of the
line above them. They are not counted.

## Result

| File | Here (main a23ba096) | Meter 20261101T131708Z (main 141ad1d2) |
| --- | ---: | ---: |
| src/compiler/transformers/es2017.ts | 6 | 6 |
| src/compiler/transformers/classFields.ts | 6 | 6 |
| src/compiler/transformers/es2015.ts | 5 | 5 |
| src/compiler/sys.ts | 4 | 4 |
| src/compiler/moduleNameResolver.ts | 3 | 3 |
| src/compiler/resolutionCache.ts | 1 | 1 |
| src/compiler/sourcemap.ts | 1 | 1 |
| src/compiler/core.ts | 1 | 1 |
| **Total** | 27 | 27 |

Between 141ad1d2 and a23ba096, nothing under internal/load, internal/lower or stage3/adapt changed.
The diff is the meter run itself, docs, conformance fixtures and stage3/census/latent.
