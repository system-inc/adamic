#!/usr/bin/env python3
"""manifest.py: the admission-delta corpora at one sha, as the JSON manifest cmd/adamic-admission-delta reads with
--manifest (@system_adamic's ruling, Oct 9 07:20; #9dse25n, @system_adamic_tests). Each corpus is a list of program
paths pinned by git blob hash at the sha, read from git, never from the working tree, so the manifest is the sha's own.

	cloud/admission-corpus/manifest.py [--sha <sha>] [--repository <path>] > manifest.json

The corpora, in the order a deterministic sample draws them:
- witnesses: small programs the test audit found that separate a miscompile from a correct build (their README says
  which), review/admission-corpus/witnesses/; the cheapest and sharpest, so the tool should never sample them away.
- fixtures: the oracle's programs, the loader's 0.1 compile programs and the flow programs.
- gaps: the stage1 ports' recorded gaps, programs Adamic refused on purpose; a change that newly admits one is
  exactly what the proof is for.
- review: the programs kept under review/ (outside review/test-audit and review/admission-corpus).
- fuzz: the fuzzer's kept programs, once it writes them to the tree (#s2n8gc6); empty until then, and said so.
"""
import argparse
import fnmatch
import json
import subprocess
import sys

corpora = [
    ("witnesses", ["review/admission-corpus/witnesses/*.a"], []),
    ("fixtures", ["internal/oracle/testdata/*.a", "internal/load/testdata/0.1/compile/*.ts", "internal/load/testdata/0.1/compile/*/main.ts", "internal/flow/testdata/*.a"], []),
    ("gaps", ["stage1/*/gaps/*.ts", "stage1/*/*/gaps/*.ts", "stage1/*/*/gaps/*.a", "stage1/*/gaps/*.a"], []),
    ("review", ["review/*/*.a", "review/*/*/*.a"], ["review/test-audit/*", "review/admission-corpus/*"]),
    ("fuzz", ["internal/fuzz/corpus/*.a"], []),
]


def matches(path, patterns):
    # fnmatch's * crosses slashes; a pattern's segment count must equal the path's, so * stays one segment.
    return any(fnmatch.fnmatchcase(path, pattern) and pattern.count("/") == path.count("/") for pattern in patterns)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--sha", default="HEAD")
    parser.add_argument("--repository", default=".")
    arguments = parser.parse_args()
    sha = subprocess.run(["git", "-C", arguments.repository, "rev-parse", arguments.sha], check=True, capture_output=True, text=True).stdout.strip()
    listing = subprocess.run(["git", "-C", arguments.repository, "ls-tree", "-r", "--full-tree", sha], check=True, capture_output=True, text=True).stdout
    blobs = {}
    for line in listing.splitlines():
        meta, path = line.split("\t", 1)
        mode, kind, blob = meta.split()
        if kind == "blob":
            blobs[path] = blob
    manifest = {"sha": sha, "generator": "cloud/admission-corpus/manifest.py", "corpora": []}
    for name, include, exclude in corpora:
        programs = [{"path": path, "blob": blobs[path]} for path in sorted(blobs) if matches(path, include) and not any(fnmatch.fnmatchcase(path, pattern) for pattern in exclude)]
        manifest["corpora"].append({"name": name, "patterns": include, "programs": programs})
    json.dump(manifest, sys.stdout, indent=1)
    sys.stdout.write("\n")
    print(" ".join("%s=%d" % (corpus["name"], len(corpus["programs"])) for corpus in manifest["corpora"]), file=sys.stderr)


if __name__ == "__main__":
    main()
