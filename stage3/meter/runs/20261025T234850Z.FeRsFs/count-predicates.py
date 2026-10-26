"""Reproduce task #tzd3gjg counts from the committed compressed streams."""
import collections
import gzip
import json
from pathlib import Path

PREFIX = "checked predicate membership by an enum tag another member can hold"
run = Path(__file__).resolve().parent
result = {}
for stream in ("main/latent.jsonl.gz", "main/tsc/latent.jsonl.gz"):
    enum_sites = collections.defaultdict(set)
    checked_sites = set()
    checked_findings = set()
    for row in map(json.loads, gzip.open(run / stream, "rt")):
        for finding in row.get("findings", []):
            reason = finding.get("reason", "")
            where = finding.get("where", "")
            if "/src/" in where:
                where = where[where.index("/src/") + 1:]
            if reason.startswith(PREFIX):
                enum_sites[where.rsplit(":", 2)[0]].add(where)
            if "checked predicate" in reason:
                checked_sites.add(where)
                checked_findings.add(tuple(finding.get(k, "") for k in
                                           ("kind", "where", "reason", "text")))
    result[stream] = {
        "enum_tag_sites_by_file": {k: len(v) for k, v in sorted(enum_sites.items())},
        "enum_tag_sites": sum(map(len, enum_sites.values())),
        "checked_predicate_sites": len(checked_sites),
        "checked_predicate_unique_findings": len(checked_findings),
    }
print(json.dumps(result, indent=2))
