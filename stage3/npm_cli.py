"""The npm stage3 runs: the gate inputs' pinned package, which Loom's runner names in ADAMIC_NPM (runner/prepare.sh),
run by the tools' node. Never a bare npm from PATH: the tools directory carries none, so a PATH's npm is whatever the
machine happened to have, a program no key names."""
import os
from pathlib import Path
import shlex
import tempfile

# The directory holding environment()'s npm, made once a process.
_directory = None


def npm(*arguments):
    """The command that runs npm with arguments, refused when ADAMIC_NPM names no npm package."""
    package = os.environ.get('ADAMIC_NPM')
    if not package:
        raise RuntimeError("ADAMIC_NPM is unset: stage3 runs the gate inputs' npm "
                           '(<ADAMIC_NPM>/package/bin/npm-cli.js), never one from PATH')
    cli = Path(package) / 'package' / 'bin' / 'npm-cli.js'
    if not cli.is_file():
        raise RuntimeError(f'ADAMIC_NPM names no npm package: {cli} is missing')
    return ['node', str(cli), *arguments]


def environment(base=None):
    """base (this process's environment by default) with an npm first on PATH that runs the same package, for the
    package scripts that call npm by name: TypeScript's build is npm run build:compiler && npm run build:tests, which
    would otherwise find no npm, or the machine's."""
    global _directory
    command = npm()
    if _directory is None:
        directory = Path(tempfile.mkdtemp(prefix='stage3-npm-'))
        script = directory / 'npm'
        script.write_text('#!/bin/sh\nexec ' + ' '.join(shlex.quote(part) for part in command) + ' "$@"\n')
        script.chmod(0o755)
        _directory = directory
    environment = dict(os.environ if base is None else base)
    environment['PATH'] = str(_directory) + os.pathsep + environment.get('PATH', '')
    return environment
