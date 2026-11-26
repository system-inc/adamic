source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && source cloud/gates/steps.sh && gate_begin projects-whole || exit 1
# projects-whole: go test ./internal/load/projects/ whole, products settled (cloud/gates/diff-to-gates.json). Its
# TestTheRootProjectsAreLoadsOwn holds the root tsconfig.json and internal/load/projects/node/tsconfig.json to what
# go run ./internal/load/projects writes from the tree as tested (#kyq0pyn, #w5zzdjp).
# This file is the whole command of a Loom sealed job at the head's own sha, from its first line, which opens sealed so a
# rented machine takes it. cloud/submit-gates.py matches it by sha256 against the job's commandSha256, so any change here
# is a new recipe, and jobs of the old one stop counting.
gate_products ./internal/load/projects
gate_go_test projects-whole ./internal/load/projects/
gate_end
