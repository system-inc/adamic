#!/usr/bin/env python3
"""manifest.py: the admission-delta corpora at one sha, as the JSON manifest cmd/adamic-admission-delta reads with
--manifest (@system_adamic's ruling, Oct 9 07:20; #9dse25n, @system_adamic_tests). Each corpus is a list of program
paths pinned by git blob hash at the sha, read from git, never from the working tree, so the manifest is the sha's own.

	cloud/admission-corpus/manifest.py [--sha <sha>] [--repository <path>] > manifest.json

The corpora, in the order a deterministic sample draws them:
- witnesses: small programs the test audit found that separate a miscompile from a correct build (their README says
  which), review/admission-corpus/witnesses/; the cheapest and sharpest, so the tool should never sample them away.
- fixtures: the oracle's programs at every depth, the loader's 0.1 compile programs, the flow, lowering and
  conformance programs.
- gaps: the stage1 ports' recorded gaps, programs Adamic refused on purpose; a change that newly admits one is
  exactly what the proof is for.
- review: the programs kept under review/ (outside review/test-audit and review/admission-corpus).
- fuzz: the fuzzer's kept programs, once it writes them to the tree (#s2n8gc6); empty until then, and said so.
- stage3: the stage 3 programs at every depth, except stage3/drivers/tsc/corpus, the test inputs the tsc driver
  feeds to a compiled tsc as data, which are tsc's programs, not Adamic's.

Every corpus collects `.ts` beside `.a` (#nc3k59z): a fixture moved to its `.ts` twin stays in the corpus, so a delta
of 0 means nothing was newly refused, never that the programs left. A `**` segment matches any number of
directories, so a nested directory is never silently outside its corpus.
"""
import argparse
import fnmatch
import json
import re
import subprocess
import sys

corpora = [
    ("witnesses", ["review/admission-corpus/witnesses/*.a", "review/admission-corpus/witnesses/*.ts"], []),
    ("fixtures", ["internal/oracle/testdata/**/*.a", "internal/oracle/testdata/**/*.ts", "internal/load/testdata/0.1/compile/*.ts", "internal/load/testdata/0.1/compile/*/main.ts", "internal/flow/testdata/**/*.a", "internal/flow/testdata/**/*.ts", "internal/lower/testdata/**/*.a", "internal/lower/testdata/**/*.ts", "internal/conformance/testdata/**/*.a", "internal/conformance/testdata/**/*.ts"], []),
    ("gaps", ["stage1/*/gaps/*.ts", "stage1/*/*/gaps/*.ts", "stage1/*/*/gaps/*.a", "stage1/*/gaps/*.a"], []),
    ("review", ["review/*/*.a", "review/*/*.ts", "review/*/*/*.a", "review/*/*/*.ts"], ["review/test-audit/*", "review/admission-corpus/*"]),
    ("fuzz", ["internal/fuzz/corpus/*.a", "internal/fuzz/corpus/*.ts"], []),
    ("stage3", ["stage3/**/*.a", "stage3/**/*.ts"], ["stage3/drivers/tsc/corpus/*"]),
]


def segments(pattern):
    # A pattern as a regular expression: * and ? stay inside one segment, and a `**/` segment matches zero or more
    # directories.
    expression = ""
    for part in re.split(r"(\*\*/|\*|\?)", pattern):
        if part == "**/":
            expression += "(?:[^/]+/)*"
        elif part == "*":
            expression += "[^/]*"
        elif part == "?":
            expression += "[^/]"
        else:
            expression += re.escape(part)
    return re.compile(expression + r"\Z")


def matches(path, patterns):
    return any(segments(pattern).match(path) for pattern in patterns)


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
