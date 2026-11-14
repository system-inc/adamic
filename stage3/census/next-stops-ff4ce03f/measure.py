"""Recipe; each command is a separate bounded invocation, not one long shell.
source /workspace/scratch/next-stops-ff4ce03f-env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/next-stops-ff4ce03f-typescript
export CENSUS_MODULE_FORMAT=CommonJSBundle
python3 stage3/census/next-stops-ff4ce03f/run.py apply bash stage3/apply.sh /workspace/scratch/next-stops-ff4ce03f-adapted
python3 stage3/census/next-stops-ff4ce03f/run.py build-ready go build -o /workspace/scratch/next-stops-ff4ce03f-census ./stage3/census/tool
For (first,last) in (1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79):
  python3 stage3/census/next-stops-ff4ce03f/run.py group-FIRST-LAST python3 stage3/census/next-stops-ff4ce03f/group.py FIRST LAST
  python3 stage3/census/next-stops-ff4ce03f/checkpoint.py
Analyze, resolve owner/phase from measured source, write report, validate, and push bundle complete.
Only then export CENSUS_MODULE_FORMAT=EcmaScript and repeat root groups (no apply or build).
Final evidence is based on ff4ce03f5f23bc385c8855da202ff81498ab5de3; change only this census directory.
Baseline: typescript/next-stops-c4v7-6, 26bbed1b9d6d58772b6131b81d8a112d11e18450 (both formats).
Checkpoint branch: typescript/next-stops-ff4ce03f-checkpoint. Final branch: typescript/next-stops-ff4ce03f.
Skip the separate sys.ts single-file census.
"""
