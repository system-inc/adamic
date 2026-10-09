import json, pathlib, subprocess, sys, time
p = pathlib.Path(__file__).resolve().parent
names = json.loads((p / "leaves.json").read_text())
index = int(sys.argv[1])
selected = names[index * 80:(index + 1) * 80]
assert selected
command = ["go", "test", "./internal/flow", "-run", "^(" + "|".join(selected) + ")$", "-count=1", "-timeout", "90s", "-json"]
began = time.monotonic()
with (p / f"corpus-{index:02}.jsonl").open("w") as log:
    result = subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, timeout=85)
status = {"command": command, "selected": selected, "exit": result.returncode, "wall_seconds": round(time.monotonic() - began, 3)}
(p / f"corpus-{index:02}.status.json").write_text(json.dumps(status, indent=2) + "\n")
print(index, len(selected), status["wall_seconds"], result.returncode)
sys.exit(result.returncode)
