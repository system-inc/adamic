#!/usr/bin/env python3
"""Extend the guarded scratch overlay for a checker-clean entry's resolved reach."""
import json
from pathlib import Path
import sys


def prepare(source, output):
    source, output = Path(source).resolve(), Path(output).resolve()
    output.mkdir(parents=True, exist_ok=True)
    metadata = json.loads((source / 'overlay.json').read_text())
    replacements = dict(metadata['Replace'])
    def target(suffix):
        matches = [name for name in replacements if name.endswith('/' + suffix)]
        if len(matches) != 1:
            raise ValueError(f'entry overlay: expected one {suffix}, found {len(matches)}')
        return matches[0]
    hook = target('internal/load/latent_hook.go')
    original = Path(replacements[hook]).read_text()
    if 'func (p *Program) LatentEntryReach(' in original:
        raise ValueError('entry overlay: duplicate Program.LatentEntryReach')
    # GetSourceFiles is the checker's resolved graph, not every file in a directory. Its order puts
    # dependencies before the entry, and lowering reads Files()[0] as the entry: module identities
    # are routes from it (internal/lower/source_identity.go). So the entry stays first, and the rest
    # follow in the checker's order. With the entry anywhere else, every module the first file
    # doesn't import panics "source module has no import-declaration route" (2,279 sites on a0ba9dd5).
    text = original + '''
// Measurement only: ordinary Load and Lower remain disabled by this overlay.
func (p *Program) LatentEntryReach() {
    entry := p.files[0]
    p.files = []*ast.SourceFile{entry}
    for _, source := range p.compiler.GetSourceFiles() {
        if source != entry && !source.IsDeclarationFile && !IsLibrary(source) {
            p.files = append(p.files, source)
        }
    }
}
'''
    destination = output / 'entry_load_hook.go'
    destination.write_text(text)
    replacements[hook] = str(destination)
    driver = target('stage3/census/latent/tool/main.go')
    destination = output / 'entry_main.go'
    destination.write_text(Path(__file__).with_name('entry.go.txt').read_text())
    replacements[driver] = str(destination)
    (output / 'overlay.json').write_text(json.dumps({'Replace': replacements}, indent=2) + '\n')


if __name__ == '__main__':
    prepare(*sys.argv[1:3])
