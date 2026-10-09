from pathlib import Path
import subprocess

hook = Path("internal/lower/optional_node_host.go")
rule = Path("internal/lower/optional_widening.go")
original_hook, original_rule = hook.read_text(), rule.read_text()

def run(name, expected):
    with open("/tmp/optional3-" + name + ".log", "w") as log:
        result = subprocess.run(["go", "test", "./internal/lower", "-run", "TestOptionalWideningRefused/(argument|field_access_paths)$", "-count=1"], stdout=log, stderr=subprocess.STDOUT)
    print(name, "exit", result.returncode, flush=True)
    if (result.returncode == 0) != (expected == 0):
        raise SystemExit("unexpected mutant result: " + name)

try:
    hook.write_text(original_hook.replace("return false", "return true"))
    run("predicate-true-guarded", 0)
    rule.write_text(original_rule.replace(" && l.optionalNodeHostCall(parent)", ""))
    run("non-node-exemption-mutant", 1)
    hook.write_text(original_hook)
    rule.write_text(original_rule.replace("which can hide fields", "which hides fields"))
    run("pinned-message-mutant", 1)
finally:
    hook.write_text(original_hook)
    rule.write_text(original_rule)
