#!/usr/bin/env python3
"""Run valid executable backend mutants, retaining complete output and restoring sources."""
from pathlib import Path
import os
import json
import subprocess

repository = Path(__file__).resolve().parents[3]
emitter = repository / "internal/javascript/javascript.go"
original = emitter.read_text()
guard = 'e.line("if (%s instanceof RangeError && %s.message === %s) panic(%s);", caught, caught, quote("Maximum call stack size exceeded"), quote(stackOverflowMessage))'
mutants = [
    ("unlisted-output", 'e.line("console.%s(%s);", stream, e.value(statement.Value))',
     'e.line("console.%s(%s + \'!\');", stream, e.value(statement.Value))',
     "TestUnlistedBackendAgreement", "unlisted backend divergence: stdout differs"),
    ("drop-terminal-guard", guard, "// Mutant: omit the terminal stack check.",
     "TestTerminalStackStop", "terminal stack stop: exit codes differ"),
    ("drop-range-error-kind", "if (%s instanceof RangeError && %s.message === %s)",
     "if (%s && %s.message === %s)",
     "TestNativeAgreesWithNode/internal/oracle/testdata/backend_stop_catches", "exit codes differ"),
    ("drop-overflow-message", "if (%s instanceof RangeError && %s.message === %s)",
     "if (%s instanceof RangeError && (%s.message || %s))",
     "TestNonOverflowRangeErrorCatch", "non-overflow RangeError was terminal: exit codes differ"),
]
logs = repository / "review/compiler/chain-slice-5/logs/backend-stop-mutants"
logs.mkdir(exist_ok=True)
environment = dict(os.environ, ADAMIC_GATE_UNCACHED="1")
for name, before, after, test, catcher in mutants:
    if original.count(before) != 1:
        raise SystemExit(f"{name}: expected one mutation site")
    try:
        changed = logs / (name + ".go.txt")
        changed.write_text(original.replace(before, after, 1))
        overlay = logs / (name + ".json")
        overlay.write_text(json.dumps({"Replace": {str(emitter): str(changed)}}))
        log = logs / (name + ".log")
        with log.open("wb") as output:
            result = subprocess.run(["go", "test", "-overlay=" + str(overlay), "./internal/oracle", "-run", test,
                                     "-count=1", "-v", "-timeout", "90s"],
                                    cwd=repository, env=environment, stdout=output,
                                    stderr=subprocess.STDOUT, timeout=150)
        observed = log.read_text()
        if result.returncode != 1 or catcher not in observed or "[build failed]" in observed:
            raise SystemExit(f"{name}: invalid kill, exit {result.returncode}; inspect {log}")
        print(f"{name}: exit 1, caught by {catcher}; {log}", flush=True)
    finally:
        pass
