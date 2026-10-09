from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
source = root / "internal/lower/view_member_read.go"
original = source.read_text()
needle = "property.ViewWhere = l.program.Where(node)"
assert original.count(needle) == 1
try:
    source.write_text(original.replace(needle, 'property.ViewWhere = ""'))
    patch = subprocess.run(["git", "diff", "--", "internal/lower/view_member_read.go"], cwd=root, capture_output=True, text=True, check=True)
    (Path(__file__).parent / "blank-view-where.patch").write_text(patch.stdout)
    result = subprocess.run(["go", "test", "./internal/lower", "-run", "^TestViewDiagnosticP", "-count=1", "-v", "-timeout", "90s"], cwd=root, capture_output=True, text=True, timeout=180)
    output = result.stdout + result.stderr
    (Path(__file__).parent / "mutant.log").write_text(output)
    assert result.returncode != 0, "blank ViewWhere survived"
    for name in ["P17", "P48", "P64"]:
        assert "--- FAIL: TestViewDiagnostic" + name in output, output
    assert "unsupported collection contract" in output, output
    print("blank ViewWhere caught by P17, P48 and P64 source-location assertions")
finally:
    source.write_text(original)
