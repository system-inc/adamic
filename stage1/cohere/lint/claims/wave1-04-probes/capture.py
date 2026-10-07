"""Run original Go tests, preserve asserted cases, and measure JSX dependencies.

This does not claim independent Adamic rule parity. Test output goes to files.
"""
import collections
import json
import os
from pathlib import Path
import subprocess
import tempfile

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[4]
COHERE = REPO / "cohere"
EVIDENCE = HERE / "evidence"
EVIDENCE.mkdir(exist_ok=True)
PATTERNS = {
    "core": "^TestRequireDescription",
    "next": "^TestGoogleFontDisplay",
    "tailwind": "^TestNoPhysicalDirection",
}
with tempfile.TemporaryDirectory(prefix="lint-wave1-04-capture-") as temporary:
    scratch = Path(temporary)
    capture = scratch / "capture"
    capture.mkdir()
    env = os.environ | {"COHERE_DOCS_CAPTURE": str(capture)}
    for family, pattern in PATTERNS.items():
        with (EVIDENCE / ("go-" + family + ".log")).open("w") as log:
            subprocess.run(
                ["go", "test", "-count=1", "-v", "-timeout=10m",
                 "./internal/lint/rules/" + family, "-run", pattern],
                cwd=COHERE, env=env, stdout=log, stderr=subprocess.STDOUT,
                check=True,
            )
    unique = {}
    for path in sorted(capture.glob("*.jsonl")):
        for line in path.read_text().split("\n"):
            if not line:
                continue
            row = json.loads(line)
            key = (row["rule"], row["file"], row["source"],
                   json.dumps(row.get("options"), sort_keys=True))
            if key in unique:
                assert unique[key].get("findings", []) == row.get("findings", [])
            unique[key] = row
    rows = [unique[key] for key in sorted(unique)]
    path = EVIDENCE / "asserted-cases.json"
    path.write_text(json.dumps(rows, ensure_ascii=True, indent=2) + "\n")
    virtual = COHERE / "adamic_wave104_gap_probe.go"
    overlay = scratch / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {str(virtual): str(HERE / "oracle.go.txt")}}))
    oracle = scratch / "oracle"
    with (EVIDENCE / "oracle-build.log").open("w") as log:
        subprocess.run(["go", "build", "-overlay=" + str(overlay), "-o", str(oracle), str(virtual)],
                       cwd=COHERE, stdout=log, stderr=subprocess.STDOUT, check=True)
    with (EVIDENCE / "go-facts.jsonl").open("w") as output, (EVIDENCE / "oracle-stderr.log").open("w") as error:
        subprocess.run([str(oracle), str(path), "--facts"], stdout=output, stderr=error, check=True)
    assert (EVIDENCE / "oracle-stderr.log").stat().st_size == 0
    facts = [json.loads(line) for line in (EVIDENCE / "go-facts.jsonl").read_text().split("\n") if line]
    assert len(facts) == len(rows)
    summary = collections.defaultdict(lambda: {"sources": 0, "jsx_sources": 0, "recovery_sources": 0, "findings": 0})
    for row, fact in zip(rows, facts):
        assert row["rule"] == fact["rule"] and row["source"] == fact["source"]
        expected = [finding["messageId"] for finding in (row.get("findings") or [])]
        assert fact["findingIDs"] == expected, (row, fact)
        item = summary[row["rule"]]
        item["sources"] += 1
        item["jsx_sources"] += fact["jsx"] > 0
        item["recovery_sources"] += fact["parseDiagnostics"] > 0
        item["findings"] += len(expected)
    result = dict(summary)
    (EVIDENCE / "corpus-size.json").write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps(result, indent=2))
