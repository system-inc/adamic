#!/usr/bin/env python3
"""The census's own-file table, attributed the way stage3/meter/report.py attributes it.

Each diagnostic belongs to the file of its primary location, deduplicated across the roots
that observed it; elaboration lines are not locations. Prints one line per src/compiler
file, its own-file diagnostic count, then the located diagnostics themselves.

    python3 own-file.py <adapted tree> <census.jsonl> <output.json>
"""
import json
import re
import sys
from pathlib import Path


def own_file(tree, census):
    compiler = (tree / 'src/compiler').resolve()
    expected = {str(path.resolve()) for path in compiler.rglob('*') if path.is_file() and path.suffix in ('.ts', '.a')}
    located = {name: set() for name in expected}
    external = set()
    unlocated = set()
    for line in census.read_text().splitlines():
        for diagnostic in json.loads(line).get('diagnostics', []):
            match = re.match(r'^(.+?):[0-9]+:[0-9]+: error TS[0-9]+:', diagnostic)
            if match is None:
                unlocated.add(diagnostic)
                continue
            name = str(Path(match[1]).resolve())
            (located[name] if name in located else external).add(diagnostic)
    relative = lambda name: str(Path(name).relative_to(tree.resolve()))
    # Diagnostics carry the tree's absolute path; keep src/ on so two trees compare.
    strip = lambda diagnostic: diagnostic[diagnostic.rindex('/src/') + 1:] if '/src/' in diagnostic else diagnostic
    return {
        'files': {relative(name): sorted(strip(d).split('\n')[0] for d in located[name]) for name in sorted(located)},
        'external': len(external),
        'unlocated': len(unlocated),
    }


def main():
    tree, census, output = Path(sys.argv[1]), Path(sys.argv[2]), Path(sys.argv[3])
    table = own_file(tree, census)
    output.write_text(json.dumps(table, indent=2) + '\n')
    files = table['files']
    for name, diagnostics in files.items():
        print(f'{len(diagnostics):5d}  {name}')
    print(f'own-file diagnostics {sum(map(len, files.values()))} in {sum(1 for d in files.values() if d)} of {len(files)} files; '
          f'external {table["external"]}, unlocated {table["unlocated"]}')


if __name__ == '__main__':
    main()
