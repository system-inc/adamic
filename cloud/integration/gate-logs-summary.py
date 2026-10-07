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
# The skip census's table classifies every skip site: required-input, measurement or
# not-applicable, under the function holding the skip and every test that reaches it (callers).
# Matching a log's skip by test name is an early mark, not the verdict: the census command
# identifies each skip by its source condition, and its exit decides the gate.
skipClasses = collections.defaultdict(set)
skipProvides = {}
try:
    for row in json.load(open(os.environ.get("GATE_LOGS_SKIP_TABLE", ""))):
        for name in [row["test"], *row.get("callers", [])]:
            skipClasses[name].add(row["class"])
            if row["class"] == "required-input":
                skipProvides.setdefault(name, row.get("provides", ""))
    haveTable = True
except (OSError, ValueError, KeyError, TypeError):
    haveTable = False
skippedForInput, skippedUnknown = [], []
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
            if action == "skip" and haveTable:
                top = test.split("/")[0]
                classes = skipClasses.get(test) or skipClasses.get(top)
                if not classes:
                    skippedUnknown.append((package, test))
                elif classes == {"required-input"}:
                    skippedForInput.append((package, test, skipProvides.get(test) or skipProvides.get(top, "")))
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
    for package, test, provides in skippedForInput:
        failures.write(f"== {package} {test} (skipped; the census table classes it required-input)\n")
        failures.write(f"census: {provides}\n")
        failures.writelines(output[(package, test)][-5:])
        failures.write("\n")
    for package, test in skippedUnknown:
        failures.write(f"== {package} {test} (skipped; no census row matches its name)\n")
        failures.writelines(output[(package, test)][-5:])
        failures.write("\n")
    for package in sorted(failedPackages - failedBuilds):
        if not any(failed[0] == package for failed in failedTests):
            failures.write(f"== {package} (failed outside a test)\n")
            failures.writelines(output[(package, None)][-20:])
            failures.write("\n")

with open(statusPath, "w") as status:
    status.write(
        f"{state}: {counts['pass']} pass, {counts['fail']} fail, {counts['skip']} skip; "
        f"{len(failedBuilds)} packages failed to build; {len(finishedPackages)} packages finished; "
        + (f"by test name, {len(skippedForInput)} required-input skips and {len(skippedUnknown)} unmatched "
           f"(census table: {os.environ.get('GATE_LOGS_SKIP_SOURCE', '?')}; the census command decides)\n" if haveTable
           else "skips not judged: no census table\n")
    )
    status.write(f"sha {sha}\n")
    status.write(os.environ.get("GATE_LOGS_SIGNALS", "inherited ignored signals: not recorded") + "\n")
    status.write(os.environ.get("GATE_LOGS_INPUTS", "gate inputs: not recorded") + "\n")
