"""Recipe; each command is a separate bounded invocation, not one long shell.
source /workspace/scratch/next-stops-8e3a784a-env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/next-stops-8e3a784a-typescript
export CENSUS_MODULE_FORMAT=CommonJSBundle
python3 stage3/census/next-stops-8e3a784a/run.py apply bash stage3/apply.sh /workspace/scratch/next-stops-8e3a784a-adapted
python3 stage3/census/next-stops-8e3a784a/run.py build-ready go build -o /workspace/scratch/next-stops-8e3a784a-census ./stage3/census/tool
For (first,last) in (1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79):
  python3 stage3/census/next-stops-8e3a784a/run.py group-FIRST-LAST python3 stage3/census/next-stops-8e3a784a/group.py FIRST LAST
  python3 stage3/census/next-stops-8e3a784a/checkpoint.py
Analyze, resolve owner/phase from measured source, write report, validate, and push bundle complete.
Only then export CENSUS_MODULE_FORMAT=EcmaScript and repeat root groups (no apply or build).
Final evidence is based on 8e3a784aa5a601f57342c07b68f145fe56252c21; change only this census directory.
Baseline: typescript/next-stops-7534c1d7, 4eb05ced35dd2e90e3c688fdf55b0aab7d79e649 (both formats).
Checkpoint branch: typescript/next-stops-8e3a784a-checkpoint. Final branch: typescript/next-stops-8e3a784a.
Skip the separate sys.ts single-file census.
"""
