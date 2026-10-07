#!/usr/bin/env python3
"""Regenerate owned source assets after checking the checker library pin."""
import hashlib
import json
import pathlib
UNIT = pathlib.Path(__file__).resolve().parent
ROOT = UNIT.parents[3]
pins = json.loads((UNIT/'library-pins.json').read_text())
LIB = ROOT/'cohere/TypeScript/tsc/internal/bundled/libs'
rows = ['// Apache-2.0 TypeScript library source data from the pinned checker.', '// Regenerate with the checker pin. See LIBRARY_NOTICE.md and library-pins.json.', 'const sources = new Map<string, string>([']
for name, digest in pins['sha256'].items():
    raw = (LIB/name).read_bytes()
    assert hashlib.sha256(raw).hexdigest() == digest, name+' changed; update the pin deliberately'
    pieces = raw.decode().split('${')
    expression = ' + "$" + "{" + '.join(json.dumps(piece, ensure_ascii=False) for piece in pieces)
    rows.append('    ['+json.dumps('bundled:///libs/'+name)+', '+expression+'],')
rows.extend([']);', 'export function librarySource(path: string): string | undefined {', '    return sources.get(path);', '}'])
(UNIT/'libraries.a').write_text('\n'.join(rows)+'\n')
print('Regenerated 114 pinned assets; run owned source formatting afterward.')
