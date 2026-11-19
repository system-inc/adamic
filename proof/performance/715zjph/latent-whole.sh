# #715zjph fix B: stage3/census/latent whole at the head, Compiler's standing rule for a new lowering field.
# The ordinary packages, then the scratch census overlay built by make_overlay.py and its in-process tests, as the
# latent README runs them. Run after environment.sh, sandboxed-job.sh, sandboxed_typescript_source,
# sandboxed_npm stage3/api and sandboxed_products ./internal/lower. Never merged.
set -u
W=$(pwd)
O=$TMPDIR/latent-whole
mkdir -p "$O"
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$O/$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*\.go:[0-9]+:)' "$O/$name.log" | head -12 | cut -c1-300; bad=1; }; grep -cE '^(ok|---) ' "$O/$name.log" | sed 's/^/  lines: /'; return $code; }
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg), at $(git rev-parse HEAD 2>/dev/null)"
step ordinary go test -count=1 -timeout 30m ./stage3/census/latent/...
grep -E '^(ok|FAIL|\?) ' "$O/ordinary.log" | cut -c1-120
step overlay python3 stage3/census/latent/make_overlay.py "$W" "$O/latent"
gofmt -w "$O/latent/"*.go
step tool-build go build -buildvcs=false -overlay="$O/latent/overlay.json" -o "$O/census" ./stage3/census/latent/tool
step census-tag go test -count=1 -timeout 30m -tags=census_overlay -overlay="$O/latent/overlay.json" ./stage3/census/latent
step tool go test -count=1 -timeout 30m -overlay="$O/latent/overlay.json" ./stage3/census/latent/tool
grep -E '^(ok|FAIL|---) ' "$O/census-tag.log" "$O/tool.log" | grep -v PASS | cut -c1-160 | head -20
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
