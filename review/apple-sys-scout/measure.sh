#!/usr/bin/env bash
# Measurement compiler: 72f23939c773ba0045aeed5d6afb9b025b6473a0.
# Upstream TypeScript: 050880ce59e30b356b686bd3144efe24f875ebc8.
# This is the bounded sys.ts-only measurement phase, after meter preparation:
# apply that compiler revision's stage3/apply.sh adaptations to the pinned
# upstream at /tmp/sys-scout/adapted, and build census and latent-census there.
# Build latent-census with stage3/census/latent/make_overlay.py's overlay.
# The preparation records apply.exit and latent-build.exit (zero on success).
# This preserves the measurement commands; only the historical fixed deadline
# is replaced with a reusable 39-minute budget (or SCOUT_DEADLINE_EPOCH).
source /workspace/adamic-tools/env.sh
deadline=${SCOUT_DEADLINE_EPOCH:-$(( $(date +%s) + 39 * 60 ))}
while ! test -f /tmp/sys-scout/apply.exit || ! test -f /tmp/sys-scout/latent-build.exit; do
  if test "$(date +%s)" -ge "$((deadline - 30))"; then exit 124; fi
  sleep 2
done
if test "$(cat /tmp/sys-scout/apply.exit)" -ne 0 || test "$(cat /tmp/sys-scout/latent-build.exit)" -ne 0; then exit 1; fi
unset LATENT_SPECULATIVE LATENT_MUTANT_NO_STUBS
entry=/tmp/sys-scout/adapted/src/compiler/sys.ts
run_bounded() {
  cap=$1; shift
  remaining=$((deadline - $(date +%s) - 20))
  if test "$remaining" -le 0; then return 124; fi
  if test "$cap" -gt "$remaining"; then cap=$remaining; fi
  timeout --kill-after=10s "$cap" "$@"
}
run_bounded 300 /tmp/sys-scout/census "$entry" /tmp/sys-scout/census.jsonl > /tmp/sys-scout/census.log 2>&1
printf '%s\n' "$?" > /tmp/sys-scout/census.exit
run_bounded 480 env LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=/tmp/sys-scout/first-times.jsonl /tmp/sys-scout/latent-census "$entry" /tmp/sys-scout/first.jsonl > /tmp/sys-scout/first.log 2>&1
printf '%s\n' "$?" > /tmp/sys-scout/first.exit
run_bounded 720 env LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=/tmp/sys-scout/full-times.jsonl /tmp/sys-scout/latent-census "$entry" /tmp/sys-scout/full.jsonl > /tmp/sys-scout/full.log 2>&1
printf '%s\n' "$?" > /tmp/sys-scout/full.exit
