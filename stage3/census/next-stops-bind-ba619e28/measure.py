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
CommonJSBundle only; do not run EcmaScript.
Final evidence is based on ba619e28774130e24331c3026a938cd6f3cd6a34; change only this census directory.
Compare typescript/next-stops-55c375d9 when available, otherwise typescript/next-stops-8e3a784a at d8a5019190cf3c176f1b11385d6437c06f04ba4e.
Checkpoint branch: typescript/next-stops-bind-ba619e28. Final branch: typescript/next-stops-bind-ba619e28.
Skip the separate sys.ts single-file census.
"""
