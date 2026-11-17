#!/usr/bin/env bash
# #r11wppy item 1's public candidates type-checked: the internal change plus the six public lines alone, and plus
# the four InterfaceTypeWithDeclaredMembers members alone. Run after environment.sh, sealed-job.sh,
# sealed_typescript_source and sealed_npm stage3/api. Never merged.
set -u
A="$TMPDIR/adapted"
bash stage3/apply.sh "$A" > "$TMPDIR/apply.log" 2>&1 || { echo "VERDICT FAIL apply"; tail -5 "$TMPDIR/apply.log"; exit 3; }
for variant in six four; do
	echo "=== $variant"
	node proof/performance/r11wppy-fixpoint/verify.mjs "$A" "proof/performance/r11wppy-fixpoint/edits-public-$variant.diff" "$TMPDIR/verify-$variant" | grep -vE '^  (kept |left out \(services)'
done
