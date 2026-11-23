"""Generate the TypeScript upstream ledger in adaptation apply order."""

import argparse
from pathlib import Path
import sys

sys.dont_write_bytecode = True
HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE.parent))
from apply import adaptation_directories


def render():
    blocks = [(HERE / 'header.md').read_bytes()]
    for directory in adaptation_directories():
        entry = (directory / 'LEDGER.md').read_bytes()
        if not entry.startswith(b'none:'):
            blocks.append(entry)
    return b'\n\n'.join(block.rstrip(b'\n') for block in blocks) + b'\n'


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('output', nargs='?', type=Path, help='output path (default: stdout)')
    args = parser.parse_args()
    if args.output is None:
        sys.stdout.buffer.write(render())
    else:
        args.output.write_bytes(render())


if __name__ == '__main__':
    main()
