set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #sbvnh96: which mirrored pair moves TestCycleFinderEquality. DROP=graph puts graph_types.go:492 and the reference's
# graphLinks back to the fresh type; DROP=cycles puts cycles.go:836 and the reference's reaches back.
echo "box: $(hostname) at $(git rev-parse HEAD) DROP=${DROP:-none}"
sandboxed_typescript_source > /dev/null 2>&1
python3 -I - "${DROP:-none}" <<'PY'
import sys
drop = sys.argv[1]
pairs = {
    "graph": [("internal/lower/graph_types.go", "IsTypeAssignableTo(f.l.relatedShape(shape), proven) {\n\t\t\t\tlinks = append(links, cycleNode{proven: shape})"),
              ("internal/lower/cycles_scanning_reference_test.go", "IsTypeAssignableTo(f.l.relatedShape(shape), proven) {\n\t\t\t\tlinks = append(links, cycleNode{proven: shape})")],
    "cycles": [("internal/lower/cycles.go", "IsTypeAssignableTo(f.l.relatedShape(shape), proven) {\n\t\t\t\tlinks = append(links, cycleNode{proven: shape})"),
               ("internal/lower/cycles_scanning_reference_test.go", "IsTypeAssignableTo(f.l.relatedShape(shape), proven) {\n\t\t\t\t\tqueue = append(queue, cycleNode{proven: shape})")],
}
for path, old in pairs.get(drop, []):
    text = open(path).read()
    if text.count(old) != 1:
        print(f"BISECT FAIL {path}: {text.count(old)} matches"); sys.exit(1)
    open(path, "w").write(text.replace(old, old.replace("f.l.relatedShape(shape)", "shape")))
    print(f"BISECT dropped {path}")
PY
go test -count=1 -run 'TestCycleFinderEquality' ./internal/lower/ > "$TMPDIR/eq.log" 2>&1; code=$?
grep -E "^(--- FAIL|    --- FAIL)|difference:" "$TMPDIR/eq.log" | head -20
echo "BISECT DROP=${DROP:-none} exit=$code"
