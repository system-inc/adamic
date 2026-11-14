"""Recipe; each command is a separate bounded invocation, not one long shell.
source /workspace/scratch/c4v7-census/head-env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/c4v7-census/head-typescript
export CENSUS_MODULE_FORMAT=CommonJSBundle
python3 stage3/census/next-stops-c4v7/run.py apply bash stage3/apply.sh /workspace/scratch/c4v7-census/adapted-head
python3 stage3/census/next-stops-c4v7/run.py build-ready go build -o /workspace/scratch/c4v7-census/head-census ./stage3/census/tool
For (first,last) in (1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79):
  python3 stage3/census/next-stops-c4v7/run.py group-FIRST-LAST python3 stage3/census/next-stops-c4v7/group.py FIRST LAST
  python3 stage3/census/next-stops-c4v7/checkpoint.py
Analyze, resolve owner/phase from measured source, write report, validate, and push bundle complete.
Only then export CENSUS_MODULE_FORMAT=EcmaScript and repeat root groups (no apply or build).
Final evidence is a fast-forward branch based on 5c8338c4, with all changed paths under this directory.
"""
