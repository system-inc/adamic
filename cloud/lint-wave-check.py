#!/usr/bin/env python3
"""Check a lint worker's ownership and run commit-bound parity/mutant evidence."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile


LINT = "stage1/cohere/lint/"
CLAIMS = LINT + "claims/"
PACKAGE = "github.com/system-inc/adamic/stage1/cohere/lint"
PARITY = ("TestRulesAgree", "TestOwnedWitnesses", "TestCompilerAndStage1Agree")
REQUIRED = PARITY + ("TestMutants", "TestRegistrationMutant")
NAME = re.compile(r"(?:@?[A-Za-z0-9_-]+/)*[A-Za-z0-9_-]+\Z")


class Rejected(Exception):
    def __init__(self, check, message):
        super().__init__(f"{check}: {message}")


def require(condition, check, message):
    if not condition:
        raise Rejected(check, message)


class Worker:
    def __init__(self, root):
        self.root = Path(root)
        self.blobs = {}
        self.trees = {}

    def git(self, *args):
        result = subprocess.run(["git", *args], cwd=self.root, capture_output=True)
        require(result.returncode == 0, "git", result.stderr.decode(errors="replace").strip())
        return result.stdout

    def text(self, *args):
        return self.git(*args).decode().strip()

    def tree(self, revision):
        if revision not in self.trees:
            files = {}
            for row in self.git("ls-tree", "-rz", revision, "--", LINT).split(b"\0"):
                if row:
                    metadata, path = row.split(b"\t", 1)
                    mode, kind, oid = metadata.decode().split()
                    require(kind == "blob", "registration", f"unexpected tree entry {path!r}")
                    files[path.decode()] = (mode, oid)
            self.trees[revision] = files
        return self.trees[revision]

    def blob(self, oid):
        if oid not in self.blobs:
            self.blobs[oid] = self.git("cat-file", "blob", oid).decode()
        return self.blobs[oid]

    def document(self, oid, check, location):
        try:
            return json.loads(self.blob(oid), object_pairs_hook=unique_keys)
        except (ValueError, UnicodeError) as error:
            raise Rejected(check, f"{location}: {error}") from error

    def clean(self):
        status = self.text("status", "--porcelain", "--untracked-files=all", "--ignore-submodules=none")
        require(not status, "evidence", "commit changes and remove untracked files first:\n" + status)

    def claim(self):
        branch = self.text("branch", "--show-current")
        require(branch, "claim", "a named worker branch is required")
        path = CLAIMS + branch + ".json"
        entry = self.tree("HEAD").get(path)
        require(entry is not None, "claim", f"missing committed {path}")
        require(entry[0] == "100644", "claim", f"{path} must be a regular non-executable file")
        document = self.document(entry[1], "claim", path)
        validate_claim(document, path)
        require(document["branch"] == branch, "claim", f"{path} belongs to {document['branch']}")
        return branch, path, document

    def origin(self):
        # A main-only remote.fetch is normal in cloud checkouts. Fetch explicitly,
        # prune stale remote heads, and avoid unrelated nested submodule fetches.
        self.git("fetch", "--prune", "--no-recurse-submodules", "origin",
                 "+refs/heads/*:refs/remotes/origin/*")
        refs = {}
        for row in self.text("for-each-ref", "--format=%(refname) %(objectname)",
                             "refs/remotes/origin/").splitlines():
            name, oid = row.split()
            if name != "refs/remotes/origin/HEAD":
                refs[name.removeprefix("refs/remotes/origin/")] = oid
        require("main" in refs, "origin", "origin/main is missing after fetch")
        return refs

    def overlaps(self, refs, branch, claim_path, claim):
        collisions = set()
        rules = set(claim["rules"])
        for remote_branch, oid in sorted(refs.items()):
            if remote_branch == branch:
                continue  # A worker can check again after publishing its own claim.
            for path, (mode, blob) in self.tree(oid).items():
                if path.startswith(CLAIMS) and path.endswith(".json"):
                    require(mode == "100644", "origin", f"origin/{remote_branch}:{path} is not a claim file")
                    other = self.document(blob, "origin", f"origin/{remote_branch}:{path}")
                    validate_claim(other, path)
                    # The same reservation can be inherited on several heads.
                    if path == claim_path and other == claim:
                        continue
                    for name in rules.intersection(other["rules"]):
                        collisions.add(f"{name} claimed in origin/{remote_branch}:{path}")
                elif path.startswith(LINT + "rules/") and path.endswith("/rule.json"):
                    other = self.document(blob, "origin", f"origin/{remote_branch}:{path}")
                    require(isinstance(other, dict) and isinstance(other.get("name"), str),
                            "origin", f"invalid descriptor origin/{remote_branch}:{path}")
                    if other["name"] in rules:
                        collisions.add(f"{other['name']} ported in origin/{remote_branch}:{path}")
                elif legacy_selection(path):
                    data = self.document(blob, "origin", f"origin/{remote_branch}:{path}")
                    entries = data.get("selection") if isinstance(data, dict) else data
                    require(isinstance(entries, list) and all(isinstance(e, dict) and isinstance(e.get("name"), str)
                                                              for e in entries),
                            "origin", f"unreadable legacy selection origin/{remote_branch}:{path}")
                    for name in rules.intersection(e["name"] for e in entries):
                        collisions.add(f"{name} reserved by legacy selection origin/{remote_branch}:{path}")
                elif re.fullmatch(re.escape(LINT) + r"BATCH[^/]*\.md", path):
                    source = self.blob(blob)
                    for name in rules:
                        if "`" + name + "`" in source:
                            collisions.add(f"{name} mentioned by legacy batch reservation origin/{remote_branch}:{path}")
                elif legacy_source(path):
                    # Conservative on purpose: a literal in a legacy production
                    # module may select/report a port. Fixtures and reports are excluded.
                    source = self.blob(blob)
                    for name in rules:
                        if re.search(r"(['\"])" + re.escape(name) + r"\1", source):
                            collisions.add(f"{name} referenced by legacy port origin/{remote_branch}:{path}")
        require(not collisions, "origin", "\n".join(sorted(collisions)))

    def registration(self, refs, claim_path, claim):
        base = self.text("merge-base", refs["main"], "HEAD")
        # Transitional prerequisite: main has not yet integrated registration.
        # Do not let an arbitrary --base turn a worker's shared edits into its baseline.
        if not descriptors(self, base):
            prerequisite = refs.get("codex/lint-registration")
            require(prerequisite is not None, "registration", "directory-registration prerequisite is absent on origin")
            ancestor = subprocess.run(["git", "merge-base", "--is-ancestor", prerequisite, "HEAD"], cwd=self.root)
            require(ancestor.returncode == 0, "registration",
                    "integrate origin/codex/lint-registration before adding a rule; main still uses shared registration")
            base = prerequisite
        current = descriptors(self, "HEAD")
        by_name = {data["name"]: (path, data) for path, data in current.items()}
        require(len(by_name) == len(current), "registration", "duplicate public rule names")
        owned = []
        mutants = {}
        for name in claim["rules"]:
            require(name in by_name, "registration", f"claimed rule {name} has no directory descriptor")
            path, data = by_name[name]
            directory = path.removesuffix("rule.json")
            require("order" not in data, "registration", f"new rule {name} must omit order")
            require(path not in self.tree(base), "registration", f"{name} already exists in the registration baseline")
            owned.append(directory)
            entry = self.tree("HEAD").get(directory + "mutant.json")
            require(entry is not None, "mutants", f"{name} has no owned mutant.json")
            mutant = self.document(entry[1], "mutants", directory + "mutant.json")
            label = mutant.get("name") if isinstance(mutant, dict) else None
            require(isinstance(label, str) and re.fullmatch(r"[A-Za-z0-9 _-]+", label),
                    "mutants", f"{name}: use a nonempty ASCII word/space/hyphen mutant name")
            mutants[name] = label
        require(len(set(mutants.values())) == len(mutants), "mutants", "owned mutant names must be distinct")
        # A rename out of a shared file still edits that shared file.
        paths = self.git("diff", "--no-renames", "--name-only", "-z", base, "HEAD").decode().split("\0")
        for path in filter(None, paths):
            if path.startswith(LINT) or path.startswith("cmd/lint-registry/"):
                require(path == claim_path or any(path.startswith(directory) for directory in owned),
                        "registration", f"shared or unclaimed lint change: {path}")
        generated = [p for p in self.tree("HEAD") if "/.generated/" in p]
        require(not generated, "registration", "generated registration must stay ignored: " + ", ".join(generated))
        return base, mutants

    def preflight(self):
        branch, path, claim = self.claim()
        self.clean()
        refs = self.origin()
        self.overlaps(refs, branch, path, claim)
        base, mutants = self.registration(refs, path, claim)
        print(f"PASS claim: {path}", flush=True)
        print(f"PASS origin: {len(refs)} fetched heads, no competing claims or ports", flush=True)
        print(f"PASS registration: owned directories only, baseline {base}", flush=True)
        return branch, path, claim, refs, mutants


def unique_keys(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError(f"duplicate JSON key {key}")
        result[key] = value
    return result


def validate_claim(document, path):
    require(isinstance(document, dict) and set(document) == {"version", "branch", "rules"},
            "claim", f"{path}: expected version, branch, rules")
    require(type(document["version"]) is int and document["version"] == 1,
            "claim", f"{path}: version must be 1")
    branch, rules = document["branch"], document["rules"]
    require(isinstance(branch, str) and branch and path == CLAIMS + branch + ".json",
            "claim", f"{path}: path must mirror branch")
    require(isinstance(rules, list) and rules and all(isinstance(n, str) and NAME.fullmatch(n) and n != "all" for n in rules),
            "claim", f"{path}: rules must be a nonempty list of public rule names")
    require(len(rules) == len(set(rules)), "claim", f"{path}: duplicate claimed rules")


def legacy_source(path):
    relative = path.removeprefix(LINT)
    return (path.endswith(".ts") and not any(part in {"testdata", "gaps", ".generated", "inventory"}
                                             for part in relative.split("/"))) or relative == "testdata/volume_rules.json"


def legacy_selection(path):
    return (path.endswith(".json") and "selection" in Path(path).name
            and not path.startswith(LINT + "inventory/"))


def descriptors(worker, revision):
    result = {}
    for path, (mode, oid) in worker.tree(revision).items():
        if re.fullmatch(re.escape(LINT) + r"rules/[a-z0-9]+(?:-[a-z0-9]+)*/rule.json", path):
            require(mode == "100644", "registration", f"{path} must be regular JSON")
            data = worker.document(oid, "registration", path)
            require(isinstance(data, dict) and isinstance(data.get("name"), str) and NAME.fullmatch(data["name"]),
                    "registration", f"{path}: invalid public name")
            result[path] = data
    return result


def verify_events(path, mutants):
    runs, passes, outputs = set(), set(), {}
    package_passed = False
    try:
        with path.open() as log:
            for line in log:
                event = json.loads(line)
                require(isinstance(event, dict), "evidence", "invalid go test event")
                if event.get("Package") != PACKAGE:
                    continue
                action, test = event.get("Action"), event.get("Test")
                require(action not in {"fail", "skip"}, "evidence", f"test {test} {action}")
                if action == "run":
                    runs.add(test)
                if action == "pass":
                    if test:
                        passes.add(test)
                    else:
                        package_passed = True
                if action == "output":
                    outputs[test] = outputs.get(test, "") + event.get("Output", "")
    except (OSError, ValueError) as error:
        raise Rejected("evidence", f"cannot read Go JSON log: {error}") from error
    require(package_passed, "evidence", "lint package did not pass")
    for test in REQUIRED:
        category = "mutants" if "Mutant" in test else "parity"
        require(test in runs and test in passes, category, f"{test} did not run and pass")
    for test in PARITY:
        require(re.search(r"Go, Node, native identical: [1-9][0-9]* bytes", outputs.get(test, "")),
                "parity", f"{test}: missing nonempty three-way comparison")
    require(re.search(r"cohere cases: [1-9][0-9]* unique", outputs.get("TestRulesAgree", "")),
            "parity", "upstream corpus is empty or unreported")
    for rule, label in mutants.items():
        test = "TestMutants/" + label.replace(" ", "_")
        require(test in runs and test in passes, "mutants", f"{rule}: owned mutant did not run and pass")
        for backend in ("Node", "native"):
            require(label + " caught on " + backend + ":" in outputs.get(test, ""),
                    "mutants", f"{rule}: no successful execution and caught comparison on {backend}")


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def check(root, verify_only=False):
    worker = Worker(root)
    branch, claim_path, claim, refs, mutants = worker.preflight()
    head = worker.text("rev-parse", "HEAD")
    directory = Path(worker.text("rev-parse", "--git-path", "lint-wave-check"))
    if not directory.is_absolute():
        directory = worker.root / directory
    receipt = directory / (branch + ".json")
    if verify_only:
        require(receipt.is_file(), "evidence", "no receipt; run this script without --verify first")
        try:
            record = json.loads(receipt.read_text())
            require(record["head"] == head and record["claim"] == claim and record["mutants"] == mutants,
                    "evidence", "receipt is stale for this commit, claim or mutant set")
            log = directory / record["log"]
            require(log.is_file() and digest(log) == record["sha256"], "evidence", "test log missing or changed")
        except (KeyError, ValueError, OSError) as error:
            raise Rejected("evidence", f"invalid receipt: {error}") from error
    else:
        require(os.environ.get("ADAMIC_TYPESCRIPT_SOURCE"), "parity", "set ADAMIC_TYPESCRIPT_SOURCE to the pinned v6.0.3 checkout")
        directory.mkdir(parents=True, exist_ok=True)
        receipt.parent.mkdir(parents=True, exist_ok=True)
        with tempfile.NamedTemporaryFile(dir=directory, suffix=".jsonl", delete=False) as output:
            log = Path(output.name)
            # Output goes directly to files. An exit code or a PASS string alone
            # does not establish that the selected tests/corpora/mutants executed.
            registry_log = log.with_suffix(".registry.txt")
            with registry_log.open("wb") as registry_output:
                result = subprocess.run(["go", "run", "./cmd/lint-registry"], cwd=worker.root,
                                        stdout=registry_output, stderr=subprocess.STDOUT)
            require(result.returncode == 0, "registration", f"generator failed; read {registry_log}")
            command = ["go", "test", "-json", "./stage1/cohere/lint", "-count=1", "-timeout=30m",
                       "-run", "^(" + "|".join(REQUIRED) + ")$"]
            print(f"Running parity and mutants; log: {log}", flush=True)
            with log.with_suffix(".stderr.txt").open("wb") as errors:
                result = subprocess.run(command, cwd=worker.root, stdout=output, stderr=errors)
        require(result.returncode == 0, "evidence", f"parity/mutant command failed; read {log} and {log.with_suffix('.stderr.txt')}")
        verify_events(log, mutants)
        worker.clean()
        require(worker.text("rev-parse", "HEAD") == head, "evidence", "HEAD changed while tests ran")
        refs = worker.origin()
        worker.overlaps(refs, branch, claim_path, claim)
        worker.registration(refs, claim_path, claim)
        record = {"head": head, "claim": claim, "mutants": mutants, "origin": refs,
                  "command": command, "log": log.name, "sha256": digest(log)}
        with tempfile.NamedTemporaryFile(dir=receipt.parent, mode="w", delete=False) as output:
            json.dump(record, output, indent=2)
            output.write("\n")
        os.replace(output.name, receipt)
    verify_events(log, mutants)
    print(f"PASS parity and mutants: current commit {head}; receipt {receipt}", flush=True)
    print("Ready for this worker's push. Origin was checked as a snapshot; concurrent pushes can still race.")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--verify", action="store_true", help="verify this commit's existing receipt instead of rerunning tests")
    args = parser.parse_args()
    try:
        root = subprocess.check_output(["git", "rev-parse", "--show-toplevel"], text=True).strip()
        check(root, args.verify)
    except (Rejected, subprocess.SubprocessError, OSError, UnicodeError) as error:
        print(f"lint-wave-check: FAIL {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
