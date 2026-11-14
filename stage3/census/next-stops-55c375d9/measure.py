"""Recipe; each command is a separate bounded invocation, not one long shell.
source /workspace/scratch/next-stops-55c375d9-env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-050880ce
export CENSUS_MODULE_FORMAT=CommonJSBundle
python3 stage3/census/next-stops-55c375d9/run.py apply bash stage3/apply.sh /workspace/scratch/next-stops-55c375d9-adapted
python3 stage3/census/next-stops-55c375d9/run.py build-ready go build -o /workspace/scratch/next-stops-55c375d9-census ./stage3/census/tool
For (first,last) in (1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79):
  python3 stage3/census/next-stops-55c375d9/run.py group-FIRST-LAST python3 stage3/census/next-stops-55c375d9/group.py FIRST LAST
  python3 stage3/census/next-stops-55c375d9/checkpoint.py
Analyze, resolve owner/phase from measured source, write report, validate, and push bundle complete.
Only then export CENSUS_MODULE_FORMAT=EcmaScript and repeat root groups (no apply or build).
Final evidence is based on 55c375d9d245b580c010c3e0b5659f2849f24fb6; change only this census directory.
Baseline: typescript/next-stops-8e3a784a, d8a5019190cf3c176f1b11385d6437c06f04ba4e (both formats).
Checkpoint branch: typescript/next-stops-55c375d9-checkpoint. Final branch: typescript/next-stops-55c375d9.
Skip the separate sys.ts single-file census.
"""
