"""Reproduce selection, using the worker's pre-claim tip for its own ref."""
import json
import re
import subprocess


def git(*arguments):
    return subprocess.check_output(["git", *arguments], text=True)


refs = git("for-each-ref", "--format=%(refname)", "refs/remotes/origin").splitlines()
claims = {}
snapshots = {}
for ref in refs:
    tip = "68304ac0" if ref == "refs/remotes/origin/codex/lint-wave1-06" else ref
    snapshots[ref] = git("rev-parse", tip).strip()
    for path in git("ls-tree", "-r", "--name-only", tip, "stage1/cohere/lint/claims/").splitlines():
        if path.endswith(".md") and "/evidence/" not in path and "-evidence/" not in path:
            claims[ref + ":" + path] = git("show", tip + ":" + path)

ready = json.loads(git("show", "origin/codex/lint-helpers:stage1/cohere/lint/helpers/readiness.json"))
inventory = json.loads(git("show", "origin/codex/lint-inventory:stage1/cohere/lint/inventory/inventory.json"))
helper_names = ready["option_ready"] + ready["policy_increment"]
syntax_names = [rule["name"] for rule in inventory["rules"]
                if not rule["needs_type_information"] and not rule["binding_only"]]
paths = git("ls-tree", "-r", "--name-only", "origin/main", "stage1/cohere/lint").splitlines()
main_sources = {path: git("show", "origin/main:" + path) for path in paths
                if path.endswith((".ts", ".a", "rule.json")) and "/gaps/" not in path}
rows = []
selected = []
seen = set()
for name in helper_names + syntax_names:
    if name in seen:
        continue
    seen.add(name)
    pattern = re.compile(r"(?<![\w/-])" + re.escape(name) + r"(?![\w/-])")
    owners = [path for path, text in claims.items() if pattern.search(text)]
    ports = [path for path, text in main_sources.items() if pattern.search(text)]
    rows.append({"name": name, "helper_ready": name in helper_names,
                 "claims": owners, "main_source_references": ports})
    if not owners and not ports:
        selected.append(name)
        if len(selected) == 3:
            break
expected = ["@typescript-eslint/no-unnecessary-type-constraint",
            "@typescript-eslint/no-unsafe-function-type",
            "@typescript-eslint/no-useless-empty-export"]
assert selected == expected, selected
print(json.dumps({"origin_snapshots": snapshots, "selection": selected, "examined": rows}, indent=2))
