import collections
import json
import pathlib
import sys

proof, before_root, after_root = map(pathlib.Path, sys.argv[1:4])
mode = sys.argv[4] if len(sys.argv) == 5 else "after"

def load(phase, root):
    text = (proof / f"census-{phase}.jsonl").read_text().replace(str(root) + "/", "")
    header, record, trailer = [json.loads(line) for line in text.splitlines()]
    assert header["status"] == "measurement" and header["latent_mode"] == "full"
    assert header["module_format"] == "CommonJSBundle", header["module_format"]
    assert not header["checker_rejected"] and not header["diagnostics"], header["diagnostics"]
    assert record["file"] == "src/compiler/checker.ts", record["file"]
    assert trailer["status"] == "relation_report"
    return record, trailer

def count(findings):
    return collections.Counter(json.dumps(d, sort_keys=True) for d in findings or [])

prefix = "scan-" if mode == "scan" else ""
before_record, before_trailer = load(prefix + "before", before_root)
after_record, after_trailer = load(prefix + ("after" if mode == "scan" else mode), after_root)
before, after = count(before_record["findings"]), count(after_record["findings"])
if mode == "mutant":
    assert before == after, "mutant findings differ from baseline"
    assert before_record["units"] == after_record["units"], "mutant units differ from baseline"
    assert before_trailer == after_trailer, "mutant relation report differs from baseline"
    print("Restored complete census, including both target groups")
    sys.exit(0)

target_contexts = json.loads((proof / "target-contexts.json").read_text())
reasons = {site["reason"] for site in target_contexts}
def target(d):
    return d["reason"] in reasons and d["kind"] == "Refused"

def expand(counts):
    return [json.loads(event) for event in counts.elements()]

removed, added = expand(before - after), expand(after - before)
groups_before = [d for d in expand(before) if target(d)]
groups_after = [d for d in expand(after) if target(d)]
unique_before = {(d["where"], d["reason"]) for d in groups_before if d["phase"] == "refusal_scan"}
unique_after = {(d["where"], d["reason"]) for d in groups_after}
target_map = json.loads((proof / "target-map.json").read_text())
reason_by_group = {d["group"]: d["reason"] for d in target_contexts}
expected = {(d["current"], reason_by_group[d["group"]]) for d in target_map}
blocked_units = {d["unit"] for d in groups_before if d["phase"] != "refusal_scan"}
scan_blockers = {
    (d["unit"], d["where"], d["reason"].split(" in ", 1)[-1])
    for d in groups_before if d["phase"] == "refusal_scan"
}
def newly_reached(d):
    if d["phase"] != "refusal_scan":
        return d["unit"] in blocked_units
    return d["reason"].startswith("optional property ") and (
        d["unit"], d["where"], d["reason"].split(" in ", 1)[-1]
    ) in scan_blockers

reached = [d for d in added if newly_reached(d)]
new = [d for d in added if not newly_reached(d)]
out_of_scope_removed = [
    d for d in removed
    if not target(d) and (d["phase"] == "refusal_scan" or d["unit"] not in blocked_units)
]
result = {
    "before_count": before.total(), "after_count": after.total(),
    "target_unique_before": sorted(unique_before), "target_unique_after": sorted(unique_after),
    "target_findings_before": groups_before, "target_findings_after": groups_after,
    "blocked_units": sorted(blocked_units),
    "removed": removed, "reached": reached, "new": new,
    "out_of_scope_removed": out_of_scope_removed,
    "missing_at_baseline": sorted(expected - unique_before),
    "extra_at_baseline": sorted(unique_before - expected),
    "trailer_findings_removed": expand(count(before_trailer["findings"]) - count(after_trailer["findings"])),
    "trailer_findings_added": expand(count(after_trailer["findings"]) - count(before_trailer["findings"])),
}
(proof / f"{prefix}census-delta.json").write_text(json.dumps(result, indent=2) + "\n")
print(f"All findings {before.total()} -> {after.total()}, target sites {len(unique_before)} -> {len(unique_after)}")
for label, events in [("REACHED", reached), ("NEW OR CHANGED", new), ("OUTSIDE REMOVED", out_of_scope_removed), ("TRAILER ADDED", result["trailer_findings_added"])]:
    for event in events:
        print(f'{label} in {event["unit"]} ({event["phase"]}, {event["kind"]}): {event["text"]}')
if expected != unique_before:
    print("Missing baseline sites:", result["missing_at_baseline"])
    print("Extra baseline sites:", result["extra_at_baseline"])
if expected != unique_before or unique_after or new or out_of_scope_removed or result["trailer_findings_added"]:
    sys.exit(1)
