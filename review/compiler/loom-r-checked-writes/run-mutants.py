import json, pathlib, subprocess, time, sys
out = pathlib.Path(__file__).resolve().parent
root = out.parents[2]
results = []
for item in json.loads((out / "mutants.json").read_text()):
    command = ["timeout", "120", "go", "test", "-overlay", item["overlay"], item["package"], "-run", item["selector"], "-count=1", "-parallel=4", "-v", "-timeout=90s"]
    start = time.monotonic()
    path = out / (item["name"] + ".log")
    with path.open("w") as log:
        result = subprocess.run(command, cwd=root, stdout=log, stderr=subprocess.STDOUT)
    body = path.read_text()
    caught = result.returncode == 1 and "--- FAIL:" in body and "error: " not in body and "build failed" not in body and "test timed out" not in body
    entry = dict(item, command=command, exit=result.returncode, caught=caught, seconds=round(time.monotonic() - start, 3))
    results.append(entry)
    (out / "mutant-results.json").write_text(json.dumps(results, indent=2) + "\n")
    print(item["name"], "caught" if caught else "NOT CAUGHT", entry["seconds"], flush=True)
    if not caught: sys.exit(1)
