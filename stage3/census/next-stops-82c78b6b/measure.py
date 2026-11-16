"""Run from the detached scratch worktree after setup.
source stage3/census/next-stops-82c78b6b/env.sh.txt
npm ci --prefix stage3/api (recorded via run.py npm-ci)
run.py apply bash stage3/apply.sh --build-product SCRATCH/adapted
run.py build go build -o SCRATCH/head-census ./stage3/census/tool
Freeze compiler source hashes before measurement.
CommonJSBundle first; for each range 1-10,11-20,...,71-79:
  python3 REPORT_DIR/run.py group-FIRST-LAST python3 REPORT_DIR/group.py FIRST LAST
  python3 REPORT_DIR/checkpoint.py
Validate all 79 pairs and push checkpoint before EcmaScript.
Set CENSUS_MODULE_FORMAT=EcmaScript and repeat groups without apply/build.
Each child emits two ordinary observations; compare everything except seconds.
Group and child bounds 1780 seconds; at most three independent root children.
Analyze, review source/owners/emitters, report, validate, finalize, checkpoint,
then push HEAD:refs/heads/typescript/next-stops-82c78b6b without force.
Baselines are archived fb8db081 rows in base/CommonJSBundle and base/EcmaScript.
No compiler overlay is used. lower.go.txt is only a source evidence snapshot.
"""
