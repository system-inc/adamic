#!/usr/bin/env python3
"""Whether a rerun of some units lands on a run's kept verdicts (Kirk, Oct 8 21:17 MDT: a test killed at 90 s is fixed,
then only it reruns, plus what its change can reach by input hash; every other unit's verdict from the same run stands).

Both records list every unit as {id, input_sha256, verdict}, Loom's definition agreed Oct 9 (#apf5vdb, #hpjftdj):
input_sha256 = sha256 of canonical JSON {shared, own, toolchain, run}, where shared is the tree of every path except
_test.go files and testdata dirs plus the submodule commits, and own is the unit's packages' test files and testdata
(plus cloud/fast-gate/test-reads.json). So a non-test change moves every unit, and a test-only fix moves only its own.

A unit the rerun ran carries its verdict; a unit it kept carries "kept". The rerun lands when:
- its sha descends from the base's;
- every unit it ran passed;
- every kept unit is in the base with the same input hash and passed there (one whose inputs moved, or that failed,
  had to be rerun);
- every unit that didn't pass in the base was rerun, or no longer exists (a killed test split into new units, which
  are new ids and so were run).

usage: rerun_merge.py <base record.json> <rerun record.json>   (in the repository, for the ancestry check)
"""
import json
import subprocess
import sys


def problems(base, rerun, descends):
    found = []
    if not descends:
        found.append("its sha %s doesn't descend from the base run's %s" % (rerun.get("sha"), base.get("sha")))
    baseUnits = {unit["id"]: unit for unit in base.get("units") or []}
    rerunUnits = {unit["id"]: unit for unit in rerun.get("units") or []}
    if not baseUnits or not rerunUnits:
        return found + ["a rerun needs every unit listed in both records (base %d, rerun %d)" % (len(baseUnits), len(rerunUnits))]
    for name, unit in sorted(rerunUnits.items()):
        if unit.get("verdict") != "kept":
            if unit.get("verdict") != "passed":
                found.append("rerun unit %s is %s" % (name, unit.get("verdict")))
            continue
        kept = baseUnits.get(name)
        if kept is None:
            found.append("unit %s is kept but the base run never ran it" % name)
        elif kept.get("input_sha256") != unit.get("input_sha256"):
            found.append("unit %s is kept but its inputs changed (%s to %s), so it must rerun"
                         % (name, str(kept.get("input_sha256"))[:12], str(unit.get("input_sha256"))[:12]))
        elif kept.get("verdict") != "passed":
            # The base's red or killed units: kept is refused here, rerun is judged above, and gone (split into new
            # ids, which ran) is fine.
            found.append("unit %s is kept but was %s in the base run" % (name, kept.get("verdict")))
    return found


def descends(base, rerun):
    return subprocess.run(["git", "merge-base", "--is-ancestor", base, rerun], capture_output=True).returncode == 0


def main(arguments):
    with open(arguments[0]) as baseFile, open(arguments[1]) as rerunFile:
        base, rerun = json.load(baseFile), json.load(rerunFile)
    found = problems(base, rerun, descends(base.get("sha", ""), rerun.get("sha", "")))
    print("; ".join(found) if found else "rerun lands on %s's kept verdicts" % str(base.get("sha"))[:12])
    return 1 if found else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
