"""Measure complete batch processes; every sample must produce the Go oracle's bytes."""
import argparse
import json
import os
from pathlib import Path
import statistics
import subprocess
import tempfile
import time

parser = argparse.ArgumentParser()
parser.add_argument("corpus", type=Path)
parser.add_argument("native", type=Path)
parser.add_argument("go", type=Path)
parser.add_argument("prettier", type=Path)
parser.add_argument("--runs", type=int, default=3)
parser.add_argument("--entry", default="main.ts", choices=["main.ts", "statementsMain.ts"])
arguments = parser.parse_args()
root = Path(__file__).resolve().parents[4]
port = root / "stage1/cohere/tsprinter" / arguments.entry
expected = (arguments.corpus / "answers.txt").read_bytes()
count = len(expected.splitlines())
commands = {
    "native": [str(arguments.native), "--cases", str(arguments.corpus / "cases.txt"), "80"],
    "Node": ["node", "--disable-warning=ExperimentalWarning", str(root / "oracle/node.mjs"), str(port), "--cases", str(arguments.corpus / "cases.txt"), "80"],
    "Go": [str(arguments.go), "-test.run=^TestAdamicExpressionBenchmark$"],
    "Prettier": ["node", str(port.parent / "testdata/expressions.mjs"), str(arguments.prettier), str(arguments.corpus / "cases.json")],
}
results = {"entry": arguments.entry, "texts": count, "width": 80, "runs": arguments.runs, "method": "sequential complete batch processes including input, output and startup", "samples": {}}
with tempfile.TemporaryDirectory(prefix="ts-printer-measure-") as temporary:
    directory = Path(temporary)
    for name, command in commands.items():
        durations = []
        for iteration in range(arguments.runs):
            output = directory / "output.txt"
            environment = os.environ.copy()
            if name == "Go":
                environment.update(ADAMIC_TS_BENCH_CASES=str(arguments.corpus / "cases.txt"), ADAMIC_TS_BENCH_OUTPUT=str(output), ADAMIC_TS_BENCH_WIDTH="80")
            with (directory / "stdout.txt" if name == "Go" else output).open("wb") as stdout, (directory / "stderr.txt").open("wb") as stderr:
                start = time.perf_counter()
                result = subprocess.run(command, stdout=stdout, stderr=stderr, env=environment, cwd=root, check=False)
                duration = time.perf_counter() - start
            assert result.returncode == 0, (name, result.returncode)
            assert (directory / "stderr.txt").read_bytes() == b"", name
            assert output.read_bytes() == expected, f"{name} sample {iteration} changed bytes"
            durations.append(duration)
        results["samples"][name] = {"seconds": durations, "median_texts_per_second": count / statistics.median(durations)}
print(json.dumps(results, indent=2))
