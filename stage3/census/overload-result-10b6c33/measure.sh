#!/bin/bash
set -eu
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/overload-upstream
label=$1
out=/workspace/overload-build-$label
exec 9>"$out/measure.lock"
set +e
flock 9
rc=$?
echo "lock_child_exit=$rc"
test "$rc" = 0 || exit "$rc"
if [[ -f $out/latent.time.json ]]; then
 python3 - "$out" <<'CHECK'
import json,pathlib,sys
p=pathlib.Path(sys.argv[1]); stats=json.loads((p/'latent.time.json').read_text())
with (p/'latent.jsonl').open() as f:
 for line in f: last=json.loads(line)
assert stats['exit_code']==0 and last['status']=='relation_report'
print('reused completed measurement; original child_exit='+str(stats['exit_code']))
CHECK
 rc=$?
 echo "validation_child_exit=$rc"
 exit "$rc"
fi
export LATENT_ASSERT_NO_OUTPUT=1 LATENT_FULL=0 LATENT_TIMES="$out/latent-times.jsonl"
unset LATENT_SPECULATIVE LATENT_MUTANT_FUNCTION LATENT_ONLY_FILE LATENT_ONLY_UNIT
set +e
python3 /workspace/overload-run.py "$out/latent.log" "$out/latent.time.json" "$out/entry-census" /workspace/overload-adapted/src/tsc/tsc.ts "$out/latent.jsonl"
rc=$?
echo "entry_census_child_exit=$rc"
exit "$rc"
