"""Run from the detached scratch worktree after setup.
source /workspace/scratch/next-stops-fb8db081/env.sh
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
then push HEAD:refs/heads/typescript/next-stops-fb8db081 without force.
Baseline is archived 89614cc0 CommonJSBundle rows in base/CommonJSBundle.
No compiler overlay is used. lower.go.txt is only a source evidence snapshot.
"""
