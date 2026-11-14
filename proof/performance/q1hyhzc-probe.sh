# #q1hyhzc's probe: why the dirty-name mutant survives. Prints each view walk's state, baseline and mutant. Never merged.
set -u
. "$HOME/adamic-tools/env.sh" > /dev/null 2>&1 || { echo "VERDICT FAIL env"; exit 3; }
git -c protocol.file.allow=always submodule update --init --recursive --depth 1 > /tmp/q1p-sub.log 2>&1 || { echo "VERDICT FAIL submodules"; exit 3; }
echo "head $(git rev-parse HEAD^) on $(hostname)"
f=internal/lower/view_sites.go
perl -0pi -e 's/import \(\n\t"fmt"/import (\n\t"fmt"\n\t"os"/' $f
perl -0pi -e 's/(\twholeFields, wholePlain := l.result.CheckedFields, l.result.PlainReceivers\n)/$1\tfmt.Fprintf(os.Stderr, "PROBE view %s fields %v writes %v seen %v keys %d whole+ %s\\n", l.program.Where(node), fields, before.checkedFieldWrites, before.viewSiteSeen, len(before.checkedFields), viewWalkOnly(wholeFields, before.checkedFields, func(key string) string { return key }))\n/' $f
perl -0pi -e 's/(\tgetters, refused := l.viewWalkIndexed\(modules, fields\)\n)/$1\tfmt.Fprintf(os.Stderr, "PROBE   indexed+ %s seenAfter %v\\n", viewWalkOnly(l.result.CheckedFields, before.checkedFields, func(key string) string { return key }), l.viewSiteSeen)\n/' $f
grep -c PROBE $f
run() { go test -count=1 -v -run '^TestViewSiteAskedAgainWhenItsNameGainsAKey$' ./internal/lower/ > /tmp/q1p-$1.log 2>&1; echo "== $1 exit $?"; grep -E 'PROBE|^--- |^ok|^FAIL|_test.go:[0-9]+:|cannot|undefined' /tmp/q1p-$1.log | cut -c1-420 | head -14; }
run baseline
sed -i 's/asked = !judged || seen != written\[name\]/asked = !judged/' $f
run mutant
git checkout -q -- $f
echo "VERDICT PROBED"
