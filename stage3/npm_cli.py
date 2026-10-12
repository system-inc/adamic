"""The npm stage3 runs: the gate inputs' pinned package, which Loom's runner names in ADAMIC_NPM (runner/prepare.sh),
run by the tools' node. Never a bare npm from PATH: the tools directory carries none, so a PATH's npm is whatever the
machine happened to have, a program no key names."""
import os
from pathlib import Path


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
