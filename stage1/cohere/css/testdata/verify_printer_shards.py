#!/usr/bin/env python3
"""Verify all printer top-level leaves: 60s budget, hard 75s kills.

Run from the repository root with fixture/library inputs enabled. Build products
precede M.Run and are inputs to the timed leaves. JSON goes to stdout; keep
measurement output outside the repository.
"""
import json
import os
import selectors
import signal
import subprocess
import sys
import time


def kill_tree(process):
    parents = []
    for entry in os.listdir("/proc"):
        if not entry.isdigit():
            continue
        try:
            with open(f"/proc/{entry}/stat") as file:
                fields = file.read().rsplit(") ", 1)[1].split()
            parents.append((int(entry), int(fields[1])))
        except (FileNotFoundError, ProcessLookupError):
            pass
    descendants = {process.pid}
    while True:
        children = {pid for pid, parent in parents if parent in descendants}
        if children <= descendants:
            break
        descendants.update(children)
    for pid in sorted(descendants, reverse=True):
        try:
            os.kill(pid, signal.SIGKILL)
        except ProcessLookupError:
            pass
    process.wait()


def main():
    process = subprocess.Popen(
        ["go", "test", "-json", "-run", "^TestCSSPrinterAgreesWithGo_[0-9]{3}$",
         "-count=1", "-timeout=0", "./stage1/cohere/css"],
        stdout=subprocess.PIPE, start_new_session=True)
    selector = selectors.DefaultSelector()
    selector.register(process.stdout, selectors.EVENT_READ)
    active, elapsed, cooked = {}, {}, []
    buffer = b""
    try:
        while selector.get_map():
            now = time.monotonic()
            expired = [name for name, start in active.items() if now - start >= 75]
            if expired:
                cooked.extend(expired)
                for name in expired:
                    print(json.dumps({"Action": "cooked", "Test": name,
                                      "Elapsed": now - active[name]}), flush=True)
                kill_tree(process)
                break
            wait = min([1.0] + [max(0, 75 - (now - start)) for start in active.values()])
            for key, _ in selector.select(wait):
                data = os.read(key.fd, 65536)
                if not data:
                    selector.unregister(key.fileobj)
                    continue
                buffer += data
                while b"\n" in buffer:
                    line, buffer = buffer.split(b"\n", 1)
                    event = json.loads(line)
                    print(line.decode(), flush=True)
                    name = event.get("Test", "")
                    if not name.startswith("TestCSSPrinterAgreesWithGo_"):
                        continue
                    if event["Action"] == "cont":
                        active[name] = time.monotonic()
                    elif event["Action"] in ("pass", "fail", "skip"):
                        active.pop(name, None)
                        if event["Action"] != "skip":
                            elapsed[name] = event.get("Elapsed", 0)
        returncode = process.wait()
    finally:
        selector.close()
        if process.poll() is None:
            kill_tree(process)
    over_budget = [name for name, seconds in elapsed.items() if seconds > 60]
    print(json.dumps({"leaves": len(elapsed), "cooked": cooked,
                      "over_budget": over_budget, "go_exit": returncode}), file=sys.stderr)
    if cooked:
        return 75
    if returncode or over_budget:
        return returncode or 60
    expected = 1 if os.getenv("ADAMIC_TEST_SHARD") else 512
    if len(elapsed) != expected:
        print(f"observed {len(elapsed)} leaves, expected {expected}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
