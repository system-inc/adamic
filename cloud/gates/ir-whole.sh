source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && source cloud/gates/steps.sh && gate_begin ir-whole || exit 1
# ir-whole: go test ./internal/ir/ whole, its call-target census among it (cloud/gates/diff-to-gates.json).
# This file is the whole command of a Loom sealed job at the head's own sha, from its first line, which opens sealed so a
# rented machine takes it. cloud/submit-gates.py matches it by sha256 against the job's commandSha256, so any change here
# is a new recipe, and jobs of the old one stop counting.
gate_products ./internal/ir
gate_go_test ir-whole ./internal/ir/
gate_end
