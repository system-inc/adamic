"""Recipe; each command is a separate bounded invocation, not one long shell.
source /workspace/scratch/next-stops-763d5fc8-env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/next-stops-763d5fc8-typescript
export CENSUS_MODULE_FORMAT=CommonJSBundle
python3 stage3/census/next-stops-763d5fc8/run.py apply bash stage3/apply.sh /workspace/scratch/next-stops-763d5fc8-adapted
python3 stage3/census/next-stops-763d5fc8/run.py build-ready go build -o /workspace/scratch/next-stops-763d5fc8-census ./stage3/census/tool
For (first,last) in (1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79):
  python3 stage3/census/next-stops-763d5fc8/run.py group-FIRST-LAST python3 stage3/census/next-stops-763d5fc8/group.py FIRST LAST
  python3 stage3/census/next-stops-763d5fc8/checkpoint.py
Analyze, resolve owner/phase from measured source, write report, validate, and push bundle complete.
Only then export CENSUS_MODULE_FORMAT=EcmaScript and repeat root groups (no apply or build).
Final evidence is based on 763d5fc83558ae8ceb0202189b18430dff499ab8; change only this census directory.
Baseline: typescript/next-stops-55c375d9, 78a6c9dad71aeec7da329de10aea5387f75392b7 (both formats).
Checkpoint branch: typescript/next-stops-763d5fc8-checkpoint. Final branch: typescript/next-stops-763d5fc8.
Skip the separate sys.ts single-file census.
"""
