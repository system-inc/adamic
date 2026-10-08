"""Historical conversion proof through the integration command, run at repo root."""
import json
import subprocess

old = json.loads(subprocess.check_output([
    "git", "show", "f8dc89f215febd29f2b523edfab987428b82776a:internal/skipcensus/testdata/skips.json"
]))
command = ["go", "run", "./internal/skipcensus/cmd", "-table"]
output = subprocess.check_output(command)
assert output == subprocess.check_output(command), "unstable table output"
rows = json.loads(output)
def key(row):
    return tuple(row[field] for field in ("file", "test", "condition", "message"))
previous = {key(row): row for row in old}
assert len(previous) == len(old) == len(rows) == 75
assert rows == sorted(rows, key=lambda row: (row["file"], row["line"]))
for row in rows:
    assert all(field in row for field in ("test", "callers", "class", "provides", "file", "line"))
    original = previous.pop(key(row))
    assert row["class"] == original["class"], key(row)
    assert row["provides"] == original["provides"], key(row)
assert not previous
print("go run ./internal/skipcensus/cmd -table: 75 rows; exact class/provider equality; stable JSON array")
