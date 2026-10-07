#!/usr/bin/env python3
"""Turns a coverage measurement's raw data into verify/coverage/REPORT.md and REPORT.json.

measure.sh runs it. It reads, for each side (the generator and the oracle fixtures):

  <side>/go.txt          Go coverage in the text profile format (go tool covdata textfmt, or
                         go test -coverprofile), blocks with their statement counts
  <side>/c.json          llvm-cov export of the runtime objects against <side>'s merged profile
  <side>/c.lcov          the same, as lcov, for per-line counts

Go's coverage counts blocks (each arm of an if, each case of a switch, is its own block) and the
statements in them; it has no branch coverage, so "branches" on the Go side are blocks. clang's
source-based coverage has regions, lines and branch outcomes (each condition true and false).

usage: analyze.py <measurement directory> <repository> <REPORT.md> <REPORT.json> <commands file>
"""

import json
import os
import re
import sys
from collections import defaultdict

module = "github.com/system-inc/adamic/"

# Go the fuzzer can't reach by construction: it drives the compiler through `adamic c` and `adamic
# js` and compiles the C with clang itself, so native.Build, the split build, the runtime cache and
# the tsgo paths (which `adamic c` refuses) run in the fuzzer's own process or not at all. These are
# listed apart from the blind map's ranking, not hidden.
structural_files = {"internal/native/units.go", "internal/native/units_tsgo.go", "internal/native/tsgo.go", "internal/native/library.go", "internal/lower/tsgo.go"}
structural_functions = {("internal/native/native.go", name) for name in ("Build", "Flags", "CoverageRequested")}


def structural(region):
    return region["file"] in structural_files or (region["file"], region["function"]) in structural_functions


# The Go groups the report holds shares for. Each is a predicate on the repository-relative path.
go_groups = [
    ("internal/native/emit*.go", lambda path: re.fullmatch(r"internal/native/emit(_[a-z]+)?\.go", path) is not None),
    ("internal/lower", lambda path: path.startswith("internal/lower/")),
    ("internal/native (all Go)", lambda path: path.startswith("internal/native/")),
    ("internal/native (Go, besides units*.go, tsgo.go, library.go)", lambda path: path.startswith("internal/native/") and path not in structural_files),
    ("internal/ir", lambda path: path.startswith("internal/ir/")),
    ("internal/flow", lambda path: path.startswith("internal/flow/")),
]

# The blind map's Go side: internal/lower and internal/native.
def go_blind_scope(path):
    return path.startswith("internal/lower/") or path.startswith("internal/native/")


def read_go_profile(path):
    """Blocks keyed by (file, start line, start column, end line, end column): [statements, count]."""
    blocks = {}
    with open(path) as handle:
        for line in handle:
            line = line.strip()
            if not line or line.startswith("mode:"):
                continue
            match = re.fullmatch(r"(.+):(\d+)\.(\d+),(\d+)\.(\d+) (\d+) (\d+)", line)
            if not match:
                continue
            name = match.group(1)
            if not name.startswith(module):
                continue
            file = name[len(module):]
            key = (file, int(match.group(2)), int(match.group(3)), int(match.group(4)), int(match.group(5)))
            statements, count = int(match.group(6)), int(match.group(7))
            if key in blocks:
                blocks[key][1] += count
            else:
                blocks[key] = [statements, count]
    return blocks


source_cache = {}


def source_lines(repository, file):
    if file not in source_cache:
        try:
            with open(os.path.join(repository, file), encoding="utf-8", errors="replace") as handle:
                source_cache[file] = handle.read().split("\n")
        except OSError:
            source_cache[file] = []
    return source_cache[file]


go_functions_cache = {}


def go_functions(repository, file):
    """Top-level functions of a gofmt'd Go file: (start line, end line, name)."""
    if file in go_functions_cache:
        return go_functions_cache[file]
    lines = source_lines(repository, file)
    functions = []
    start, name = None, None
    for number, text in enumerate(lines, 1):
        if text.startswith("func "):
            match = re.match(r"func (?:\(([^)]*)\) )?([A-Za-z_][A-Za-z0-9_]*)", text)
            receiver = ""
            if match and match.group(1):
                receiver = match.group(1).split()[-1].lstrip("*").split("[")[0] + "."
            name = receiver + (match.group(2) if match else "?")
            if text.rstrip().endswith("}"):
                functions.append((number, number, name))
                start = None
            else:
                start = number
        elif text == "}" and start is not None:
            functions.append((start, number, name))
            start = None
    go_functions_cache[file] = functions
    return functions


def enclosing_go_function(repository, file, line):
    for start, end, name in go_functions(repository, file):
        if start <= line <= end:
            return name, start, end
    return "?", line, line


def go_shares(blocks):
    shares = {}
    for label, predicate in go_groups:
        total_blocks = covered_blocks = total_statements = covered_statements = 0
        for key, (statements, count) in blocks.items():
            if not predicate(key[0]):
                continue
            total_blocks += 1
            total_statements += statements
            if count > 0:
                covered_blocks += 1
                covered_statements += statements
        shares[label] = {
            "blocks": total_blocks,
            "blocksReached": covered_blocks,
            "blockShare": round(covered_blocks / total_blocks, 4) if total_blocks else None,
            "statements": total_statements,
            "statementsReached": covered_statements,
            "statementShare": round(covered_statements / total_statements, 4) if total_statements else None,
        }
    return shares


def go_case_label(repository, file, line):
    lines = source_lines(repository, file)
    if 0 < line <= len(lines):
        text = lines[line - 1].strip()
        if text.startswith("case ") or text.startswith("default:"):
            return text[:110]
    return None


def enclosing_case(repository, file, line, function_start):
    """The case a line sits in, nearest first: the closest line above it, inside its function, that
    opens a case or default at a smaller indentation."""
    lines = source_lines(repository, file)
    if not 0 < line <= len(lines):
        return None
    indentation = len(lines[line - 1]) - len(lines[line - 1].lstrip("\t"))
    for number in range(line - 1, function_start - 1, -1):
        text = lines[number - 1]
        depth = len(text) - len(text.lstrip("\t"))
        stripped = text.strip()
        if depth < indentation and (stripped.startswith("case ") or stripped.startswith("default:")):
            return "in " + stripped[:100]
        if depth < indentation and stripped.startswith("func"):
            return None
    return None


def go_blind_map(repository, blocks):
    """Untouched Go regions in internal/lower and internal/native, largest first.

    A function no generated program enters is one region. Inside a function that ran, a region is a
    run of consecutive blocks that never ran, cut at each case and default so every switch arm stands
    on its own."""
    by_function = defaultdict(list)
    for key, (statements, count) in blocks.items():
        file = key[0]
        if not go_blind_scope(file) or file.endswith("_test.go"):
            continue
        name, start, end = enclosing_go_function(repository, file, key[1])
        by_function[(file, name, start, end)].append((key, statements, count))
    regions = []
    for (file, name, start, end), members in by_function.items():
        members.sort(key=lambda member: (member[0][1], member[0][2]))
        if all(count == 0 for _, _, count in members):
            statements = sum(statements for _, statements, _ in members)
            regions.append({"file": file, "start": start, "end": end, "function": name, "kind": "function never entered", "statements": statements, "blocks": len(members), "label": None, "keys": [member[0] for member in members]})
            continue
        run = []

        def close():
            if run:
                first, last = run[0][0], run[-1][0]
                regions.append({
                    "file": file, "start": first[1], "end": max(member[0][3] for member in run), "function": name,
                    "kind": "case never taken" if go_case_label(repository, file, first[1]) else "blocks never run",
                    "statements": sum(member[1] for member in run), "blocks": len(run),
                    "label": go_case_label(repository, file, first[1]) or enclosing_case(repository, file, first[1], start),
                    "keys": [member[0] for member in run],
                })
                run.clear()

        for member in members:
            key, statements, count = member
            if count > 0:
                close()
                continue
            if run and go_case_label(repository, file, key[1]) and key[1] > run[-1][0][3]:
                close()
            run.append(member)
        close()
    regions.sort(key=lambda region: (-region["statements"], region["file"], region["start"]))
    return regions


def read_c(side_directory, repository):
    with open(os.path.join(side_directory, "c.json")) as handle:
        export = json.load(handle)
    data = export["data"][0]
    runtime = os.path.join("internal", "native", "runtime")

    def relative(path):
        return os.path.join(runtime, os.path.basename(path))

    lines = defaultdict(dict)  # file -> line -> count
    current = None
    with open(os.path.join(side_directory, "c.lcov")) as handle:
        for text in handle:
            text = text.strip()
            if text.startswith("SF:"):
                current = relative(text[3:])
            elif text.startswith("DA:") and current:
                number, count = text[3:].split(",")[:2]
                lines[current][int(number)] = max(lines[current].get(int(number), 0), int(count))
    functions = {}  # (file, line, column) -> record
    regions = {}  # (file, l1, c1, l2, c2) -> [count, function key]
    branches = {}  # (file, l1, c1, l2, c2) -> [true count, false count]
    for function in data["functions"]:
        filenames = function["filenames"]
        code = [region for region in function["regions"] if region[7] == 0]
        if not code:
            continue
        body = code[0]
        file = relative(filenames[body[5]])
        key = (file, body[0], body[1])
        name = function["name"].split(":")[-1]
        record = functions.setdefault(key, {"file": file, "start": body[0], "end": body[2], "name": name, "count": 0})
        record["count"] += function["count"]
        for region in code:
            region_key = (relative(filenames[region[5]]), region[0], region[1], region[2], region[3])
            entry = regions.setdefault(region_key, [0, key])
            entry[0] += region[4]
        for branch in function.get("branches", []):
            branch_key = (relative(filenames[branch[6]]), branch[0], branch[1], branch[2], branch[3])
            entry = branches.setdefault(branch_key, [0, 0])
            entry[0] += branch[4]
            entry[1] += branch[5]
    totals = data["totals"]
    return {"lines": lines, "functions": functions, "regions": regions, "branches": branches, "totals": totals}


def c_shares(c):
    totals = c["totals"]
    branch_outcomes = 2 * len(c["branches"])
    branch_reached = sum((1 if true > 0 else 0) + (1 if false > 0 else 0) for true, false in c["branches"].values())
    line_total = sum(len(lines) for lines in c["lines"].values())
    line_reached = sum(1 for lines in c["lines"].values() for count in lines.values() if count > 0)
    function_total = len(c["functions"])
    function_reached = sum(1 for record in c["functions"].values() if record["count"] > 0)
    region_total = len(c["regions"])
    region_reached = sum(1 for count, _ in c["regions"].values() if count > 0)
    return {
        "internal/native/runtime": {
            "branchOutcomes": branch_outcomes, "branchOutcomesReached": branch_reached,
            "branchShare": round(branch_reached / branch_outcomes, 4) if branch_outcomes else None,
            "lines": line_total, "linesReached": line_reached,
            "lineShare": round(line_reached / line_total, 4) if line_total else None,
            "regions": region_total, "regionsReached": region_reached,
            "regionShare": round(region_reached / region_total, 4) if region_total else None,
            "functions": function_total, "functionsReached": function_reached,
            "functionShare": round(function_reached / function_total, 4) if function_total else None,
            "llvmCovTotals": {name: totals[name] for name in ("lines", "regions", "functions", "branches") if name in totals},
        }
    }


def unexecuted_lines(c, file, start, end):
    return sum(1 for number, count in c["lines"].get(file, {}).items() if start <= number <= end and count == 0)


def c_blind_map(c, repository):
    """Untouched C in the runtime, largest first: functions never called, and inside functions that
    ran, the outermost regions that never ran. Size is the instrumented lines that never ran."""
    entries = []
    for key, record in c["functions"].items():
        if record["count"] == 0:
            entries.append({"file": record["file"], "start": record["start"], "end": record["end"], "function": record["name"], "kind": "function never called", "lines": unexecuted_lines(c, record["file"], record["start"], record["end"]), "label": None})
    zero = defaultdict(list)
    for region_key, (count, function_key) in c["regions"].items():
        if count == 0 and c["functions"][function_key]["count"] > 0:
            zero[function_key].append(region_key)
    for function_key, keys in zero.items():
        def contains(outer, inner):
            return (outer != inner and outer[0] == inner[0]
                    and (outer[1], outer[2]) <= (inner[1], inner[2])
                    and (inner[3], inner[4]) <= (outer[3], outer[4]))
        for key in keys:
            if any(contains(other, key) for other in keys):
                continue
            file, l1, c1, l2, c2 = key
            text = ""
            source = source_lines(repository, file)
            if 0 < l1 <= len(source):
                text = source[l1 - 1].strip()[:110]
            size = unexecuted_lines(c, file, l1, l2)
            entries.append({"file": file, "start": l1, "end": l2, "function": c["functions"][function_key]["name"], "kind": "region never run", "lines": max(size, 1 if l1 == l2 else 0), "label": text})
    entries.sort(key=lambda entry: (-entry["lines"], entry["file"], entry["start"]))
    return entries


def compare_go(generator, fixtures):
    result = {}
    for label, predicate in go_groups:
        generator_only = fixtures_only = generator_only_statements = fixtures_only_statements = both = 0
        for key in set(generator) | set(fixtures):
            if not predicate(key[0]):
                continue
            reached_generator = generator.get(key, [0, 0])[1] > 0
            reached_fixtures = fixtures.get(key, [0, 0])[1] > 0
            statements = (generator.get(key) or fixtures.get(key))[0]
            if reached_generator and not reached_fixtures:
                generator_only += 1
                generator_only_statements += statements
            elif reached_fixtures and not reached_generator:
                fixtures_only += 1
                fixtures_only_statements += statements
            elif reached_fixtures and reached_generator:
                both += 1
        result[label] = {"blocksBoth": both, "blocksGeneratorOnly": generator_only, "statementsGeneratorOnly": generator_only_statements, "blocksFixturesOnly": fixtures_only, "statementsFixturesOnly": fixtures_only_statements}
    return result


def go_only_by_function(repository, reached, other):
    """Statements one side reaches and the other doesn't, by function, largest first."""
    totals = defaultdict(int)
    for key, (statements, count) in reached.items():
        if count > 0 and other.get(key, [0, 0])[1] == 0 and go_blind_scope(key[0]):
            name, start, end = enclosing_go_function(repository, key[0], key[1])
            totals[(key[0], start, end, name)] += statements
    ranked = sorted(totals.items(), key=lambda item: -item[1])
    return [{"file": file, "start": start, "end": end, "function": name, "statements": statements} for (file, start, end, name), statements in ranked]


def compare_c(generator, fixtures):
    def reached_lines(c):
        return {(file, number) for file, lines in c["lines"].items() for number, count in lines.items() if count > 0}

    def reached_functions(c):
        return {key for key, record in c["functions"].items() if record["count"] > 0}

    def reached_outcomes(c):
        outcomes = set()
        for key, (true, false) in c["branches"].items():
            if true > 0:
                outcomes.add(key + (True,))
            if false > 0:
                outcomes.add(key + (False,))
        return outcomes

    lines_generator, lines_fixtures = reached_lines(generator), reached_lines(fixtures)
    functions_generator, functions_fixtures = reached_functions(generator), reached_functions(fixtures)
    outcomes_generator, outcomes_fixtures = reached_outcomes(generator), reached_outcomes(fixtures)

    def by_function(only, c):
        totals = defaultdict(int)
        for file, number in only:
            for key, record in c["functions"].items():
                if record["file"] == file and record["start"] <= number <= record["end"]:
                    totals[key] += 1
                    break
        ranked = sorted(totals.items(), key=lambda item: -item[1])
        return [{"file": key[0], "start": c["functions"][key]["start"], "end": c["functions"][key]["end"], "function": c["functions"][key]["name"], "lines": count} for key, count in ranked]

    return {
        "summary": {
            "linesBoth": len(lines_generator & lines_fixtures),
            "linesGeneratorOnly": len(lines_generator - lines_fixtures),
            "linesFixturesOnly": len(lines_fixtures - lines_generator),
            "functionsBoth": len(functions_generator & functions_fixtures),
            "functionsGeneratorOnly": len(functions_generator - functions_fixtures),
            "functionsFixturesOnly": len(functions_fixtures - functions_generator),
            "branchOutcomesBoth": len(outcomes_generator & outcomes_fixtures),
            "branchOutcomesGeneratorOnly": len(outcomes_generator - outcomes_fixtures),
            "branchOutcomesFixturesOnly": len(outcomes_fixtures - outcomes_generator),
        },
        "generatorOnlyByFunction": by_function(lines_generator - lines_fixtures, generator),
        "fixturesOnlyByFunction": by_function(lines_fixtures - lines_generator, fixtures),
    }


def percent(value):
    return "n/a" if value is None else f"{100 * value:.1f}%"


def main():
    measurement, repository, report_path, json_path, commands_path = sys.argv[1:6]
    sides = {}
    for side in ("generator", "fixtures"):
        directory = os.path.join(measurement, side)
        if not os.path.exists(os.path.join(directory, "go.txt")):
            continue
        go = read_go_profile(os.path.join(directory, "go.txt"))
        c = read_c(directory, repository)
        meta = {}
        if os.path.exists(os.path.join(directory, "meta.json")):
            with open(os.path.join(directory, "meta.json")) as handle:
                meta = json.load(handle)
        sides[side] = {"go": go, "c": c, "meta": meta}

    report = {"sides": {}, "notes": []}
    for side, value in sides.items():
        report["sides"][side] = {
            "meta": value["meta"],
            "goShares": go_shares(value["go"]),
            "cShares": c_shares(value["c"]),
        }
    generator = sides["generator"]
    go_blind = go_blind_map(repository, generator["go"])
    c_blind = c_blind_map(generator["c"], repository)
    if "fixtures" in sides:
        fixtures = sides["fixtures"]
        report["comparison"] = {
            "go": compare_go(generator["go"], fixtures["go"]),
            "goGeneratorOnlyByFunction": go_only_by_function(repository, generator["go"], fixtures["go"])[:60],
            "goFixturesOnlyByFunction": go_only_by_function(repository, fixtures["go"], generator["go"])[:60],
            "c": compare_c(generator["c"], fixtures["c"]),
        }
        report["comparison"]["c"]["generatorOnlyByFunction"] = report["comparison"]["c"]["generatorOnlyByFunction"][:60]
        report["comparison"]["c"]["fixturesOnlyByFunction"] = report["comparison"]["c"]["fixturesOnlyByFunction"][:60]
        # How much of each blind region the fixtures reach: none means nothing in the test suite runs it.
        for region in go_blind:
            reached = sum(generator["go"][key][0] for key in region["keys"] if fixtures["go"].get(key, [0, 0])[1] > 0)
            region["fixturesReachStatements"] = reached
        for region in c_blind:
            lines = fixtures["c"]["lines"].get(region["file"], {})
            region["fixturesReachLines"] = sum(1 for number, count in lines.items() if region["start"] <= number <= region["end"] and count > 0 and generator["c"]["lines"].get(region["file"], {}).get(number, 0) == 0)
        report["blindMap"] = {"go": go_blind[:200], "c": c_blind[:200], "goRegionCount": len(go_blind), "cRegionCount": len(c_blind)}
        report["blindToBoth"] = {
            "goRegions": sum(1 for region in go_blind if region["fixturesReachStatements"] == 0),
            "goStatements": sum(region["statements"] for region in go_blind if region["fixturesReachStatements"] == 0),
            "cRegions": sum(1 for region in c_blind if region["fixturesReachLines"] == 0),
            "cLines": sum(region["lines"] for region in c_blind if region["fixturesReachLines"] == 0),
        }
    for region in go_blind:
        region.pop("keys", None)
        region["structural"] = structural(region)
    structural_regions = [region for region in go_blind if region["structural"]]
    go_blind = [region for region in go_blind if not region["structural"]]
    emit = [region for region in go_blind if re.fullmatch(r"internal/native/emit(_[a-z]+)?\.go", region["file"])]
    report["blindMap"] = {"go": go_blind[:200], "goEmit": emit[:100], "goStructural": structural_regions, "c": c_blind[:200], "goRegionCount": len(go_blind), "goEmitRegionCount": len(emit), "cRegionCount": len(c_blind)}

    with open(json_path, "w") as handle:
        json.dump(report, handle, indent=1)
        handle.write("\n")

    with open(commands_path) as handle:
        commands = handle.read()
    out = []
    write = out.append
    write("# What the generator never reaches\n")
    write("Coverage of the compiler (Go: internal/lower, internal/native, internal/ir, internal/flow) and the C runtime (internal/native/runtime) under the randomized test-program generator, against the hand-written oracle fixtures. Written by verify/coverage/measure.sh; REPORT.json beside it holds the same numbers and the longer lists.\n")
    for side, value in sides.items():
        meta = value["meta"]
        if meta:
            write(f"- **{side}**: " + ", ".join(f"{key} {meta[key]}" for key in meta) + "\n")
    write("\nGo's coverage counts blocks (each arm of an `if`, each `case`, is its own block) and the statements in them; Go has no branch coverage, so on the Go side the branch share is the block share. clang's source-based coverage counts branch outcomes (each condition's true and its false), regions and lines.\n")

    write("\n## (b) What the 2,000 seeds reach\n")
    write("| Area | Branches | Statements or lines |\n|---|---|---|\n")
    shares = report["sides"]["generator"]
    for label, share in shares["goShares"].items():
        write(f"| {label} | {percent(share['blockShare'])} of {share['blocks']} blocks | {percent(share['statementShare'])} of {share['statements']} statements |\n")
    runtime = shares["cShares"]["internal/native/runtime"]
    write(f"| internal/native/runtime (C) | {percent(runtime['branchShare'])} of {runtime['branchOutcomes']} branch outcomes | {percent(runtime['lineShare'])} of {runtime['lines']} lines; {percent(runtime['functionShare'])} of {runtime['functions']} functions; {percent(runtime['regionShare'])} of {runtime['regions']} regions |\n")

    if "fixtures" in sides:
        write("\n## (c) The oracle fixtures, for comparison\n")
        write("| Area | Generator branches | Fixtures branches | Generator statements or lines | Fixtures statements or lines |\n|---|---|---|---|---|\n")
        fixture_shares = report["sides"]["fixtures"]
        for label, share in shares["goShares"].items():
            other = fixture_shares["goShares"][label]
            write(f"| {label} | {percent(share['blockShare'])} | {percent(other['blockShare'])} | {percent(share['statementShare'])} | {percent(other['statementShare'])} |\n")
        other = fixture_shares["cShares"]["internal/native/runtime"]
        write(f"| internal/native/runtime (C) | {percent(runtime['branchShare'])} | {percent(other['branchShare'])} | {percent(runtime['lineShare'])} | {percent(other['lineShare'])} |\n")
        blind = report["blindToBoth"]
        write(f"\nOf the generator's blind regions, {blind['goRegions']} Go regions ({blind['goStatements']} statements) and {blind['cRegions']} C regions ({blind['cLines']} lines) are blind to the fixtures too.\n")
        write("\nWhat one side reaches and the other doesn't:\n\n")
        write("| Area | Both | Generator only | Fixtures only |\n|---|---|---|---|\n")
        for label, value in report["comparison"]["go"].items():
            write(f"| {label} | {value['blocksBoth']} blocks | {value['blocksGeneratorOnly']} blocks, {value['statementsGeneratorOnly']} statements | {value['blocksFixturesOnly']} blocks, {value['statementsFixturesOnly']} statements |\n")
        summary = report["comparison"]["c"]["summary"]
        write(f"| runtime lines (C) | {summary['linesBoth']} | {summary['linesGeneratorOnly']} | {summary['linesFixturesOnly']} |\n")
        write(f"| runtime functions (C) | {summary['functionsBoth']} | {summary['functionsGeneratorOnly']} | {summary['functionsFixturesOnly']} |\n")
        write(f"| runtime branch outcomes (C) | {summary['branchOutcomesBoth']} | {summary['branchOutcomesGeneratorOnly']} | {summary['branchOutcomesFixturesOnly']} |\n")
        for title, key, unit in (("Go the generator reaches and the fixtures don't", "goGeneratorOnlyByFunction", "statements"), ("Go the fixtures reach and the generator doesn't", "goFixturesOnlyByFunction", "statements")):
            write(f"\n### {title} (top 20 functions, internal/lower and internal/native)\n\n| Statements | Function | Where |\n|---|---|---|\n")
            for entry in report["comparison"][key][:20]:
                write(f"| {entry[unit]} | `{entry['function']}` | {entry['file']}:{entry['start']}-{entry['end']} |\n")
        for title, key in (("C the generator reaches and the fixtures don't", "generatorOnlyByFunction"), ("C the fixtures reach and the generator doesn't", "fixturesOnlyByFunction")):
            write(f"\n### {title} (top 20 functions)\n\n| Lines | Function | Where |\n|---|---|---|\n")
            for entry in report["comparison"]["c"][key][:20]:
                write(f"| {entry['lines']} | `{entry['function']}` | {entry['file']}:{entry['start']}-{entry['end']} |\n")

    write("\n## (a) The blind map: what no generated program executes\n")
    write(f"\n### Go, internal/lower and internal/native (top 40 of {len(go_blind)} untouched regions, by statements)\n\n")
    write("A function no program enters is one region; inside a function that ran, a region is a run of blocks that never ran, cut at each `case` so every switch arm stands alone. The First line column names the case a region opens, or the case it sits inside.\n\n")
    structural_statements = sum(region["statements"] for region in report["blindMap"]["goStructural"])
    write(f"Left out of the ranking, as unreachable by construction ({len(report['blindMap']['goStructural'])} regions, {structural_statements} statements): the fuzzer compiles through `adamic c` and `adamic js` and runs clang itself, so native.Build and Flags, the split build (units.go, units_tsgo.go), the runtime cache (library.go) and tsgo (tsgo.go in native and lower, which `adamic c` refuses) never run in the measured binary. REPORT.json lists them under blindMap.goStructural.\n\n")
    write("The last column is how many of the region's statements the oracle fixtures run; 0 means nothing in the fixtures reaches it either.\n\n")
    write("| # | Statements | Kind | Function | Where | First line | Fixtures run |\n|---|---|---|---|---|---|---|\n")
    for index, region in enumerate(go_blind[:40], 1):
        label = (region["label"] or "").replace("|", "\\|")
        write(f"| {index} | {region['statements']} | {region['kind']} | `{region['function']}` | {region['file']}:{region['start']}-{region['end']} | {('`' + label + '`') if label else ''} | {region.get('fixturesReachStatements', '')} |\n")
    emit = [region for region in go_blind if re.fullmatch(r"internal/native/emit(_[a-z]+)?\.go", region["file"])]
    write(f"\n### Go, internal/native/emit*.go alone (top 40 of {len(emit)} untouched regions, by statements)\n\n")
    write("The emitter's regions are small, so few reach the table above; here they are on their own.\n\n")
    write("| # | Statements | Kind | Function | Where | First line | Fixtures run |\n|---|---|---|---|---|---|---|\n")
    for index, region in enumerate(emit[:40], 1):
        label = (region["label"] or "").replace("|", "\\|")
        write(f"| {index} | {region['statements']} | {region['kind']} | `{region['function']}` | {region['file']}:{region['start']}-{region['end']} | {('`' + label + '`') if label else ''} | {region.get('fixturesReachStatements', '')} |\n")
    write(f"\n### C, internal/native/runtime (top 40 of {len(c_blind)} untouched regions, by lines that never ran)\n\n")
    write("A function no program calls is one region; inside a function that ran, a region is an outermost code region (a branch's arm, a loop body, the rest of a function after a return) that never ran.\n\n")
    write("The last column is how many of the region's lines the oracle fixtures run.\n\n")
    write("| # | Lines | Kind | Function | Where | First line | Fixtures run |\n|---|---|---|---|---|---|---|\n")
    for index, region in enumerate(c_blind[:40], 1):
        label = (region["label"] or "").replace("|", "\\|")
        write(f"| {index} | {region['lines']} | {region['kind']} | `{region['function']}` | {region['file']}:{region['start']}-{region['end']} | {('`' + label + '`') if label else ''} | {region.get('fixturesReachLines', '')} |\n")
    write("\n## How it was measured\n\n")
    write(commands)
    with open(report_path, "w") as handle:
        handle.write("".join(out))


if __name__ == "__main__":
    main()
