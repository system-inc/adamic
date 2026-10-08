import argparse
import json
from pathlib import Path
import subprocess
import tempfile

parser = argparse.ArgumentParser()
parser.add_argument("library_worktree", type=Path, help="worktree containing the real node host predicate and this patch")
args = parser.parse_args()
root = args.library_worktree.resolve()
rule = root / "internal/lower/optional_widening.go"
source = rule.read_text()
old = "ast.SkipParentheses(argument) == node"
if source.count(old) != 1:
    raise SystemExit("expected one argument skip")
with tempfile.TemporaryDirectory(prefix="optional-parens-mutant-") as scratch:
    scratch = Path(scratch)
    mutant = scratch / "optional_widening.go"
    mutant.write_text(source.replace(old, "argument == node", 1))
    overlay = scratch / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {str(rule): str(mutant)}}))
    for variant in ["argument", "nested"]:
        log = Path("/tmp/optional-parens-mutant-" + variant + ".log")
        with log.open("w") as output:
            result = subprocess.run(["go", "test", "-overlay", str(overlay), "./internal/oracle", "-run", "^TestOptionalWideningNodeParentheses$/^" + variant + "$", "-count=1", "-timeout=10m"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        report = log.read_text()
        if result.returncode != 1 or "adamic/no-optional-widening" not in report or "--- FAIL:" not in report:
            raise SystemExit("mutant was not caught by the refusal: " + variant + "; inspect " + str(log))
        print(variant + ": skip removal caught by optional-widening refusal", flush=True)
