"""Summarizes a go test -json stream for gate-logs.sh: failures with their output, and a status line.

usage: gate-logs-summary.py <sha> <running|finished> <test.jsonl> <failures.txt> <status.txt>
"""
import collections
import json
import os
import sys

sha, state, log, failuresPath, statusPath = sys.argv[1:6]
output = collections.defaultdict(list)
buildOutput = collections.defaultdict(list)
failedTests, failedBuilds, failedPackages, finishedPackages = [], set(), set(), set()
counts = collections.Counter()
for line in open(log, errors="replace"):
    try:
        event = json.loads(line)
    except ValueError:
        continue
    action = event.get("Action")
    package = event.get("Package") or (event.get("ImportPath") or "?").split(" ")[0]
    test = event.get("Test")
    if action == "build-output":
        buildOutput[package].append(event.get("Output") or "")
    elif action == "output":
        output[(package, test)].append(event.get("Output") or "")
    elif action == "build-fail":
        failedBuilds.add(package)
    elif action in ("pass", "fail", "skip"):
        if test:
            counts[action] += 1
            if action == "fail":
                failedTests.append((package, test))
        else:
            finishedPackages.add(package)
            if action == "fail":
                if event.get("FailedBuild"):
                    failedBuilds.add(package)
                else:
                    failedPackages.add(package)

with open(failuresPath, "w") as failures:
    for package in sorted(failedBuilds):
        failures.write(f"== {package} (build failed)\n")
        failures.writelines(buildOutput[package][-20:])
        failures.write("\n")
    for package, test in failedTests:
        failures.write(f"== {package} {test}\n")
        failures.writelines(output[(package, test)][-20:])
        failures.write("\n")
    for package in sorted(failedPackages - failedBuilds):
        if not any(failed[0] == package for failed in failedTests):
            failures.write(f"== {package} (failed outside a test)\n")
            failures.writelines(output[(package, None)][-20:])
            failures.write("\n")

with open(statusPath, "w") as status:
    status.write(
        f"{state}: {counts['pass']} pass, {counts['fail']} fail, {counts['skip']} skip; "
        f"{len(failedBuilds)} packages failed to build; {len(finishedPackages)} packages finished\n"
    )
    status.write(f"sha {sha}\n")
    status.write(os.environ.get("GATE_LOGS_SIGNALS", "inherited ignored signals: not recorded") + "\n")
