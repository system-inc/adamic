"""Recipe; each command is a separate bounded invocation, not one long shell.
source /tmp/next-stops-bind-ba619e28-env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/next-stops-bind-ba619e28-typescript
export CENSUS_MODULE_FORMAT=CommonJSBundle
python3 stage3/census/next-stops-bind-ba619e28/run.py apply bash stage3/apply.sh /tmp/next-stops-bind-ba619e28-adapted
python3 stage3/census/next-stops-bind-ba619e28/run.py build-ready go build -o /tmp/next-stops-bind-ba619e28-census ./stage3/census/tool
For (first,last) in (1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79):
  python3 stage3/census/next-stops-bind-ba619e28/run.py group-FIRST-LAST python3 stage3/census/next-stops-bind-ba619e28/group.py FIRST LAST
  python3 stage3/census/next-stops-bind-ba619e28/checkpoint.py
Analyze, resolve owner/phase from measured source, write report, validate, and push bundle complete.
Only then export CENSUS_MODULE_FORMAT=EcmaScript and repeat root groups (no apply or build).
Final evidence is based on ba619e28774130e24331c3026a938cd6f3cd6a34; change only this census directory.
Baseline: typescript/next-stops-7534c1d7, 4eb05ced35dd2e90e3c688fdf55b0aab7d79e649 (both formats).
Checkpoint branch: typescript/next-stops-bind-ba619e28-checkpoint. Final branch: typescript/next-stops-bind-ba619e28.
Skip the separate sys.ts single-file census.
"""
