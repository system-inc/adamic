# 99w proofs

The base and ordinary compiler are unmodified main a28effb2c. No compiler overlay,
refusal suppression, backend change or partial-root fallback is used.
provenance.json retains the full commit and exact compiler binary hashes.

| Check | Result |
| --- | --- |
| Full apply before and after, --build-product | Both exit 0 |
| Adapter replay on full applied tree | Zero files, zero edits, javascriptChanging false |
| Owner contract | Pass, ten no-write refusals, foreign edit and LF/CRLF preserved |
| Replay contract | Source bytes and mtime preserved; zero writes |
| Adamic types, all 79 sorted roots before and after | Both exit 0; zero diagnostics |
| Ordinary compiler builds before and after | Both exit 0 |
| Both public declarations | Byte-identical |
| Complete core.ts JavaScript with comments kept and removed | Byte-identical |
| All nine normally emitted JavaScript files | Byte-identical |
| Source scope | All 746 source files compared; only the pinned core.ts edit differs |
| generate_shards --check | Exit 0 |
| stage3/lane Go fixtures | Exit 0, 262.678 seconds |
| Original-cast mutant | core.ts:810:89, Refused: a cast the runtime can't check |

comparison.json contains hashes and sizes of declarations and JavaScript, and
asserts the exact owner-only source delta. These complete-file hashes are
measurement evidence, never adaptation gates. Cast-node count stays 26; any-node
count decreases from 15 to 14. Only the bridge spelling changes.

Both full applies ran before their compiler builds. The baseline apply enumerated
the original main adaptation set; the accepted apply includes 99w immediately
after 99u in the real Python sort. The census reads that same completed accepted
source tree directly. Source writes by replay are zero. Later evidence files
change product keys without changing the frozen transformation.

Commands used after sourcing /workspace/adamic-tools/env.sh:

```sh
bash cloud/setup.sh > /workspace/scratch/99w-setup.log 2>&1
npm ci --prefix stage3/api
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/99w-before
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/99w-after
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-w/contract.cjs /workspace/typescript-source
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-w/adapt.cjs /workspace/99w-after
go build -o /workspace/scratch/99w-bin/ ./cmd/adamic ./stage3/census/tool
# Both types commands receive all 79 sorted src/compiler/**/*.ts paths together.
/workspace/scratch/99w-bin/adamic types ROOTS
# In each applied tree, using the same upstream lockfile dependency install:
npm run build:compiler
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-w/compare.cjs /workspace/99w-before /workspace/99w-after /workspace/99w-after
python3 stage3/lane/generate_shards.py --check
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source go test -count=1 -parallel=1 -skip 'TestShard_|TestProduct_' ./stage3/lane/
/workspace/scratch/99w-bin/tool /workspace/99w-after/src/compiler census-after.jsonl
# In an isolated copy of accepted src:
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-w/mutant.cjs /workspace/99w-mutant
/workspace/scratch/99w-bin/tool /workspace/99w-mutant/src/compiler/binder.ts mutant-binder.jsonl
```

Every command's output was written to a log file. Retained compressed logs and
before-checks.json / after-checks.json record the results. No Go file is added.
The ordinary per-root census result is documented separately in CENSUS.md.
The oracle is skipped under the requested type-only exception.
