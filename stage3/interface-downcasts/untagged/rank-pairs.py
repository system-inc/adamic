#!/usr/bin/env python3
"""Freeze untagged demand without claiming component coverage as admission."""
import gzip
import hashlib
import json
from pathlib import Path

root = Path(__file__).resolve().parent
source = root.parent / "lane4/read-demand-pairs.json.gz"
pairs = [pair for pair in json.loads(gzip.decompress(source.read_bytes()))
         if "untagged object union" in pair["families"]]
pairs.sort(key=lambda pair: (-pair["reads"], pair["receiver_type_id"], pair["field"]))
assert len(pairs) == 84 and sum(pair["reads"] for pair in pairs) == 186
for rank, pair in enumerate(pairs, 1):
    pair["rank"] = rank
    pair["status"] = "pending source integration"
result = {"source_sha256": hashlib.sha256(source.read_bytes()).hexdigest(),
          "pairs": 84, "reads": 186, "remaining_pairs": 84,
          "remaining_reads": 186, "ranked_pairs": pairs}
(root / "pair-progress.json").write_text(json.dumps(result, indent=2) + "\n")
print("84 pairs, 186 reads; remaining 84 pairs, 186 reads")
