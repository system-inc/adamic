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
packageElapsed = {}
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
            packageElapsed[package] = event.get("Elapsed") or 0
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

# A package can fail with no failing test (a timeout, a panic outside a test, a build failure), so
# the status line counts failed packages beside failed tests and names each with how it ended; a run
# whose only failure was a package timeout must not read as "0 fail".
def short(package):
    return package.replace("github.com/system-inc/adamic/", "")

def ending(package):
    if package in failedBuilds:
        return "failed to build"
    # Go attributes the timeout panic to the test that was running, so read all of the package's output.
    said = "".join("".join(lines) for (owner, _), lines in output.items() if owner == package)
    seconds = f"{packageElapsed.get(package, 0):,.0f} s"
    if "test timed out after" in said:
        return f"timed out at {seconds}"
    if any(failed[0] == package for failed in failedTests):
        return f"failed tests ({seconds})"
    return f"failed outside a test at {seconds}"

failedPackageNames = sorted(failedPackages | failedBuilds)
packageLine = f"{len(failedPackageNames)} packages failed" + (": " + ", ".join(f"{short(p)} {ending(p)}" for p in failedPackageNames) if failedPackageNames else "")

with open(statusPath, "w") as status:
    status.write(
        f"{state}: {counts['pass']} pass, {counts['fail']} tests failed, {counts['skip']} skip; {packageLine}; "
        f"{len(finishedPackages)} packages finished; "
        + (f"by test name, {len(skippedForInput)} required-input skips and {len(skippedUnknown)} unmatched "
           f"(census table: {os.environ.get('GATE_LOGS_SKIP_SOURCE', '?')}; the census command decides)\n" if haveTable
           else "skips not judged: no census table\n")
    )
    status.write(f"sha {sha}\n")
    status.write(os.environ.get("GATE_LOGS_SIGNALS", "inherited ignored signals: not recorded") + "\n")
    status.write(os.environ.get("GATE_LOGS_INPUTS", "gate inputs: not recorded") + "\n")
