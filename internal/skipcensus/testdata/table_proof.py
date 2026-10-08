"""Historical conversion proof through the integration command, run at repo root."""
import json
import subprocess
import sys

revision = sys.argv[1] if len(sys.argv) > 1 else "1db5ab2b1506e699233022eefba0ee7110fe24ce"
old = json.loads(subprocess.check_output([
    "git", "show", revision + ":internal/skipcensus/testdata/skips.json"
]))
command = ["go", "run", "./internal/skipcensus/cmd", "-table"]
output = subprocess.check_output(command)
assert output == subprocess.check_output(command), "unstable table output"
rows = json.loads(output)
def key(row):
    return tuple(row[field] for field in ("file", "test", "condition", "message"))
previous = {key(row): row for row in old}
assert len(previous) == len(old) == len(rows)
assert rows == sorted(rows, key=lambda row: (row["file"], row["line"]))
for row in rows:
    assert all(field in row for field in ("test", "callers", "class", "provides", "file", "line"))
    original = previous.pop(key(row))
    assert row["class"] == original["class"], key(row)
    assert row["provides"] == original["provides"], key(row)
assert not previous
print(f"go run ./internal/skipcensus/cmd -table: {len(rows)} rows; exact class/provider equality against {revision}; stable JSON array")
