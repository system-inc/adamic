import json, pathlib, subprocess, time, sys
root = pathlib.Path(__file__).resolve().parents[3]
out = pathlib.Path(__file__).resolve().parent
results = []
for package, tests in json.loads((out / "focused-tests.json").read_text()).items():
    tests = list(dict.fromkeys(tests))
    for index in range(0, len(tests), 20):
        selected = tests[index:index + 20]
        label = package.removeprefix("./").replace("/", "-") + "-" + str(index // 20)
        command = ["timeout", "120", "go", "test", package, "-run", "^(" + "|".join(selected) + ")$", "-count=1", "-parallel=4", "-json", "-timeout=90s"]
        start = time.monotonic()
        with (out / (label + ".log")).open("w") as log:
            result = subprocess.run(command, cwd=root, stdout=log, stderr=subprocess.STDOUT)
        leaves = []
        for line in (out / (label + ".log")).read_text().splitlines():
            try: event = json.loads(line)
            except ValueError: continue
            if event.get("Action") in ["pass", "fail", "skip"] and event.get("Test"):
                leaves.append({"test": event["Test"], "action": event["Action"], "seconds": event.get("Elapsed")})
        entry = {"command": command, "exit": result.returncode, "seconds": round(time.monotonic() - start, 3), "tests": leaves}
        results.append(entry)
        (out / "focused-results.json").write_text(json.dumps(results, indent=2) + "\n")
        print(label, result.returncode, entry["seconds"], flush=True)
        if result.returncode or any(t["seconds"] and t["seconds"] >= 60 for t in leaves): sys.exit(1)
