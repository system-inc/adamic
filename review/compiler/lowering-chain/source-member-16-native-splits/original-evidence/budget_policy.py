"""Prove reference-box enforcement and log-only behavior on other boxes."""
import argparse
import contextlib
import io
import json
import os
from pathlib import Path
import subprocess
import sys

from budget import check


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("output", type=Path)
    output = parser.parse_args().output.resolve()
    output.mkdir(parents=True, exist_ok=False)
    results = []
    original = os.environ.get("ADAMIC_UNIT_BUDGET")
    try:
        for flag in ["1", "0", None]:
            if flag is None:
                os.environ.pop("ADAMIC_UNIT_BUDGET", None)
            else:
                os.environ["ADAMIC_UNIT_BUDGET"] = flag
            for seconds in [30.0, 30.001]:
                warning = io.StringIO()
                failed = False
                with contextlib.redirect_stderr(warning):
                    try:
                        check({"probe": dict(Action="pass", Elapsed=seconds)})
                    except ValueError:
                        failed = True
                assert failed == (flag == "1" and seconds > 30)
                assert bool(warning.getvalue()) == (flag != "1" and seconds > 30)
                results.append(dict(flag=flag, leaf_seconds=seconds,
                                    rejected=failed, warning=warning.getvalue()))
                directory = output / f"{flag}-{seconds}"
                directory.mkdir()
                event = dict(Action="pass", Test="probe", Elapsed=1.0)
                row = dict(package="internal/native", test="probe", exit=0,
                           unit_selector=True, wall=seconds, result=event)
                (directory / "probe.jsonl").write_text(json.dumps(event) + "\n")
                (directory / "results.jsonl").write_text(json.dumps(row) + "\n")
                command = [sys.executable, str(Path(__file__).with_name("budget.py")), str(directory)]
                log = directory / "audit.log"
                with log.open("w") as stream:
                    completed = subprocess.run(command, stdout=stream, stderr=stream)
                assert (completed.returncode != 0) == (flag == "1" and seconds > 30)
                if seconds > 30:
                    assert "invocation took 30.001 seconds" in log.read_text()
                results.append(dict(flag=flag, invocation_seconds=seconds,
                                    exit=completed.returncode, command=command))
            try:
                check({"incorrect": dict(Action="fail", Elapsed=0)})
            except ValueError:
                results.append(dict(flag=flag, correctness_failure_rejected=True))
            else:
                raise AssertionError("correctness failure was ignored")
    finally:
        if original is None:
            os.environ.pop("ADAMIC_UNIT_BUDGET", None)
        else:
            os.environ["ADAMIC_UNIT_BUDGET"] = original
    (output / "results.json").write_text(json.dumps(results, indent=2) + "\n")
    print("PASS: leaf and invocation boundaries; reference enforcement; log-only elsewhere; correctness in all modes")


if __name__ == "__main__":
    main()
