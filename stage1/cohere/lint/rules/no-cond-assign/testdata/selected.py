"""Select this port and its native mutant through an ephemeral Go overlay."""
import json
import subprocess
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[6]
SOURCE = ROOT / "stage1/cohere/lint/lint_test.go"
REGISTRATION = ROOT / "stage1/cohere/lint/registration_test.go"
TEST = r'''
func TestNoCondAssignPort(t *testing.T) {
    t.Parallel()
    directory, err := filepath.Abs(".")
    if err != nil { t.Fatal(err) }
    oracle := goOracle(t)
    var rows []string
    count := 0
    for _, row := range upstream(t) {
        fields := strings.Split(row, "\t")
        if len(fields) > 1 && fields[1] == "no-cond-assign" {
            rows = append(rows, row)
            count++
        }
    }
    t.Logf("no-cond-assign upstream cases: %d", count)
    for _, descriptor := range prepareRegistry(t, ".") {
        if descriptor.Name == "no-cond-assign" {
            witnesses := ownedWitnessRows(t, directory, descriptor)
            for _, row := range witnesses {
                answer := execute(t, "", oracle, "--manifest", manifest(t, []string{row}), "--count")
                if string(answer.output) == "0\n" { t.Fatal("witness has no finding: " + row) }
                rows = append(rows, row)
                rows = append(rows, strings.SplitN(row, "\t", 2)[0] + "\tall")
            }
            t.Logf("no-cond-assign witnesses: %d", len(witnesses))
        }
    }
    rows = append(rows, generated(t)...)
    compare(t, oracle, buildPort(t, directory, true), directory, manifest(t, recoveryRows(t, oracle, rows)))
}
'''
with tempfile.TemporaryDirectory(prefix="no-cond-assign-selected-") as directory:
    directory = Path(directory)
    old = 'const nativeCanaryRule = "react/jsx-no-comment-textnodes"'
    text = SOURCE.read_text()
    assert text.count(old) == 1
    lint = directory / "lint_test.go"
    lint.write_text(text.replace(old, 'const nativeCanaryRule = "no-cond-assign"'))
    registration = directory / "registration_test.go"
    registration.write_text(REGISTRATION.read_text() + TEST)
    overlay = directory / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {str(SOURCE): str(lint), str(REGISTRATION): str(registration)}}))
    subprocess.run([
        "go", "test", "-overlay=" + str(overlay), "./stage1/cohere/lint",
        "-run", "^TestNoCondAssignPort$|^TestMutants/conditional-assignment-ternary-parentheses$",
        "-count=1", "-v", "-timeout=30m",
    ], cwd=ROOT, check=True)
