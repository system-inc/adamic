"""Recipe; each command is a separate bounded invocation, not one long shell.
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-f2546f1f-source
export CENSUS_MODULE_FORMAT=CommonJSBundle
python3 stage3/census/next-stops-f2546f1f/run.py apply bash stage3/apply.sh /workspace/scratch/next-stops-f2546f1f-adapted
python3 stage3/census/next-stops-f2546f1f/run.py build-ready go build -o /workspace/scratch/next-stops-f2546f1f-census ./stage3/census/tool
For (first,last) in (1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79):
  python3 stage3/census/next-stops-f2546f1f/run.py group-FIRST-LAST python3 stage3/census/next-stops-f2546f1f/group.py FIRST LAST
  python3 stage3/census/next-stops-f2546f1f/checkpoint.py
Analyze, resolve owner/phase from measured source, write report, validate, and push bundle complete.
Only then export CENSUS_MODULE_FORMAT=EcmaScript and repeat root groups (no apply or build).
Final evidence is a fast-forward branch based on f2546f1f, with all changed paths under this directory.
"""
