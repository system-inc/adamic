"""Compare prepared JSX rule cores on source Node and emitted JavaScript."""
import argparse
import json
import pathlib
import subprocess
import time

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("repository", type=pathlib.Path)
parser.add_argument("artifacts", type=pathlib.Path)
parser.add_argument("output", type=pathlib.Path)
args = parser.parse_args()
args.output.mkdir(parents=True, exist_ok=True)


def run(name, command):
    with (args.output / (name + ".stdout")).open("wb") as stdout, (
        args.output / (name + ".stderr")
    ).open("wb") as stderr:
        started = time.monotonic()
        result = subprocess.run(command, cwd=args.repository, stdout=stdout, stderr=stderr)
        elapsed = time.monotonic() - started
    if result.returncode:
        raise RuntimeError(f"{name}: exit {result.returncode}")
    return (
        (args.output / (name + ".stdout")).read_bytes(),
        (args.output / (name + ".stderr")).read_bytes(),
        elapsed,
    )


results = []
runner = args.repository / "oracle/node.mjs"
for entry in sorted(args.artifacts.glob("default-*.a")):
    name = entry.stem
    expected, _, _ = run(
        name + "-go",
        [str(args.artifacts / "oracle"), str(args.artifacts / "config.json"), str(args.artifacts / (name + ".manifest"))],
    )
    source, errors, source_time = run(
        name + "-source",
        ["node", "--disable-warning=ExperimentalWarning", str(runner), str(entry)],
    )
    if errors or source != expected:
        raise RuntimeError(f"{name}: source Node/Go mismatch")
    emitted, errors, build_time = run(name + "-emit", [str(args.artifacts / "adamic"), "js", str(entry)])
    if errors:
        raise RuntimeError(f"{name}: JavaScript emission stderr")
    js = args.output / (name + ".mjs")
    js.write_bytes(emitted)
    actual, errors, node_time = run(
        name + "-emitted", ["node", "--disable-warning=ExperimentalWarning", str(runner), str(js)]
    )
    if errors or actual != expected:
        raise RuntimeError(f"{name}: emitted JavaScript/Go mismatch")
    results.append({"batch": name, "bytes": len(expected), "source_node_seconds": source_time,
                    "emitted_node_seconds": node_time, "emit_seconds": build_time})
(args.output / "results.json").write_text(json.dumps(results, indent=2) + "\n")
print(f"{len(results)} default batches: source Node and emitted JavaScript match production Go byte for byte")
