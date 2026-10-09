#!/usr/bin/env python3
"""Check the frozen baseline against the assigned intervals and table."""
import copy
import json
from pathlib import Path

root = Path(__file__).resolve().parent
expected = [
    ("hidden-01-large", "transformers/declarations.ts", 68180, 81805),
    ("hidden-01-small", "transformers/declarations.ts", 82928, 89827),
    ("hidden-05-large", "transformers/es2018.ts", 35690, 41768),
    ("hidden-05-small", "transformers/es2015.ts", 144178, 149926),
    ("hidden-06", "transformers/esDecorators.ts", 61324, 72741),
    ("hidden-13", "transformers/es2017.ts", 30237, 37526),
    ("hidden-14", "utilities.ts", 455532, 462634),
]

def validate(data):
    assert data["compiler"] == "c68bf26cb4225e1f24a84f1a412f10e7eaa3da9a"
    assert data["all_adapted_hashes_verified"] == 82
    assert len(data["regions"]) == len(expected)
    for row, (label, name, start, end) in zip(data["regions"], expected):
        assert (row["label"], row["file"], row["start"], row["end"]) == (label, name, start, end)
        assert row["bytes"] == end - start
        for side in ("before", "after"):
            assert row[side]["ranges"] == [[start, end]]
            assert row[side]["hidden_bytes"] == end - start
        assert row["revealed_bytes"] == 0

baseline = json.loads((root / "regions.json").read_text())
validate(baseline)
print("baseline: seven intervals pass")
for name, field in (("wrong byte count", "bytes"), ("shifted endpoint", "start"), ("false reveal", "revealed_bytes")):
    mutant = copy.deepcopy(baseline)
    mutant["regions"][0][field] += 1
    try:
        validate(mutant)
    except AssertionError:
        print(name + ": caught")
    else:
        raise AssertionError(name + " survived")
replays = json.loads((root / "replays.json").read_text())
assert {row["label"]: row["exit"] for row in replays} == {"01": 1, "05": 0, "05-optional": 1, "06": 0, "13": 0, "14": 0}
assert "MUTANT" in (root / "replay-mutant.log.txt").read_text()
assert "replay signature did not reproduce" in (root / "replay-mutant.log.txt").read_text()
print("replay outcomes and wrong-reason rejection: pass")
