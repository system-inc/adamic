"""Reproduction recipe: CommonJSBundle only, fresh main baseline and head.
Create detached scratch worktrees at the SHAs in provenance.json, recursively initialize
submodules with --depth 1, clone TypeScript v6.0.3 at the pinned upstream SHA, and npm ci
in each worktree/stage3/api. Copy env.sh.txt to scratch/env.sh, create its cache/temp
directories, then source /workspace/scratch/next-stops-37d96306/env.sh.
The unchanged stage3 tree is identical at both SHAs: build one adapted tree:
  CENSUS_REVISION=head python3 REPORT_DIR/run.py apply bash stage3/apply.sh --build-product /workspace/scratch/next-stops-37d96306/adapted
Build each ordinary tool from its own worktree:
  CENSUS_REVISION=REV python3 REPORT_DIR/run.py build go build -o /workspace/scratch/next-stops-37d96306/REV-census ./stage3/census/tool
REV is head (37d96306133ed2782184bee1bd8406d61b83a858), then base
(89614cc0f9d6aa7b7fe96b56c5c6c1c3eace9825). CENSUS_MODULE_FORMAT=CommonJSBundle.
For each REV, and each range (1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79):
  CENSUS_REVISION=REV python3 REPORT_DIR/run.py group-FIRST-LAST python3 REPORT_DIR/group.py FIRST LAST
  CENSUS_REVISION=REV python3 REPORT_DIR/checkpoint.py
Each invocation is separately bounded; do not wrap all groups in one run.py call.
Each root child emits two observations, compared except seconds. Three children at most.
Freeze the adapted source manifest before root measurement. Hash both binaries.
Then analyze.py, review emitting source into reviewed-stops.json, analyze.py again,
write-report.py, validate.py, finalize.py, checkpoint.py, and final non-forced push.
REPORT_DIR is report/stage3/census/next-stops-37d96306 under scratch.
Checkpoint: typescript/next-stops-37d96306-checkpoint.
Final: typescript/next-stops-37d96306. Only this census directory changes.
No separate sys.ts single-file census. No compiler or adaptation changes.
"""
