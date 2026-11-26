"""Summarize the door's recorded stops from its JSON report."""

import argparse
from collections import Counter
from hashlib import sha256
import html
import json
from pathlib import Path
import re
import sys

sys.dont_write_bytecode = True
sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "census" / "firststops"))
from families import familyOf

# Keep the same position stripping as census/firststops/fsdiff.py.
POSITION = re.compile(r"(\.[cm]?ts|\.a)(?::\d+(?::\d+)?)")
# The five body phases in ir.go:157 to :161 at 51d5dada.
BODY_PHASES = {"Refusal", "Signature", "Statics", "Body", "Statement"}

# ir.go:157 to :161 at 51d5dada names whole-program steps as a class.
# The names come from each first argument of a firstLinkPass call at 51d5dada
# (36 calls), plus the two firstLinkProgramStop calls.
WHOLE_PROGRAM_PHASES = {
    "NormalizeAsync",
    "ProgramRegion",
    "RelationOverflow",
    "asyncPayloadFailures",
    "certifySourceSlots",
    "checkAccessorSpreads",
    "checkArrayArmViews",
    "checkBoundArguments",
    "checkCheckedPredicateParameters",
    "checkClosureSpreadLayouts",
    "checkDynamicKeyedGetters",
    "checkEnumerationReachViews",
    "checkFlowLiveness",
    "checkFunctionConstructorRoads",
    "checkGeneratorRoads",
    "checkGenericSpreadCopyReads",
    "checkHostMemberWrites",
    "checkInspectorDoomedContracts",
    "checkIterableViews",
    "checkLibrarySpreads",
    "checkOptionalMemberPresence",
    "checkOverloadVerdictCalls",
    "checkPromiseConsumption",
    "checkRangeUnionReads",
    "checkRecordGuardedReads",
    "checkRegExpServedValues",
    "checkRelationViewsTaken",
    "checkSlotWrites",
    "checkSourceSlotPending",
    "checkViewCopies",
    "checkViewLogicalWrites",
    "checkViewWidenedReads",
    "exceptions",
    "findCycles",
    "finishAccessors",
    "layFunctionConstructorWindows",
    "unlowerable",
    "voidDiscards",
}
NO_PASS_STOPS = "this head records no whole-program pass stops"
TEXT_FIELDS = ("Kind", "Reason", "Fix", "Site", "Unit", "Phase")
JUDGMENT_FIELDS = ("Pass", "Site", "Why")
# A door whose judging pass gave up partway (#nxb3xh1, Lower's door-2) writes its report alone and exits 1 before emit.
REPORT_ONLY = {"traps.json", "traps.md"}


def validate(source):
    if not isinstance(source, dict) or not isinstance(source.get("findings"), str):
        raise ValueError("expected a report object with a findings string")
    rows = source.get("traps")
    if not isinstance(rows, list):
        raise ValueError("expected a list in traps")
    skipped = source.get("rewrites_skipped")
    if not isinstance(skipped, list) or not all(isinstance(name, str) for name in skipped):
        raise ValueError("expected a list of strings in rewrites_skipped")
    abandoned = source.get("judgments_abandoned", [])
    if not isinstance(abandoned, list) or not all(
            isinstance(judgment, dict) and all(isinstance(judgment.get(key), str) for key in JUDGMENT_FIELDS)
            for judgment in abandoned):
        raise ValueError("expected a list of passes with a Site and a Why in judgments_abandoned")
    for index, row in enumerate(rows):
        if (not isinstance(row, dict)
                or any(not isinstance(row.get(key), str) for key in TEXT_FIELDS)
                or not row["Kind"]
                or type(row.get("Suspect")) is not bool
                or type(row.get("Echo")) is not bool):
            raise ValueError(f"invalid recorded stop at index {index}")
        phase = row["Phase"]
        if phase not in BODY_PHASES and phase not in WHOLE_PROGRAM_PHASES:
            raise ValueError(f"unknown phase {json.dumps(phase)} in record {index}")

    counts = {
        "total": len(rows),
        "by_phase": dict(Counter(row["Phase"] for row in rows)),
        "by_kind": dict(Counter(row["Kind"] for row in rows)),
        "by_reason": dict(Counter(f'{row["Kind"]}: {row["Reason"]}' for row in rows)),
        "unplaced": sum(not row["Unit"] for row in rows),
        "suspect": sum(row["Suspect"] for row in rows),
        "echo": sum(row["Echo"] for row in rows),
    }
    for key, expected in counts.items():
        actual = source.get(key)
        if isinstance(expected, dict):
            valid = isinstance(actual, dict) and all(type(count) is int for count in actual.values())
        else:
            valid = type(actual) is int
        if not valid or actual != expected:
            raise ValueError(f"{key} disagrees with the recorded stops")
    return rows


def strip_tree_root(data, tree_root):
    """The door's report without the tree it read, so two runs in different scratch trees hash alike.

    The door makes sites relative, but a reason can name a file by its absolute path. The root is removed from the
    bytes, so the hash is still over the door's own text, less the root, and the count says how many places held it.
    """
    root = str(tree_root).rstrip("/") + "/"
    # JSON writes these characters escaped, so the bytes could not be found as written.
    if (root == "/" or not Path(root).is_absolute() or not root.isascii()
            or any(character in root for character in '"\\<>&')):
        raise ValueError(f"tree root {json.dumps(str(tree_root))} must be an absolute path JSON writes as it is")
    prefix = root.encode()
    return data.replace(prefix, b""), data.count(prefix)


def summarize(data, tree_root=None):
    stripped = 0
    if tree_root is not None:
        data, stripped = strip_tree_root(data, tree_root)
    source = json.loads(data)
    rows = validate(source)
    placed = {}
    for row in rows:
        reason = POSITION.sub(r"\1", row["Reason"])
        placed.setdefault(row["Phase"], []).append({
            "record": row,
            "normalized_reason": reason,
            "family": familyOf(reason),
        })

    phases = {}
    for phase, sites in placed.items():
        by_reason = Counter(
            (site["record"]["Kind"], site["normalized_reason"], site["family"]) for site in sites
        )
        flags = Counter()
        for site in sites:
            row = site["record"]
            echo, suspect = row["Echo"], row["Suspect"]
            flags["echo"] += echo
            flags["suspect"] += suspect
            flags["without_flags"] += not echo and not suspect
        phases[phase] = {
            "total": len(sites),
            "by_kind": dict(Counter(site["record"]["Kind"] for site in sites)),
            "by_reason": [
                {"kind": kind, "reason": reason, "family": family, "count": count}
                for (kind, reason, family), count in sorted(by_reason.items())
            ],
            "echo": flags["echo"],
            "suspect": flags["suspect"],
            "without_flags": flags["without_flags"],
            "sites": sites,
        }

    derived = {
        "sha256": sha256(data).hexdigest(),
        "tree_root_stripped": stripped,
        "phases": phases,
        "flag_counts": {
            key: sum(phase[key] for phase in phases.values())
            for key in ("echo", "suspect", "without_flags")
        },
        "has_whole_program_pass_stops": bool(WHOLE_PROGRAM_PHASES.intersection(phases)),
        "absolute_sites": sum(Path(row["Site"]).is_absolute() for row in rows),
        "empty_sites": sum(row["Site"] == "" for row in rows),
        "unknown_top_level_keys": sorted(source.keys() - {
            "findings", "traps", "total", "by_phase", "by_kind", "by_reason",
            "unplaced", "suspect", "echo", "rewrites_skipped", "judgments_abandoned",
        }),
    }
    for key in source.keys() & derived.keys():
        if source[key] != derived[key]:
            raise ValueError(f"door field {key} conflicts with the summary")
    return {**source, **derived}


def cell(value):
    text = "" if value is None else str(value)
    text = html.escape(text.replace("\r", r"\r").replace("\n", r"\n"), quote=False)
    return re.sub(r"([\\|`*_])", r"\\\1", text)


def table(headers, rows):
    yield "| " + " | ".join(headers) + " |"
    yield "| " + " | ".join("---" for _ in headers) + " |"
    for row in rows:
        yield "| " + " | ".join(map(cell, row)) + " |"


def markdown(summary):
    flags = summary["flag_counts"]
    lines = [
        "# First link over tsc",
        "",
        "Recorded stops are measurement leads. Reproduce in ordinary mode before filing.",
        "",
    ]
    if not summary["has_whole_program_pass_stops"]:
        lines.extend([NO_PASS_STOPS, ""])
    stripped = summary["tree_root_stripped"]
    lines.extend([
        f'Door report SHA256 `{summary["sha256"]}`'
        + (f', the tree root stripped from {stripped} places.' if stripped else '.'),
        "",
        f'{summary["total"]} recorded stops, {summary["unplaced"]} unplaced, '
        f'{flags["echo"]} echo, {flags["suspect"]} suspect, {flags["without_flags"]} without either flag.',
        "",
        "Echo and suspect counts can overlap. A Gap or an unplaced recorded stop marks a measurement gap.",
    ])
    if summary["rewrites_skipped"]:
        skipped = ", ".join(map(cell, summary["rewrites_skipped"]))
        lines.extend(["", f"Rewrites skipped past recorded stops, in door order: {skipped}."])
    if summary.get("judgments_abandoned"):
        given = "; ".join(
            f'{cell(judgment["Pass"])} at {cell(judgment["Site"])} ({cell(judgment["Why"])})'
            for judgment in summary["judgments_abandoned"]
        )
        lines.extend(["", f"Judgments abandoned past a stop, in door order: {given}. "
                      "Every unit past each went unjudged, so the door emitted nothing."])
    for name, phase in summary["phases"].items():
        lines.extend([
            "",
            f'## {cell(name)}',
            "",
            f'{phase["total"]} recorded stops, {phase["echo"]} echo, '
            f'{phase["suspect"]} suspect, {phase["without_flags"]} without either flag.',
            "",
        ])
        lines.extend(table(("Kind", "Recorded stops"), sorted(phase["by_kind"].items())))
        lines.append("")
        lines.extend(table(("Kind", "Normalized reason", "Family", "Recorded stops"), (
            (group["kind"], group["reason"], group["family"], group["count"])
            for group in phase["by_reason"]
        )))
        lines.append("")
        rows = []
        for site in phase["sites"]:
            row = site["record"]
            labels = [label for label in ("Echo", "Suspect") if row[label]]
            rows.append((row["Kind"], row["Reason"], site["family"], row["Site"],
                         row["Unit"], ", ".join(labels)))
        lines.extend(table(("Kind", "Raw reason", "Family", "Site", "Unit", "Flags"), rows))
    return "\n".join(lines) + "\n"


def files_under(output):
    return {path.relative_to(output).as_posix() for path in output.rglob("*") if not path.is_dir()}


def reading(door_exit, written, summary):
    """Whether the door's exit leaves a completed report: 0, or 1 before emit with only its report written, naming
    the judging pass that gave up (#nxb3xh1). That is a reading of the program, never the harness's failure."""
    if door_exit == 0:
        return True
    return door_exit == 1 and written == REPORT_ONLY and bool(summary.get("judgments_abandoned"))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("output", type=Path, help="directory containing the door report")
    parser.add_argument("--tree-root", type=Path, help="the tree the door read, stripped from the report")
    parser.add_argument("--door-exit", type=int, default=0,
                        help="the door's exit; one that leaves no completed report exits 3 once the row is written")
    args = parser.parse_args()
    try:
        written = files_under(args.output)
        summary = summarize((args.output / "traps.json").read_bytes(), args.tree_root)
        (args.output / "firstlink.json").write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
        (args.output / "firstlink.md").write_text(markdown(summary), encoding="utf-8")
    except (OSError, ValueError) as error:
        parser.exit(1, f"firstlink report: {error}\n")
    print(f'REPORT-SHA256 {summary["sha256"]}')
    if args.door_exit == 0:
        return
    if reading(args.door_exit, written, summary):
        print("DOOR exit 1 before emit, its report alone: a judging pass gave up, so this is a reading, not a failure")
        return
    print(f"DOOR exit {args.door_exit} after its report: row 1 is written, and the run fails")
    sys.exit(3)


if __name__ == "__main__":
    main()
