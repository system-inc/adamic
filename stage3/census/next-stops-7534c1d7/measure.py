"""Recipe; each command is a separate bounded invocation, not one long shell.
source /workspace/scratch/next-stops-7534c1d7-env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/next-stops-7534c1d7/typescript
export CENSUS_MODULE_FORMAT=CommonJSBundle
python3 stage3/census/next-stops-7534c1d7/run.py apply bash stage3/apply.sh /workspace/scratch/next-stops-7534c1d7-adapted
python3 stage3/census/next-stops-7534c1d7/run.py build-ready go build -o /workspace/scratch/next-stops-7534c1d7-census ./stage3/census/tool
For (first,last) in (1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79):
  python3 stage3/census/next-stops-7534c1d7/run.py group-FIRST-LAST python3 stage3/census/next-stops-7534c1d7/group.py FIRST LAST
  python3 stage3/census/next-stops-7534c1d7/checkpoint.py
Analyze, resolve owner/phase from measured source, write report, validate, and push bundle complete.
Only then export CENSUS_MODULE_FORMAT=EcmaScript and repeat root groups (no apply or build).
Final evidence is based on 7534c1d7575db02bb7b6512119e0685987b7e333; change only this census directory.
Baseline: typescript/next-stops-ff4ce03f, 2147e992cf42301c1e20c64c8d39c9ab027c65d7 (both formats).
Checkpoint branch: typescript/next-stops-7534c1d7-checkpoint. Final branch: typescript/next-stops-7534c1d7.
Skip the separate sys.ts single-file census.
"""
