"""Recipe; each command is a separate bounded invocation, not one long shell.
source /workspace/scratch/next-stops-37d96306-env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/next-stops-37d96306-typescript
export CENSUS_MODULE_FORMAT=CommonJSBundle
python3 stage3/census/next-stops-37d96306/run.py apply bash stage3/apply.sh /workspace/scratch/next-stops-37d96306-adapted
python3 stage3/census/next-stops-37d96306/run.py build-ready go build -o /workspace/scratch/next-stops-37d96306-census ./stage3/census/tool
For (first,last) in (1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79):
  python3 stage3/census/next-stops-37d96306/run.py group-FIRST-LAST python3 stage3/census/next-stops-37d96306/group.py FIRST LAST
  python3 stage3/census/next-stops-37d96306/checkpoint.py
Analyze, resolve owner/phase from measured source, write report, validate, and push bundle complete.
Only then export CENSUS_MODULE_FORMAT=EcmaScript and repeat root groups (no apply or build).
Final evidence is based on 37d96306133ed2782184bee1bd8406d61b83a858; change only this census directory.
Baseline: typescript/next-stops-89614cc0, 89614cc0f9d6aa7b7fe96b56c5c6c1c3eace9825 (both formats).
Checkpoint branch: typescript/next-stops-37d96306-checkpoint. Final branch: typescript/next-stops-37d96306.
Skip the separate sys.ts single-file census.
"""
