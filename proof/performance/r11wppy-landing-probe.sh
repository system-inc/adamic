# #r11wppy: the landing probe at the head, then at main's invariance.go. Never merged.
set -u
for side in head main; do
	if [ $side = main ]; then git show HEAD~2:internal/lower/invariance.go > internal/lower/invariance.go; fi
	echo "=== $side"
	go test -count=1 -timeout 20m -v -run '^TestZZLandingProbe$' ./internal/oracle/ 2>&1 | grep -E '^PROBE|^(ok|FAIL|---)|panic' | cut -c1-700
done
echo VERDICT DONE
