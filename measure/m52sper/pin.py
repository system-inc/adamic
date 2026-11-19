# Pins every crate in bench/game/rust-deps/Cargo.lock to the newest release crates.io published before
# 2025-02-01 that its dependents allow, with cargo update --precise, pass after pass until nothing moves,
# then prints each crate's version and publish time. Runs on the box, with network. Never lands.
import json, re, subprocess, sys, time, urllib.request

cutoff = "2025-02-01T00:00:00"
lock = "Cargo.lock"
cache = {}

def versions(name):
    if name not in cache:
        request = urllib.request.Request(f"https://crates.io/api/v1/crates/{name}/versions", headers={"User-Agent": "benchmark-lockfile-pinner"})
        with urllib.request.urlopen(request, timeout=60) as response:
            cache[name] = json.load(response)["versions"]
        time.sleep(1.1)
    return cache[name]

def packages():
    text = open(lock).read()
    found = []
    for block in text.split("[[package]]")[1:]:
        name = re.search(r'^name = "(.+)"', block, re.M).group(1)
        version = re.search(r'^version = "(.+)"', block, re.M).group(1)
        if 'source = "registry+' in block:
            found.append((name, version))
    return found

def parse(version):
    core = version.split("+")[0]
    if "-" in core:
        return None
    return tuple(int(part) for part in core.split("."))

def compatible(a, b):
    if a[0] != b[0]:
        return False
    if a[0] > 0:
        return True
    if a[1] != b[1]:
        return False
    return a[1] > 0 or a[2] == b[2]

for round in range(1, 8):
    moved = 0
    for name, current in packages():
        now = parse(current)
        candidates = []
        for each in versions(name):
            parsed = parse(each["num"])
            if parsed is None or each["yanked"] or each["created_at"] >= cutoff or not compatible(parsed, now):
                continue
            candidates.append((parsed, each["num"]))
        if not candidates:
            print(f"round {round}: {name} {current}: no release of its line before the cutoff")
            continue
        target = max(candidates)[1]
        if parse(target) >= now:
            continue
        result = subprocess.run(["cargo", "update", "-p", f"{name}@{current}", "--precise", target], capture_output=True, text=True)
        if result.returncode == 0:
            moved += 1
            print(f"round {round}: {name} {current} -> {target}")
        else:
            print(f"round {round}: {name} {current} -> {target} refused: {result.stderr.strip().splitlines()[-1] if result.stderr.strip() else result.returncode}")
    if moved == 0:
        break

late = 0
for name, current in sorted(packages()):
    stamp = next((each["created_at"] for each in versions(name) if each["num"] == current), "unknown")
    mark = "" if stamp < cutoff else "  after the cutoff"
    late += stamp >= cutoff
    print(f"crate {name} {current} published {stamp[:10]}{mark}")
print(f"crates after the cutoff: {late}")
sys.exit(1 if late else 0)
