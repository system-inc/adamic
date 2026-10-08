"""Run this rule's mutant as the sanitized native canary without editing shared tests."""
import json
import subprocess
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[6]
SOURCE = ROOT / "stage1/cohere/lint/lint_test.go"
OLD = 'const nativeCanaryRule = "react/jsx-no-comment-textnodes"'
NEW = 'const nativeCanaryRule = "constructor-super"'
with tempfile.TemporaryDirectory(prefix="constructor-super-native-mutant-") as directory:
    directory = Path(directory)
    text = SOURCE.read_text()
    assert text.count(OLD) == 1, "shared native canary anchor changed"
    patched = directory / "lint_test.go"
    proof = r"""
func TestConstructorSuperParity(t *testing.T) {
    directory, err := filepath.Abs(".")
    if err != nil { t.Fatal(err) }
    oracle := goOracle(t)
    var rows []string
    counts := map[string]int{}
    for _, row := range upstream(t) {
        fields := strings.Split(row, "\t")
        if len(fields) > 1 { counts[fields[1]]++ }
        if len(fields) > 1 && fields[1] == "constructor-super" { rows = append(rows, row) }
    }
    t.Logf("constructor-super: %d unique captured upstream cases", counts["constructor-super"])
    for _, descriptor := range prepareRegistry(t, ".") {
        if descriptor.Name == "constructor-super" { rows = append(rows, ownedWitnessRows(t, directory, descriptor)...) }
    }
    path := manifest(t, recoveryRows(t, oracle, rows))
    compare(t, oracle, buildPort(t, directory, true), directory, path)
}
"""
    patched.write_text(text.replace(OLD, NEW) + proof)
    overlay = directory / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {str(SOURCE): str(patched)}}))
    subprocess.run([
        "go", "test", "-overlay=" + str(overlay), "./stage1/cohere/lint",
        "-run", "^TestConstructorSuperParity$|^TestMutants$/^constructor-super-conditional-call-guaranteed$",
        "-count=1", "-v", "-timeout=10m",
    ], cwd=ROOT, check=True)
