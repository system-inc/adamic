#!/usr/bin/env python3
"""Check lint output on unchanged repository inputs, including exact fix bytes.

The linter implementation itself can be a corpus file. Archive the starting
commit so editing it cannot masquerade as a changed lint judgment. Compare Go,
native and ASan stdout directly; normalize only the snapshot path prefix when
holding them to previously saved stdout from the original repository location.
"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess
import tarfile


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("repository", type=Path)
    parser.add_argument("commit")
    parser.add_argument("manifest", type=Path)
    parser.add_argument("binaries", type=Path)
    parser.add_argument("directory", type=Path)
    parser.add_argument("--baseline", nargs=2, action="append", required=True,
                        metavar=("RUNNER", "STDOUT"))
    args = parser.parse_args()
    repository = args.repository.resolve()
    directory = args.directory.resolve()
    directory.mkdir(parents=True, exist_ok=True)
    snapshot = directory / "repository-frozen"
    if snapshot.exists():
        raise RuntimeError("Snapshot directory must be fresh to exclude stale inputs")
    snapshot.mkdir()
    archive_path = directory / "repository.tar"
    subprocess.run(["git", "-C", str(repository), "archive", "--format=tar",
                    "-o", str(archive_path), args.commit], check=True)
    with tarfile.open(archive_path) as archive:
        archive.extractall(snapshot, filter="data")
    cohere = snapshot / "cohere"
    if not cohere.is_symlink():
        if cohere.exists():
            cohere.rmdir()
        cohere.symlink_to(repository / "cohere", target_is_directory=True)
    paths = [Path(p) for p in args.manifest.read_text().splitlines() if p]
    frozen = [snapshot / p.relative_to(repository) for p in paths]
    hashes = {}
    for original, copied in zip(paths, frozen):
        hashes[str(original)] = hashlib.sha256(copied.read_bytes()).hexdigest()
    manifest = directory / "repository-frozen.manifest"
    manifest.write_text("\n".join(map(str, frozen)) + "\n")
    (directory / "repository-sources.json").write_text(json.dumps(hashes, indent=2) + "\n")
    results = {}
    for runner, baseline in args.baseline:
        expected = Path(baseline).read_bytes()
        row = {}
        truth = None
        for label, executable in [("go", runner + "-oracle"), ("native", runner),
                                  ("native_asan", runner + "-asan")]:
            stem = directory / ("frozen-" + runner + "-" + label)
            with stem.with_suffix(".stdout").open("wb") as out, stem.with_suffix(".stderr").open("wb") as err:
                subprocess.run([str(args.binaries.resolve() / executable),
                                str(snapshot / "tsconfig.json"), str(manifest)],
                               stdout=out, stderr=err, check=True)
            stderr = stem.with_suffix(".stderr").read_bytes()
            if label == "go":
                if re.fullmatch(rb"cohere: load_ns=\d+ rule_ns=\d+ run_ns=\d+\n", stderr) is None:
                    raise RuntimeError(f"Unexpected oracle stderr: {stem}")
            elif stderr:
                raise RuntimeError(f"Native or sanitizer stderr: {stem}")
            raw = stem.with_suffix(".stdout").read_bytes()
            if truth is not None and raw != truth:
                raise RuntimeError(f"Native output differs from Go: {stem}")
            truth = raw
            canonical = raw.replace(str(snapshot).encode(), str(repository).encode())
            if canonical != expected:
                raise RuntimeError(f"Output differs from previous findings/fixes: {stem}")
            row[label] = dict(raw_sha256=hashlib.sha256(raw).hexdigest(),
                              canonical_sha256=hashlib.sha256(canonical).hexdigest(),
                              canonical_bytes=len(canonical))
        row["previous_native_sha256"] = hashlib.sha256(expected).hexdigest()
        results[runner] = row
    (directory / "frozen-findings.json").write_text(json.dumps(results, indent=2) + "\n")
    print(json.dumps(results, indent=2))


if __name__ == "__main__":
    main()
