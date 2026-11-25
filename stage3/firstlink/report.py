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
# The five body phases in ir.go:153 to :157 at ea68172c.
BODY_PHASES = {"Refusal", "Signature", "Statics", "Body", "Statement"}

# ir.go:153 to :157 at ea68172c names whole-program steps as a class.
# The names come from each first argument of a firstLinkPass call at ea68172c
# (33 calls), plus the two firstLinkProgramStop calls.
WHOLE_PROGRAM_PHASES = {
    "NormalizeAsync",
    "ProgramRegion",
    "RelationOverflow",
    "asyncPayloadFailures",
    "certifySourceSlots",
    "checkAccessorSpreads",
    "checkBoundArguments",
    "checkCheckedPredicateParameters",
    "checkDynamicKeyedGetters",
    "checkEnumerationReachViews",
    "checkFlowLiveness",
    "checkFunctionConstructorRoads",
    "checkGeneratorRoads",
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


def validate(source):
    if not isinstance(source, dict) or not isinstance(source.get("findings"), str):
        raise ValueError("expected a report object with a findings string")
    rows = source.get("traps")
    if not isinstance(rows, list):
        raise ValueError("expected a list in traps")
    skipped = source.get("rewrites_skipped")
    if not isinstance(skipped, list) or not all(isinstance(name, str) for name in skipped):
        raise ValueError("expected a list of strings in rewrites_skipped")
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


def summarize(data):
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
            "unplaced", "suspect", "echo", "rewrites_skipped",
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
    lines.extend([
        f'Door report SHA256 `{summary["sha256"]}`.',
        "",
        f'{summary["total"]} recorded stops, {summary["unplaced"]} unplaced, '
        f'{flags["echo"]} echo, {flags["suspect"]} suspect, {flags["without_flags"]} without either flag.',
        "",
        "Echo and suspect counts can overlap. A Gap or an unplaced recorded stop marks a measurement gap.",
    ])
    if summary["rewrites_skipped"]:
        skipped = ", ".join(map(cell, summary["rewrites_skipped"]))
        lines.extend(["", f"Rewrites skipped past recorded stops, in door order: {skipped}."])
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


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("output", type=Path, help="directory containing the door report")
    args = parser.parse_args()
    try:
        summary = summarize((args.output / "traps.json").read_bytes())
        (args.output / "firstlink.json").write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
        (args.output / "firstlink.md").write_text(markdown(summary), encoding="utf-8")
    except (OSError, ValueError) as error:
        parser.exit(1, f"firstlink report: {error}\n")
    print(f'REPORT-SHA256 {summary["sha256"]}')


if __name__ == "__main__":
    main()
