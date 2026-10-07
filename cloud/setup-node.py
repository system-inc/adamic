#!/usr/bin/env python3
"""Install only the published, checksum-pinned Node release, independent of PATH."""
import hashlib
import json
import os
from pathlib import Path
import platform
import subprocess
import sys
import tarfile
import tempfile
import time

SOURCE = Path(__file__).resolve().parent
PIN = SOURCE / 'node-pin.json'


def digest(data):
    return hashlib.sha256(data).hexdigest()


def installation_key(version, checksum, system, architecture, pin, helper):
    return digest(json.dumps(dict(version=version, checksum=checksum, system=system,
                                  architecture=architecture, pin=pin, helper=helper), sort_keys=True).encode())


def asset(system=None, machine=None):
    system = system or platform.system()
    machine = machine or platform.machine()
    systems = {'Linux': 'linux', 'Darwin': 'darwin'}
    architectures = {'x86_64': 'x64', 'aarch64': 'arm64', 'arm64': 'arm64'}
    if system not in systems or machine not in architectures:
        raise ValueError(f'unsupported Node platform: {system}/{machine}')
    return systems[system], architectures[machine]


def verify_archive(archive, expected):
    if digest(archive.read_bytes()) != expected:
        raise ValueError(f'SHA-256 mismatch for {archive.name}')


def download(url, destination):
    subprocess.run(['curl', '-fsSL', '--connect-timeout', '30', '--max-time', '600',
                    url, '-o', str(destination)], check=True, timeout=610)


def prepare(tools):
    tools = Path(tools)
    if tools.resolve().is_relative_to('/root'):
        raise ValueError('Node tools must be outside /root')
    pin_bytes = PIN.read_bytes()
    pin = json.loads(pin_bytes)
    version = pin['version']
    system, architecture = asset()
    name = f'node-{version}-{system}-{architecture}.tar.gz'
    checksum = pin['sha256'][name]
    key = installation_key(version, checksum, system, architecture, digest(pin_bytes),
                           digest(Path(__file__).read_bytes()))
    destination = tools / 'bin/node'
    stamp = tools / 'node.stamp'
    if os.environ.get('ADAMIC_GATE_UNCACHED') != '1':
        try:
            saved = json.loads(stamp.read_text())
            if (saved['key'] == key and not destination.is_symlink()
                    and destination.stat().st_mode & 0o777 == 0o755
                    and saved['binary'] == digest(destination.read_bytes())):
                return f'node {version} skipped (validated version, checksum and installed bytes)'
        except (OSError, ValueError, KeyError):
            pass
    (tools / 'bin').mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix='node-install-', dir=tools) as temporary:
        scratch = Path(temporary)
        archive = scratch / name
        sums = scratch / 'SHASUMS256.txt'
        base = f'https://nodejs.org/dist/{version}'
        download(base + '/SHASUMS256.txt', sums)
        published = [line.split()[0] for line in sums.read_text().splitlines()
                     if len(line.split()) == 2 and line.split()[1] == name]
        if published != [checksum]:
            raise ValueError(f'published SHA-256 differs from pin for {name}')
        download(base + '/' + name, archive)
        verify_archive(archive, checksum)
        # Read one regular member without extracting untrusted paths or links.
        with tarfile.open(archive, 'r:gz') as release:
            member = release.getmember(name[:-7] + '/bin/node')
            if not member.isfile():
                raise ValueError('Node archive member must be a regular file')
            binary = scratch / 'node'
            with release.extractfile(member) as source, binary.open('wb') as target:
                while chunk := source.read(1024 * 1024):
                    target.write(chunk)
        binary.chmod(0o755)
        actual = subprocess.run([str(binary), '--version'], check=True, capture_output=True,
                                text=True, timeout=30).stdout.strip()
        if actual != version:
            raise ValueError(f'node: got {actual}, want {version}')
        metadata = scratch / 'stamp'
        metadata.write_text(json.dumps(dict(key=key, binary=digest(binary.read_bytes())), sort_keys=True))
        os.replace(binary, destination)
        os.replace(metadata, stamp)
    return f'node {version} installed (published SHA-256 verified)'


if __name__ == '__main__':
    started = time.monotonic()
    print(f'setup: {prepare(sys.argv[1])}; step-duration={time.monotonic() - started:.3f}s')
