source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && source cloud/gates/steps.sh && gate_begin oracle-whole || exit 1
# ask: --size F64
# oracle-whole: go test ./internal/oracle/ whole, products settled, as one job (its three recorded parts are not this recipe) (cloud/gates/diff-to-gates.json).
# This file is the whole command of a Loom sealed job at the head's own sha, from its first line, which opens sealed so a
# rented machine takes it. cloud/submit-gates.py matches it by sha256 against the job's commandSha256, so any change here
# is a new recipe, and jobs of the old one stop counting.
# A few of the oracle's tests read an npm install, so the image's projects are installed offline first.
gate_preamble npm sandboxed_npm
gate_products ./internal/oracle
gate_go_test oracle-whole ./internal/oracle/
gate_end
