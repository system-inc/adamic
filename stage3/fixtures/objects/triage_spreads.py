"""Record source-Node and stage-0 lowering outcomes without building fixture binaries."""
import hashlib
import json
import subprocess
from pathlib import Path

bucket = Path(__file__).resolve().parent
repository = bucket.parents[2]
lanes = {
    "01": "spreads; conditional own-field presence",
    "02": "spreads; truthiness adaptation; closed-key override adaptation",
    "03": "spreads (resolved)",
    "04": "spreads; JSON serialization of structural object types",
    "05": "post-creation writes; array metadata wrapper adaptation",
    "06": "post-creation writes (README: declared-slot exception)",
    "07": "post-creation writes (README: declared-slot exception; already lowers)",
    "08": "post-creation writes (README: declared-slot exception); truthiness adaptation",
    "10": "post-creation writes (README: declared-slot exception); records",
    "12": "post-creation writes (README: declared-slot exception); chained assignment",
    "14": "definite-assignment marker; switch-fallthrough adaptation",
    "15": "definite-assignment marker; actual undefined result slots",
    "16": "destructuring with defaults",
    "17": "destructuring with array rest",
    "19": "class shapes; truthiness adaptation",
    "20": "optional calls; mutable-view invariance adaptation",
    "22": "optional calls; explicit undefined-return adaptation",
    "23": "getters and setters (already lowers)",
    "24": "getters and setters (already lowers)",
    "25": "generators (separate lane)",
    "26": "presence/discriminants (separate lane)",
    "27": "records and deletion adaptation (separate lane)",
    "28": "debugger removal adaptation (separate lane)",
    "29": "destructuring with defaults and paired parameter default",
    "30": "optional calls",
}
original = {row["file"]: row["stage0"]["outcome"] for row in json.loads((bucket / "status.json").read_text())}
rows = []
for item in json.loads((bucket / "manifest.json").read_text()):
    path = str((bucket / item["file"]).relative_to(repository))
    node = subprocess.run(["node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", path], cwd=repository, capture_output=True, text=True)
    stage0 = subprocess.run(["go", "run", "./cmd/adamic", "c", path], cwd=repository, capture_output=True, text=True)
    diagnostic = stage0.stderr.removesuffix("exit status 1\n")
    outcome = "Lowers" if stage0.returncode == 0 else "Refused" if "refuses" in diagnostic else "NotYet" if "can't lower" in diagnostic else "Checker" if "error TS" in diagnostic else "Error"
    row = dict(file=item["file"], source_sha256=hashlib.sha256((repository / path).read_bytes()).hexdigest(), lane=lanes[item["file"][:2]], original_outcome=original[item["file"]], stage0=dict(outcome=outcome, diagnostic=diagnostic, exit=stage0.returncode), node=dict(stdout=node.stdout, stderr=node.stderr, exit=node.returncode))
    rows.append(row)
    print(json.dumps(row), flush=True)
    if outcome == "Error" or node.returncode != 0:
        raise RuntimeError("unexpected outcome: " + path)
result = dict(integrated_main=subprocess.check_output(["git", "rev-parse", "origin/main"], cwd=repository, text=True).strip(), compiler_base="ef3d907ecdc4c771b016f7d9c52372def057a340", fixture_source="2632f9c", native_builds_in_this_script=0, fixture_count=len(rows), source_sites_in_inventory=17, registered_spread_representatives=["03_resolution_cache_spreads.a"], sites_represented_by_registered_fixture=["src/compiler/moduleNameResolver.ts:1302", "src/compiler/moduleNameResolver.ts:1303"], compiler_object_sha256=hashlib.sha256((repository / "internal/lower/object.go").read_bytes()).hexdigest(), note="This script performs no native comparison. The reduced cache fixture is registered under TestNativeAgreesWithNode; see the focused oracle log for its comparison. The 17-site original corpus has not been compiled or dynamically verified; original upstream operands are not proven closed by this run.", fixtures=rows)
(bucket / "spread_triage.json").write_text(json.dumps(result, indent=2) + "\n")
print(f"{len(rows)} fixtures triaged, {sum(row['stage0']['outcome'] == 'Lowers' for row in rows)} lower, no native binaries built", flush=True)
