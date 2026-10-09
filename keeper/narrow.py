"""narrow.py: the packages a mutant must replay on, read from the line map (#mp71kkr, @system_adamic_tests).

	narrow.py <cover dir> <deps.json> <mutant.diff>...
	# one line per diff, tab-separated: the diff, why, the packages a profile proves reach it, then the packages added only
	# because their line is unknown (unmapped parts, or a mutant coverage can't see)

The map is ~/.loom/cover/<sha>/: job.json names each unit's parts (argv after the sha, "<package>=<run regex>", in
part order), units/<unit>/part-N.cover is a Go set-mode profile of what part N's tests executed. A diff is read at the
map's sha: its changed lines are the old side's removed lines, or for a pure insertion the old line it lands after.

A part goes into the replay when:
- its profile executed a block holding a changed line (the map's whole point), or
- its line is unknown: the unit isn't fetched, or the part left no profile (it timed out under coverage), and its
  package depends on the mutated package (deps.json, from go list -deps at the map's sha), or
- the mutant isn't Go (the C runtime, a stage1 port's .ts or .a): coverage can't see it, so its home package and every
  package that depends on it, and the profiles say nothing.
"""
import json, os, re, sys

module = "github.com/system-inc/adamic/"
cover, depsPath, diffs = sys.argv[1], sys.argv[2], sys.argv[3:]
job = json.load(open(os.path.join(cover, "job.json")))
deps = json.load(open(depsPath))  # import path -> every import path it depends on, itself excluded

parts = []  # (unit, index, package)
for unit in job["units"]:
    specs = [argument for argument in unit["argv"][5:] if "=" in argument]
    for index, spec in enumerate(specs):
        parts.append((unit["id"], index, spec.split("=", 1)[0]))

profiles = {}  # (unit, index) -> {file: [(start, end)]} of executed blocks, or None when the part left no profile
for unit, index, package in parts:
    path = os.path.join(cover, "units", unit, "part-%d.cover" % index)
    if not os.path.exists(path):
        profiles[(unit, index)] = None
        continue
    blocks = {}
    for line in open(path):
        found = re.match(r"(\S+):(\d+)\.\d+,(\d+)\.\d+ \d+ (\d+)$", line.strip())
        if found and found.group(4) != "0":
            blocks.setdefault(found.group(1), []).append((int(found.group(2)), int(found.group(3))))
    profiles[(unit, index)] = blocks


def changedLines(diff):
    """{repository path: set of old-side line numbers} for one unified diff."""
    changed, path, old = {}, None, 0
    for line in open(diff):
        if line.startswith("--- "):
            path = re.sub(r"^(a/)?", "", line[4:].strip().split("\t")[0])
        elif line.startswith("+++ "):
            continue
        elif line.startswith("@@"):
            old = int(re.match(r"@@ -(\d+)", line).group(1))
        elif path and line.startswith("-"):
            changed.setdefault(path, set()).add(old)
            old += 1
        elif path and line.startswith("+"):
            changed.setdefault(path, set()).add(max(old - 1, 1))
        elif path and line.startswith(" "):
            old += 1
    return changed


def dependents(package):
    return {name for name, needed in deps.items() if package in needed or name == package}


for diff in diffs:
    targets, unknown, why = set(), set(), set()
    for path, lines in changedLines(diff).items():
        home = module + os.path.dirname(path)
        if path.endswith(".go") and not path.endswith("_test.go") and re.match(r"(internal|bridge|cmd)/", path):
            reach = dependents(home)
            for unit, index, package in parts:
                blocks = profiles[(unit, index)]
                if blocks is None:
                    if package in reach:
                        unknown.add(package)
                        why.add("unmapped")
                    continue
                if any(start <= line <= end for start, end in blocks.get(module + path, []) for line in lines):
                    targets.add(package)
                    why.add("covered")
        else:
            # The C runtime is internal/native's; a port's source is its own directory's package.
            if path.startswith("internal/native/runtime/"):
                home = module + "internal/native"
            reach = dependents(home)
            unknown |= {package for unit, index, package in parts if package in reach}
            why.add("not Go: home and dependents")
    print("%s\t%s\t%s\t%s" % (os.path.basename(diff), ",".join(sorted(why)) or "nothing reaches it", " ".join(sorted(targets)), " ".join(sorted(unknown - targets))))
