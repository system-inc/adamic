"""Read the fetched snapshot that preceded this unit's second claim."""
import json
import re
import subprocess


def git(*args):
    return subprocess.check_output(["git", *args], text=True)


refs = git("for-each-ref", "--format=%(refname:short)", "refs/remotes/origin").splitlines()
claims = {}
for ref in refs:
    snapshot = "b4656c7c" if ref == "origin/codex/lint-wave1-12" else ref
    for row in git("ls-tree", "-r", snapshot, "stage1/cohere/lint/claims/").splitlines():
        metadata, path = row.split("\t")
        if path.endswith(".md"):
            blob = metadata.split()[2]
            if blob not in claims:
                claims[blob] = (ref, path, git("show", blob))
main = git("rev-parse", "origin/main").strip()
sources = []
for path in git("ls-tree", "-r", "--name-only", main, "stage1/cohere/lint").splitlines():
    if path.endswith((".a", ".ts")):
        sources.append(git("show", main + ":" + path))
source = "\n".join(sources)
claim_text = "\n".join(value[2] for value in claims.values())
helpers = git("show", "b4656c7c:stage1/cohere/lint/HELPERS.md")
helper_names = re.findall(r"^- `([^`]+)`", helpers.split("### Remaining option gaps")[0], re.M)
print("main=" + main)
print("origin_refs=" + str(len(refs)) + " unique_claim_markdown=" + str(len(claims)))
for name in helper_names:
    holders = [ref + ":" + path for ref, path, text in claims.values() if name in text]
    print("helper " + name + " excluded by " + (holders[0] if holders else "main" if name in source else "NONE"))
inventory = json.loads(git("show", "origin/codex/lint-inventory:stage1/cohere/lint/inventory/inventory.json"))
selected = [rule["name"] for rule in inventory["rules"]
            if not rule["needs_type_information"] and rule["name"] not in source
            and rule["name"] not in claim_text][:3]
print("syntax-only selection=" + json.dumps(selected))
assert selected == ["@next/next/no-before-interactive-script-outside-document", "@next/next/no-css-tags", "@next/next/no-document-import-in-page"]
