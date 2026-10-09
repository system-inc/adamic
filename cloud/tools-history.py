#!/usr/bin/env python3
"""Append-only promotion history and serialized serve-time tools snapshots."""
import argparse
import fcntl
import json
import os
from pathlib import Path
import tempfile
import time


def rows(path):
    if not path.exists():
        return []
    result = []
    for line in path.read_text().splitlines():
        sha, stamp = line.split("\t")
        result.append((sha, int(stamp)))
    if any(a[1] >= b[1] for a, b in zip(result, result[1:])):
        raise ValueError("promotion history is not strictly ordered")
    return result


def was_live(history, record):
    """Reject unserved tools and tools promoted only after this job was served."""
    served = record.get("tools_served_at")
    if not isinstance(served, int) or isinstance(served, bool):
        return False
    return any(sha == record.get("tools_sha") and stamp <= served
               for sha, stamp in rows(Path(history)))


def snapshot(state, promote=None):
    state = Path(state)
    state.mkdir(parents=True, exist_ok=True)
    with (state / "tools-history.lock").open("a") as lock:
        fcntl.flock(lock, fcntl.LOCK_EX)
        history = state / "tools-good-history.tsv"
        previous = rows(history)
        good = promote if promote is not None else (state / "tools-good").read_text().strip()
        if not good or any(c.isspace() for c in good):
            raise ValueError("missing or malformed tools-good")
        now = max(time.time_ns(), previous[-1][1] + 1 if previous else 0)
        if not previous or previous[-1][0] != good:
            with history.open("a") as output:
                output.write(f"{good}\t{now}\n")
                output.flush()
                os.fsync(output.fileno())
        if promote is not None:
            with tempfile.NamedTemporaryFile(mode="w", dir=state, delete=False) as output:
                output.write(good + "\n")
                temporary = output.name
            os.replace(temporary, state / "tools-good")
        return good, max(time.time_ns(), now + 1)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("operation", choices=["serve", "promote", "accept"])
    parser.add_argument("state")
    parser.add_argument("value", nargs="?")
    args = parser.parse_args()
    if args.operation == "accept":
        record = json.loads(Path(args.value).read_text())
        raise SystemExit(0 if was_live(Path(args.state) / "tools-good-history.tsv", record) else 1)
    good, served = snapshot(args.state, args.value if args.operation == "promote" else None)
    print(good, served)


if __name__ == "__main__":
    main()
