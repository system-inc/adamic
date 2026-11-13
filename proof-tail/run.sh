#!/bin/bash
# Proof for #sx8wdg7's tail switch: compiler_test whole, then each named mutant red.
set -u
cd "$(dirname "$0")/.." || exit 9
status=0
case=CompilerSelectionTests.test_daily_runs_both_latent_passes_with_scoped_tail_sx8wdg7
run() { (cd stage3/meter && timeout 1500 python3 -B -m unittest "$@" 2>&1); }
echo "== compiler_test whole"
out=$(run compiler_test); rc=$?
echo "$out" | tail -n 6
[ $rc -eq 0 ] && echo "PASS suite" || { echo "FAIL suite rc=$rc"; echo "$out" | grep -E '^(FAIL|ERROR):'; status=1; }
f=stage3/meter/twice-daily.sh
mutate() { # name before after
  cp "$f" "$f.orig"
  python3 -I -c 'import sys;p,b,a=sys.argv[1:];s=open(p).read();assert s.count(b)==1,b;open(p,"w").write(s.replace(b,a))' "$f" "$2" "$3" || { echo "FAIL $1: did not apply"; status=1; mv "$f.orig" "$f"; return; }
  out=$(run "compiler_test.$case"); rc=$?
  if [ $rc -ne 0 ] && echo "$out" | grep -q "^FAIL: "; then echo "PASS $1 red"; echo "$out" | grep -m1 AssertionError; else echo "FAIL $1 survived rc=$rc"; echo "$out" | tail -n 4; status=1; fi
  mv "$f.orig" "$f"
}
mutate entry-no-tail 'LATENT_TAIL=1 "$binaries/entry-census"' '"$binaries/entry-census"'
mutate file-no-tail 'LATENT_TAIL=1 "$binaries/latent-census"' '"$binaries/latent-census"'
mutate census-leak '"$binaries/census" "$entry"' 'LATENT_TAIL=1 "$binaries/census" "$entry"'
git diff --quiet || { echo "FAIL tree left modified"; status=1; }
echo "VERDICT $([ $status -eq 0 ] && echo PASS || echo FAIL)"
exit $status
