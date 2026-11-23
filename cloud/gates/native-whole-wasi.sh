source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && source cloud/gates/steps.sh && gate_begin native-whole-wasi || exit 1
# ask: --size F64
# native-whole-wasi: go test ./internal/native/ whole with ADAMIC_TEST_WASI=1 (cloud/gates/diff-to-gates.json).
# This file is the whole command of a Loom sealed job at the head's own sha, from its first line, which opens sealed so a
# rented machine takes it. cloud/submit-gates.py matches it by sha256 against the job's commandSha256, so any change here
# is a new recipe, and jobs of the old one stop counting.
gate_products ./internal/native
ADAMIC_TEST_WASI=1 gate_go_test native-whole-wasi ./internal/native/
gate_end
