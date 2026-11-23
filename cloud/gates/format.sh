source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && source cloud/gates/steps.sh && gate_begin format || exit 1
# format: gofmt -l on the head's changed .go files prints nothing, and go vet of every touched package
# (cloud/gates/diff-to-gates.json). The changed files are the head's own, HEAD^ to HEAD; a file under cohere/ is the
# submodule's and isn't read here.
# This file is the whole command of a Loom sealed job at the head's own sha, from its first line, which opens sealed so a
# rented machine takes it. cloud/submit-gates.py matches it by sha256 against the job's commandSha256, so any change here
# is a new recipe, and jobs of the old one stop counting.
gate_base
git diff --no-renames --diff-filter=d --name-only -z "$gateBase" HEAD -- '*.go' ':(exclude)cohere' > "$gateScratch/changed" ||
	gate_abort changed "git diff of the head's .go files"
mapfile -d '' -t gateFiles < "$gateScratch/changed"
printf 'changed .go files: %s\n' "${#gateFiles[@]}"
if [ ${#gateFiles[@]} -eq 0 ]; then
	gate_pass "gofmt: no .go file changed"
	gate_pass "go vet: no package touched"
	gate_end
fi
gofmt -l "${gateFiles[@]}" > "$gateScratch/gofmt" 2>&1 || gate_fail gofmt "gofmt -l exited nonzero: $(head -5 "$gateScratch/gofmt")"
if [ -s "$gateScratch/gofmt" ]; then
	gate_fail gofmt "gofmt -l lists $(wc -l < "$gateScratch/gofmt") files: $(head -10 "$gateScratch/gofmt" | tr '\n' ' ')"
else
	gate_pass "gofmt -l over ${#gateFiles[@]} changed files"
fi
mapfile -t gatePackages < <(for file in "${gateFiles[@]}"; do printf './%s\n' "$(dirname -- "$file")"; done | sort -u)
if go vet "${gatePackages[@]}" > "$gateScratch/vet" 2>&1; then
	gate_pass "go vet of ${#gatePackages[@]} touched packages: ${gatePackages[*]}"
else
	tail -30 "$gateScratch/vet"
	gate_fail vet "go vet ${gatePackages[*]}"
fi
gate_end
