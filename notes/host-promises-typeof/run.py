"""Build every async fixture and coverage probe and record independent observations."""
import json
import subprocess
import tempfile
from pathlib import Path

repository = Path(__file__).resolve().parents[2]
results = []
paths = sorted((repository / "internal/oracle/testdata").glob("async*.a"))
paths += sorted(Path(__file__).resolve().parent.glob("*.a"))
with tempfile.TemporaryDirectory(prefix="host-promises-typeof-") as scratch:
    for path in paths:
        output = str(Path(scratch) / path.stem)
        build = subprocess.run(["go", "run", "./cmd/adamic", "build", str(path), "-o", output], cwd=repository, capture_output=True, text=True)
        node = subprocess.run(["node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", str(path)], cwd=repository, capture_output=True, text=True)
        row = {"program": str(path.relative_to(repository)), "build_exit": build.returncode, "build_stderr": build.stderr.replace(str(repository) + "/", ""), "node_stdout": node.stdout, "node_stderr": node.stderr, "node_exit": node.returncode}
        if build.returncode == 0:
            native = subprocess.run([output], capture_output=True, text=True)
            row.update(native_stdout=native.stdout, native_stderr=native.stderr, native_exit=native.returncode)
            row["agrees"] = (node.stdout, node.stderr, node.returncode) == (native.stdout, native.stderr, native.returncode)
        results.append(row)
        print(row["program"], "build", build.returncode, "agrees", row.get("agrees", "no executable"), flush=True)
Path(__file__).with_name("observations.json").write_text(json.dumps(results, indent=2) + "\n")
