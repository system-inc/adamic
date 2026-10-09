#!/usr/bin/env python3
"""Content identities for the gate ledger; checkout paths never enter the digest."""
import argparse
import glob
import hashlib
import json
import os
import subprocess
import sys
import tempfile
import shutil
import re

from product_inputs import bounded_run, prepare_overlay, read_recipes
import time

BOX_PATHS = ('cloud/fast-gate.sh cloud/fast-gate cloud/darwin-leg.sh '
             'cloud/fast-gate-classify.sh cloud/idle-preempt.sh internal/skipcensus go.mod go.sum').split()


def tools_fingerprint(tree):
    # Byte-for-byte boxTools(): non-recursive ls-tree, including its final newline, then SHA-1.
    listing = subprocess.run(['git', '-C', tree, 'ls-tree', 'HEAD', '--'] + BOX_PATHS,
                             capture_output=True, check=True, timeout=10).stdout
    return hashlib.sha1(listing).hexdigest()


def objects(text):
    decoder = json.JSONDecoder()
    while text.strip():
        value, end = decoder.raw_decode(text.lstrip())
        yield value
        text = text.lstrip()[end:]


class InputHashes:
    def __init__(self, tree):
        self.tree = os.path.realpath(tree)
        self.packages = None
        self.cache = {}
        self.product_key_binary = None
        self.product_scratch = None
        self.deadline = time.monotonic() + 120

    def check_time(self):
        if time.monotonic() >= self.deadline:
            raise TimeoutError('unit input hashing exceeded 120 s')

    def inventory(self):
        if self.packages is not None:
            return
        self.check_time()
        result = subprocess.run(['go', 'list', '-deps', '-test', '-e', '-json', './...'],
                                cwd=self.tree, capture_output=True, text=True,
                                check=True, timeout=min(90, max(1, self.deadline - time.monotonic())))
        self.packages = {}
        for row in objects(result.stdout):
            name = row['ImportPath'].split(' [', 1)[0]
            if not row.get('ForTest') and not name.endswith('.test'):
                self.packages[name] = row

    def package_files(self, roots):
        if not roots:
            return {}
        self.inventory()
        files, seen, pending = {}, set(), list(roots)
        while pending:
            self.check_time()
            name = pending.pop().split(' [', 1)[0]
            if name in seen or name == 'C':
                continue
            seen.add(name)
            row = self.packages.get(name)
            if row is None or row.get('Error') or row.get('DepsErrors'):
                raise ValueError('incomplete go list closure: ' + name)
            pending.extend(row.get('Imports', []))
            if name in roots:
                pending.extend(row.get('TestImports', []) + row.get('XTestImports', []))
            directory = row.get('Dir')
            if not directory:
                continue  # unsafe has no source directory.
            paths = set(glob.glob(os.path.join(directory, '*.go')))
            for field in ('EmbedFiles', 'TestEmbedFiles', 'XTestEmbedFiles'):
                paths.update(os.path.join(directory, path) for path in row.get(field, []))
            for path in paths:
                files['go:' + name + '/' + os.path.relpath(path, directory)] = path
        return files

    def product_workspace(self):
        if self.product_scratch is None:
            self.product_scratch = tempfile.TemporaryDirectory(prefix="gate-product-hashes-")
        return self.product_scratch.name

    def close(self):
        if self.product_scratch is not None:
            self.product_scratch.cleanup()

    def seconds(self):
        self.check_time()
        return min(90, max(1, self.deadline - time.monotonic()))

    def product_recipes(self, package, test):
        self.check_time()
        scratch = tempfile.mkdtemp(dir=self.product_workspace())
        overlay = prepare_overlay(self.tree, scratch)
        log = os.path.join(scratch, "inputs.jsonl")
        from run import gateEnvironment
        environment = dict(os.environ, **gateEnvironment)
        environment.update(ADAMIC_GATE_PRODUCT_INPUTS=log,
                           ADAMIC_GATE_PRODUCT_INPUTS_ONLY="1", GOMAXPROCS="4")
        result = bounded_run(["go", "test", "-json", "-count=1", "-timeout", "90s", "-overlay", overlay,
                              "-run", "^" + re.escape(test) + "$", package], self.seconds(),
                             cwd=self.tree, env=environment, stdout=subprocess.PIPE,
                             stderr=subprocess.PIPE, text=True)
        if not os.path.isfile(log):
            raise ValueError("product recipe discovery failed: " + (result.stdout + result.stderr)[-2000:])
        events = [json.loads(line) for line in result.stdout.splitlines() if line.startswith('{')]
        if not any(event.get("Test") == test and event.get("Action") == "run" for event in events):
            raise ValueError("product recipe was reported before the selected unit ran")
        return read_recipes(log)

    def product_hash(self, inputs):
        self.check_time()
        if self.product_key_binary is None:
            binary = os.path.join(self.product_workspace(), "product-key")
            # Internal-package imports need a source directory inside the candidate.
            # Go ignores this dot directory in ./...; remove it before hashing files.
            source = tempfile.mkdtemp(prefix=".gate-product-key-", dir=self.tree)
            try:
                with open(os.path.join(os.path.dirname(__file__), "product_key.go")) as handle:
                    helper = handle.read().removeprefix("//go:build ignore\n\n")
                with open(os.path.join(source, "main.go"), "w") as handle:
                    handle.write(helper)
                built = bounded_run(["go", "build", "-o", binary, os.path.join(source, "main.go")], self.seconds(),
                                    cwd=self.tree, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
                if built.returncode != 0:
                    raise ValueError("candidate buildcache.Key helper failed: " + built.stderr[-2000:])
            finally:
                shutil.rmtree(source)
            self.product_key_binary = binary
        result = bounded_run([self.product_key_binary], self.seconds(),
                             input=json.dumps({"Root": self.tree, "Recipes": inputs["buildcache"]}),
                             stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        if result.returncode != 0:
            raise ValueError("candidate buildcache.Key failed: " + result.stderr[-2000:])
        reply = json.loads(result.stdout)
        keys = sorted(set(reply["keys"]))
        if not keys:
            raise ValueError("product has no recipe keys")
        # One recipe is exactly buildcache.Key. A unit fetching multiple products
        # keeps all addresses in order; the same shared function handles both.
        digest = keys[0] if len(keys) == 1 else hashlib.sha256(json.dumps(keys).encode()).hexdigest()
        return {"input_hash": digest, "input_paths": sorted(set(reply["paths"])),
                "inputs": inputs, "product_keys": keys}

    def hash(self, inputs):
        self.check_time()
        key = json.dumps(inputs, sort_keys=True)
        if key in self.cache:
            return self.cache[key]
        if "buildcache" in inputs:
            value = self.product_hash(inputs)
            self.cache[key] = value
            return value
        files = self.package_files(inputs['packages'])
        paths = sorted(set(inputs['paths'] + ['go.mod', 'go.sum']))
        listing = subprocess.run(['git', '-C', self.tree, 'ls-files', '-z', '-c', '-o',
                                  '--exclude-standard', '--'] + paths,
                                 capture_output=True, check=True, timeout=10).stdout
        for raw in listing.split(b'\0'):
            if raw:
                relative = os.fsdecode(raw)
                files['path:' + relative] = os.path.join(self.tree, relative)
        digest = hashlib.sha256()
        digest.update(b'adamic-unit-input-v1\0')
        # Declared absent paths count too: creating one changes the identity.
        digest.update(json.dumps(inputs, sort_keys=True, separators=(',', ':')).encode())
        repository_paths = set()
        for name, path in sorted(files.items()):
            self.check_time()
            digest.update(name.encode() + b'\0')
            if os.path.islink(path):
                content = b'link\0' + os.fsencode(os.readlink(path))
            elif os.path.isfile(path):
                with open(path, 'rb') as handle:
                    content = handle.read()
            else:
                content = b'missing\0'
            digest.update(hashlib.sha256(content).digest())
            relative = os.path.relpath(path, self.tree)
            if relative != '..' and not relative.startswith('../'):
                repository_paths.add(relative)
        value = {'input_hash': digest.hexdigest(), 'input_paths': sorted(repository_paths), 'inputs': inputs}
        self.cache[key] = value
        return value


def hash_units(tree, units, full=False, phase_inputs=None, product_observations=None):
    """The shared gate/integration entry point: one identity result per unit.

    Accept --list-units --with-inputs rows, recorded ledger rows, or plain unit
    strings. Path declarations stay fixed when rehashing a moved tree; Go
    closures and product recipes are evaluated on that tree. Only a gate can
    supply its just-captured product_observations to avoid replaying recipes. Never reuse a
    supplied input_hash. Missing evidence emits null and an error.
    """
    identities = InputHashes(tree)
    try:
        for unit in units:
            row = {'unit': unit} if isinstance(unit, str) else unit
            if not isinstance(row, dict):
                raise ValueError('each unit must be a string or a JSON object')
            name = row.get('unit') or ('%s %s' % (row.get('package', ''), row.get('test', ''))).strip()
            result = {'unit': name}
            try:
                identities.check_time()
                if not isinstance(name, str) or not name:
                    raise ValueError('unit needs a name or package/test')
                inputs = row.get('inputs')
                test = row.get('test') or name.partition(' ')[2]
                product = test.split('/')[0].removesuffix(' (setup)').startswith('TestProduct_')
                if product:
                    recipes = (product_observations or {}).get(name)
                    if recipes is None:
                        package = row.get('package') or name.partition(' ')[0]
                        recipes = identities.product_recipes(package, test.split('/')[0].removesuffix(' (setup)'))
                        expected = row.get('product_recipes') or (inputs or {}).get('buildcache', [])
                        names = {recipe['Name'] for recipe in recipes}
                        if any(recipe['Name'] not in names for recipe in expected):
                            raise ValueError('product recipe discovery stopped at an unavailable prerequisite; rerun this unit')
                    inputs = {'packages': [], 'paths': sorted(set(path for recipe in recipes for path in recipe.get('Files') or [])),
                              'buildcache': recipes}
                if inputs is None:
                    if row.get('package'):
                        inputs = {'packages': [row['package']], 'paths': []}
                    else:
                        package, separator, test = name.partition(' ')
                        if separator and '/' in package:
                            inputs = {'packages': [package], 'paths': []}
                        else:
                            if phase_inputs is None:
                                from run import phaseInputs
                                phase_inputs = phaseInputs
                            inputs = phase_inputs(tree, name, full)['inputs']
                if not isinstance(inputs, dict) or any(
                        not isinstance(inputs.get(field), list) or
                        any(not isinstance(value, str) for value in inputs[field])
                        for field in ('packages', 'paths')):
                    raise ValueError('inputs must declare packages and paths as string lists')
                # Keep resolved declarations even when Key cannot read a file.
                result.update(inputs=inputs, input_paths=inputs['paths'])
                result.update(identities.hash(inputs))
            except (OSError, ValueError, subprocess.SubprocessError, TimeoutError) as error:
                result.update(input_hash=None, input_hash_error=str(error))
            yield result
    finally:
        identities.close()


def main():
    parser = argparse.ArgumentParser(description='Compute the same per-unit content hashes the gate records.')
    parser.add_argument('--tree', required=True, help='checkout whose contents to hash')
    parser.add_argument('--units', required=True, help='JSONL or plain unit list; - reads stdin')
    parser.add_argument('--full', action='store_true', help='whole-gate declarations for phase names without inputs')
    args = parser.parse_args()
    def rows(handle):
        for line in handle:
            line = line.strip()
            if line:
                yield json.loads(line) if line.startswith(('{', '"', '[')) else line
    failed = False
    handle = sys.stdin if args.units == '-' else open(args.units)
    try:
        for row in hash_units(args.tree, rows(handle), args.full):
            print(json.dumps(row, sort_keys=True), flush=True)
            failed = failed or row['input_hash'] is None
    finally:
        if handle is not sys.stdin:
            handle.close()
    return 1 if failed else 0


if __name__ == '__main__':
    sys.exit(main())
