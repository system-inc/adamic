#!/usr/bin/env python3
"""Names the area a worker's branch merges into, or holds it for a person to decide.

usage: cloud/integration/area-route.py <branch>...
Prints one line per branch, "<branch>\t<area>\t<how>" or "<branch>\thold\t<why, and who to ask>", and
exits 4 when any branch is held. The automatic area merges call route() (or this) for every green
worker branch: a held branch is reported to integration and its Circle by name, never merged into a
guessed area and never marked done.

The order: census-owners.tsv's override when its owner holds exactly one area; then the longest
prefix in area-routes.tsv; then the fleet of the Codex sessions that name the branch (fleet-areas.tsv);
otherwise held. Never changed-file heuristics: a person decides each family once, in area-routes.tsv.
"""
import json
import os
import sqlite3
import sys
from pathlib import Path

directory = Path(__file__).resolve().parent
ahra = Path(os.environ.get("AHRA_DIRECTORY", "/Users/kirkouimet/Projects/ahra"))
databasePath = Path(os.environ.get("ADAMIC_AI_DATABASE", str(ahra / "modules/ai/data/ai.db")))
rosterPath = Path(os.environ.get("ADAMIC_FLEET_ROSTER", str(ahra / "modules/ai/data/fleet-roster.json")))

# census-owners.tsv still names the two cohere sub-Circles that held the stage 1 areas before Oct 7.
ownerAliases = {"system_cohere_lint": "stage1-lint", "system_cohere_format": "stage1-format"}


def rows(name):
    return [line.split("\t") for line in (directory / name).read_text().splitlines() if line.strip() and not line.startswith("#")]


def areaNames():
    return {row[0] for row in rows("areas.tsv")}


def ownerArea(owner):
    """The one area this owner holds, or None when it holds none or several."""
    held = {row[0] for row in rows("areas.tsv") if row[1] == owner}
    if owner in ownerAliases:
        held = {ownerAliases[owner]}
    return held.pop() if len(held) == 1 else None


def prefixRoute(branch):
    """(area or "hold", note, prefix) for the longest matching prefix in area-routes.tsv, or None."""
    matches = [row for row in rows("area-routes.tsv") if branch.startswith(row[0])]
    if not matches:
        return None
    prefix, area, *note = max(matches, key=lambda row: len(row[0]))
    return area, (note[0] if note else ""), prefix


def fleetArea(fleet):
    for pattern, area in rows("fleet-areas.tsv"):
        if fleet == pattern or (pattern.endswith("*") and fleet.startswith(pattern[:-1])):
            return area
    return None


def sessionTexts():
    """(session id, text) for every Codex reply and title ai.db holds, read once per process."""
    if not hasattr(sessionTexts, "cache"):
        texts = []
        if databasePath.exists():
            with sqlite3.connect(databasePath.resolve().as_uri() + "?mode=ro", uri=True) as connection:
                texts += connection.execute("select session_id, text from replies").fetchall()
                texts += connection.execute("select session_id, coalesce(title, '') from sessions").fetchall()
        sessionTexts.cache = texts
    return sessionTexts.cache


def roster():
    if not hasattr(roster, "cache"):
        roster.cache = {entry["id"]: entry.get("fleet") for entry in json.loads(rosterPath.read_text())} if rosterPath.exists() else {}
    return roster.cache


def fleetsNaming(branch):
    return {roster().get(session) for session, text in sessionTexts() if branch in text} - {None}


def route(branch):
    """(area, how) when the branch routes, or ("hold", why)."""
    valid = areaNames()
    overrides = {row[0]: row[1] for row in rows("census-owners.tsv")}
    if branch in overrides:
        area = ownerArea(overrides[branch])
        if area in valid:
            return area, f"census-owners.tsv names {overrides[branch]}"
    prefixed = prefixRoute(branch)
    if prefixed:
        area, note, prefix = prefixed
        if area == "hold":
            return "hold", f"area-routes.tsv holds {prefix}*: ask {note}" if note else f"area-routes.tsv holds {prefix}*"
        if area not in valid:
            return "hold", f"area-routes.tsv sends {prefix}* to {area}, which areas.tsv doesn't list"
        return area, f"area-routes.tsv {prefix}"
    fleets = fleetsNaming(branch)
    areas = {fleetArea(fleet) for fleet in fleets}
    if len(areas) == 1 and None not in areas and areas <= valid:
        return areas.pop(), f"fleet {', '.join(sorted(fleets))}"
    if fleets:
        return "hold", f"its sessions' fleets ({', '.join(sorted(fleets))}) don't name one area in fleet-areas.tsv"
    return "hold", "no prefix in area-routes.tsv and no fleet session names it: integration decides, then adds its family"


def main():
    if len(sys.argv) < 2:
        print(__doc__.strip().splitlines()[2], file=sys.stderr)
        return 2
    held = False
    for branch in sys.argv[1:]:
        area, how = route(branch)
        held = held or area == "hold"
        print(f"{branch}\t{area}\t{how}")
    return 4 if held else 0


if __name__ == "__main__":
    sys.exit(main())
