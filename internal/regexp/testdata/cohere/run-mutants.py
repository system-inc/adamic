#!/usr/bin/env python3
"""Run the requested mutants, restoring both files even when a run fails."""
import gzip
import json
import os
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[4]
corpus = Path(__file__).resolve().parent
observations_path = corpus / "observations.json.gz"
test_path = root / "internal/regexp/cohere_patterns_test.go"
logs = Path(sys.argv[1] if len(sys.argv) > 1 else "/tmp/regex-cohere-mutants")
logs.mkdir(parents=True, exist_ok=True)
environment = dict(os.environ)
environment.pop("ADAMIC_COHERE_RECORD", None)


def run(name, pattern, expected):
    with (logs / (name + ".log")).open("wb") as log:
        result = subprocess.run(
            ["go", "test", "./internal/regexp", "-run", pattern,
             "-count=1", "-timeout", "10m", "-v"],
            cwd=root, env=environment, stdout=log, stderr=subprocess.STDOUT,
            timeout=660,
        )
    output = (logs / (name + ".log")).read_text()
    if result.returncode == 0 or expected not in output:
        raise RuntimeError(f"{name} survived or failed for the wrong reason: {output}")
    print(f"{name}: killed, exit={result.returncode}; {logs / (name + '.log')}")


original_observations = observations_path.read_bytes()
original_test = test_path.read_bytes()
try:
    observations = json.loads(gzip.decompress(original_observations))
    for number, case in enumerate(observations["cases"]):
        captures = case["expected"]["captures"]
        if captures and len(captures) > 1 and captures[1] is not None:
            captures[1][1] += 1
            print(f"flip expected capture: pattern={case['pattern']}, case={number}")
            break
    else:
        raise RuntimeError("no real recorded capture")
    observations_path.write_bytes(gzip.compress(json.dumps(observations).encode(), mtime=0))
    run("expected-capture", "^TestCoherePatterns$", "cohere failure inventory changed")
finally:
    observations_path.write_bytes(original_observations)

try:
    source = original_test.decode()
    comparison = '''\tif !reflect.DeepEqual(got.Groups, want.Groups) || !reflect.DeepEqual(got.GroupValues, want.GroupValues) {
\t\treturn "named groups"
\t}
'''
    if source.count(comparison) != 1:
        raise RuntimeError("named comparison must have exactly one mutation site")
    test_path.write_text(source.replace(comparison, "", 1))
    run("omit-named-comparison", "^TestCohereNamedGroupGuard$", "named-only mutant survived")
finally:
    test_path.write_bytes(original_test)
extra = [
    ("omit-index-comparison", "index", "index-only mutant survived", '\tif !reflect.DeepEqual(got.Index, want.Index) {\n\t\treturn "match index"\n\t}\n'),
    ("omit-value-comparison", "values", "values-only mutant survived", '\tif !reflect.DeepEqual(got.Values, want.Values) {\n\t\treturn "capture values"\n\t}\n'),
    ("omit-lastindex-comparison", "lastIndex", "lastIndex-only mutant survived", '\tif got.LastIndex != want.LastIndex {\n\t\treturn "lastIndex"\n\t}\n'),
    ("omit-named-value-comparison", "named_values", "named values-only mutant survived", " || !reflect.DeepEqual(got.GroupValues, want.GroupValues)"),
]
for name, subtest, expected, comparison in extra:
    try:
        source = original_test.decode()
        if source.count(comparison) != 1:
            raise RuntimeError(f"{name}: expected exactly one mutation site")
        test_path.write_text(source.replace(comparison, "", 1))
        run(name, "^TestCohereObservationGuards/" + subtest + "$", expected)
    finally:
        test_path.write_bytes(original_test)
print("All files restored. No matcher, compiler, or runtime source was mutated.")
