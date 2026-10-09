#!/usr/bin/env python3
"""Content identities for the gate ledger; checkout paths never enter the digest."""
import argparse
import fnmatch
import glob
import hashlib
import json
import os
import re
import subprocess
import sys
import time

BOX_PATHS = ('cloud/fast-gate.sh cloud/fast-gate cloud/darwin-leg.sh '
             'cloud/fast-gate-classify.sh cloud/idle-preempt.sh internal/skipcensus go.mod go.sum').split()


# The Mac-only tests under cloud/fast-gate that boxTools() leaves out (969b9dc0): a test edit doesn't change the box.
macOnlyTest = re.compile(rb'\tcloud/fast-gate/[^/]*_test\.py$')


def tools_fingerprint(tree):
    # Byte-for-byte boxTools() in cloud/fast-gate-watch.sh: recursive ls-tree of boxSide, the Mac-only fast-gate tests
    # dropped, then SHA-1. UnitInputHashes runs the watcher's own function against this one.
    listing = subprocess.run(['git', '-C', tree, 'ls-tree', '-r', 'HEAD', '--'] + BOX_PATHS,
                             capture_output=True, check=True, timeout=10).stdout
    kept = b''.join(line for line in listing.splitlines(keepends=True) if not macOnlyTest.search(line.rstrip(b'\n')))
    return hashlib.sha1(kept).hexdigest()


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

    def hash(self, inputs):
        self.check_time()
        key = json.dumps(inputs, sort_keys=True)
        if key in self.cache:
            return self.cache[key]
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


def reads_lines(path):
    """(package, glob) for each reads line of an executors.txt."""
    if not os.path.exists(path):
        return []
    with open(path) as handle:
        return [tuple(line.split()[1:3]) for line in handle if line.split()[:1] == ['reads']]


def test_unit_inputs(identities, package, tools):
    """What a test unit reads beyond its Go closure (#pgnnb67): its package's own testdata, read by path and never
    compiled; each path a reads line names for its package (executors.txt, the tools' and then the tree's); and the
    compiler packages compiler-dependencies.json declares its tests run. Without them an edit to
    internal/oracle/testdata left oracle's, flow's and lower's hashes where they were."""
    if tools is None:
        raise ValueError('a test unit needs the gate tools for its reads lines (--tools)')
    identities.inventory()
    row = identities.packages.get(package)
    if row is None or not row.get('Dir'):
        raise ValueError('incomplete go list closure: ' + package)
    directories = {os.path.relpath(os.path.realpath(value['Dir']), identities.tree): name
                   for name, value in identities.packages.items() if value.get('Dir')}
    directory = os.path.relpath(os.path.realpath(row['Dir']), identities.tree)
    listing = subprocess.run(['git', '-C', identities.tree, 'ls-files', '-z', '-c', '-o', '--exclude-standard'],
                             capture_output=True, check=True, timeout=10).stdout
    files = [os.fsdecode(raw) for raw in listing.split(b'\0') if raw]
    paths = set()
    for path in files:
        parts = path.split('/')
        if 'testdata' not in parts:
            continue
        owner = '/'.join(parts[:parts.index('testdata')]) or '.'
        while owner not in directories and owner not in ('', '.'):
            owner = os.path.dirname(owner)
        if owner == directory:
            paths.add('/'.join(parts[:parts.index('testdata') + 1]))
    roots = []
    for root in dict.fromkeys(os.path.realpath(value) for value in (tools, identities.tree)):
        roots.append(root)
        for reader, pattern in reads_lines(os.path.join(root, 'cloud/fast-gate/executors.txt')):
            if reader != directory:
                continue
            matched = [path for path in files if fnmatch.fnmatchcase(path, pattern)]
            # A declared path that isn't there yet still counts: creating it changes the identity.
            paths.update(matched or ([] if any(mark in pattern for mark in '*?[') else [pattern]))
    declared = []
    for root in roots:
        name = os.path.join(root, 'cloud/fast-gate/compiler-dependencies.json')
        if os.path.exists(name):
            with open(name) as handle:
                declared = json.load(handle).get('packages', {}).get(directory, declared)
    packages = [package] + sorted(directories[value] for value in declared if value in directories and directories[value] != package)
    return {'packages': packages, 'paths': sorted(paths)}


def hash_units(tree, units, full=False, phase_inputs=None, tools=None):
    """The shared gate/integration entry point: one identity result per unit.

    Accept --list-units --with-inputs rows, recorded ledger rows, or plain unit
    strings. Supplied declarations stay fixed when rehashing a moved tree; Go
    closures and content are always recomputed on that tree. Never reuse a
    supplied input_hash. Missing evidence emits null and an error.
    """
    identities = InputHashes(tree)
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
            if inputs is None:
                if row.get('package'):
                    inputs = test_unit_inputs(identities, row['package'], tools)
                else:
                    package, separator, test = name.partition(' ')
                    if separator and '/' in package:
                        inputs = test_unit_inputs(identities, package, tools)
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
            result.update(identities.hash(inputs))
        except (OSError, ValueError, subprocess.SubprocessError, TimeoutError) as error:
            result.update(input_hash=None, input_hash_error=str(error))
        yield result


def main():
    parser = argparse.ArgumentParser(description='Compute the same per-unit content hashes the gate records.')
    parser.add_argument('--tree', required=True, help='checkout whose contents to hash')
    parser.add_argument('--units', required=True, help='JSONL or plain unit list; - reads stdin')
    parser.add_argument('--full', action='store_true', help='whole-gate declarations for phase names without inputs')
    parser.add_argument('--tools', help='gate tools checkout whose executors.txt and compiler map declare test units\' reads')
    args = parser.parse_args()
    def rows(handle):
        for line in handle:
            line = line.strip()
            if line:
                yield json.loads(line) if line.startswith(('{', '"', '[')) else line
    failed = False
    handle = sys.stdin if args.units == '-' else open(args.units)
    try:
        for row in hash_units(args.tree, rows(handle), args.full, tools=args.tools):
            print(json.dumps(row, sort_keys=True), flush=True)
            failed = failed or row['input_hash'] is None
    finally:
        if handle is not sys.stdin:
            handle.close()
    return 1 if failed else 0


if __name__ == '__main__':
    sys.exit(main())
