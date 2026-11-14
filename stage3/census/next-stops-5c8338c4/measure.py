"""Recipe; each command is a separate bounded invocation, not one long shell.
source /workspace/scratch/next-stops-5c8338c4-env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/next-stops-5c8338c4-typescript
export CENSUS_MODULE_FORMAT=CommonJSBundle
python3 stage3/census/next-stops-5c8338c4/run.py apply bash stage3/apply.sh /workspace/scratch/next-stops-5c8338c4-adapted
python3 stage3/census/next-stops-5c8338c4/run.py build-ready go build -o /workspace/scratch/next-stops-5c8338c4-census ./stage3/census/tool
For (first,last) in (1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79):
  python3 stage3/census/next-stops-5c8338c4/run.py group-FIRST-LAST python3 stage3/census/next-stops-5c8338c4/group.py FIRST LAST
  python3 stage3/census/next-stops-5c8338c4/checkpoint.py
Analyze, resolve owner/phase from measured source, write report, validate, and push bundle complete.
Only then export CENSUS_MODULE_FORMAT=EcmaScript and repeat root groups (no apply or build).
Final evidence is a fast-forward branch based on 5c8338c4, with all changed paths under this directory.
"""
