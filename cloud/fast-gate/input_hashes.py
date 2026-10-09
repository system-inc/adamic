"""Content identities for the gate ledger; checkout paths never enter the digest."""
import glob
import hashlib
import json
import os
import subprocess
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
