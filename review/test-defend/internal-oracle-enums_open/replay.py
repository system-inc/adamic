#!/usr/bin/env python3
# Run from repository root. Each diff contains only one production change.
import json, pathlib, subprocess, tempfile
root = pathlib.Path.cwd()
evidence = root / "review/test-defend/internal-oracle-enums_open"
for item in json.loads((evidence / "plan.json").read_text()):
    scratch = pathlib.Path(tempfile.mkdtemp(prefix="defend-enum-"))
    source = item["file_line"].rsplit(":", 1)[0]
    original = subprocess.check_output(["git", "show", json.loads((evidence / "metadata.json").read_text())["base"] + ":" + source])
    target = scratch / source
    target.parent.mkdir(parents=True)
    target.write_bytes(original)
    subprocess.run(["git", "apply", "--unsafe-paths", "--directory=" + str(scratch), str(evidence / (item["id"] + ".diff"))], check=True)
    overlay = scratch / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {str(root / source): str(target)}}))
    print(item["id"], overlay)
