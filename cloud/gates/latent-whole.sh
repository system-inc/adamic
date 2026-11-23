source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && source cloud/gates/steps.sh && gate_begin latent-whole || exit 1
# latent-whole: go test ./stage3/census/latent/... whole, the statecopy guard and the pass ledger among it
# (cloud/gates/diff-to-gates.json).
# This file is the whole command of a Loom sealed job at the head's own sha, from its first line, which opens sealed so a
# rented machine takes it. cloud/submit-gates.py matches it by sha256 against the job's commandSha256, so any change here
# is a new recipe, and jobs of the old one stop counting.
gate_preamble npm sandboxed_npm stage3/api
gate_products ./stage3/census/latent
gate_go_test latent-whole ./stage3/census/latent/...
gate_end
