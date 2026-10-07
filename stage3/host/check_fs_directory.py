#!/usr/bin/env python3
"""Gate the five owned upstream host fixtures against their recorded Node bytes.

Unlike the historical check.py, successful compilation is required and both
backends run. No recorded stage0 diagnostic or fixture source is rewritten.
"""
import argparse
import json
import os
import sys
from pathlib import Path
import subprocess
import tempfile

repo = Path(__file__).resolve().parents[2]
bucket = repo / "stage3/fixtures/host"
owned = {
    "07_directoryExists.a": ("return stat.isDirectory();", "return stat.isFile();"),
    "08_getDirectories.a": ("directories.sort();", "directories.reverse();"),
    "09_realpath.a": ("return path;", "return _path.resolve(path);"),
    "24_useCaseSensitiveFileNames.a": (
        "return !fileExists(swapCase(__filename));",
        "return fileExists(swapCase(__filename));",
    ),
    "25_readDirectory.a": (
        "if (extensions && !fileExtensionIsOneOf(name, extensions)) continue;",
        "if (false) continue;",
    ),
}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--logs", required=True, type=Path)
    parser.add_argument("--mutants", action="store_true")
    args = parser.parse_args()
    logs = args.logs.resolve()
    logs.mkdir(parents=True, exist_ok=True)
    recorded = {row["file"]: row["node"] for row in json.loads((bucket / "status.json").read_text())}
    results = []
    failed = False

    def run(command, label, environment=None):
        result = subprocess.run(command, cwd=repo, capture_output=True, env=environment, timeout=600)
        (logs / (label + ".stdout")).write_bytes(result.stdout)
        (logs / (label + ".stderr")).write_bytes(result.stderr)
        (logs / (label + ".exit")).write_text(str(result.returncode) + "\n")
        return dict(stdout=result.stdout.decode(), stderr=result.stderr.decode(), exit=result.returncode)

    def node(path, label):
        return run(["node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", str(path)], label)

    def stage(built):
        if built["exit"] == 0:
            return "Compiles"
        diagnostic = built["stdout"] + built["stderr"]
        if ": error TS" in diagnostic:
            return "Checker"
        if "stage 0 can't lower" in diagnostic:
            return "NotYet"
        if "Adamic 0.1 refuses" in diagnostic:
            return "Refused"
        return "ToolFailure"

    with tempfile.TemporaryDirectory(prefix="adamic-fs-directory-acceptance-") as scratch_name:
        scratch = Path(scratch_name)
        compiler = scratch / "adamic"
        built = run(["go", "build", "-o", str(compiler), "./cmd/adamic"], "compiler")
        if built["exit"] != 0:
            raise RuntimeError("compiler build failed; see compiler.stderr")
        for name, mutation in owned.items():
            source = bucket / name
            truth = node(source, name + ".node")
            row = dict(file=name, node="Agrees" if truth == recorded[name] else "DIFFERS", backends={})
            failed |= truth != recorded[name]
            for backend in ("native", "JavaScript"):
                label = name + "." + backend
                binary = scratch / (name + ".binary")
                command = [str(compiler), "build", str(source), "-o", str(binary), "--sanitize"] if backend == "native" else [str(compiler), "js", str(source)]
                emitted = run(command, label + ".compile")
                outcome = stage(emitted)
                detail = dict(stage=outcome)
                if outcome == "Compiles":
                    if backend == "native":
                        environment = dict(os.environ)
                        if sys.platform == "linux":
                            environment["ASAN_OPTIONS"] = "detect_leaks=1"
                        observed = run([str(binary)], label + ".run", environment)
                    else:
                        module = scratch / (name + ".mjs")
                        module.write_text(emitted["stdout"])
                        observed = node(module, label + ".run")
                    detail["comparison"] = "Agrees" if observed == recorded[name] else "DIFFERS"
                    failed |= observed != recorded[name]
                else:
                    failed = True
                row["backends"][backend] = detail
            if args.mutants:
                before, after = mutation
                text = source.read_text()
                assert before in text, (name, before)
                mutant = scratch / name
                mutant.write_text(text.replace(before, after, 1))
                changed = node(mutant, name + ".mutant")
                # A crash cannot prove that the semantic comparison can fail.
                caught = changed["exit"] == truth["exit"] and changed["stderr"] == truth["stderr"] and changed["stdout"] != truth["stdout"]
                row["mutant"] = "Caught only by stdout comparison" if caught else "FAILED"
                failed |= not caught
            results.append(row)
            print(name + ": " + json.dumps(row, ensure_ascii=False), flush=True)
    (logs / "results.json").write_text(json.dumps(results, ensure_ascii=False, indent=2) + "\n")
    return 1 if failed else 0


if __name__ == "__main__":
    raise SystemExit(main())
