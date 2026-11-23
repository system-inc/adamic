source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && source cloud/gates/steps.sh && gate_begin lower-whole || exit 1
# ask: --size F16
# lower-whole: go test ./internal/lower/ whole, products settled (cloud/gates/diff-to-gates.json).
# This file is the whole command of a Loom sealed job at the head's own sha, from its first line, which opens sealed so a
# rented machine takes it. cloud/submit-gates.py matches it by sha256 against the job's commandSha256, so any change here
# is a new recipe, and jobs of the old one stop counting.
gate_products ./internal/lower
gate_go_test lower-whole ./internal/lower/
gate_end
