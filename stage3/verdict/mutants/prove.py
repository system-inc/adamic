"""Exercise the real verdict comparison on live Node and six named mutants."""
import json
import os
from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parent
sys.path.insert(0, str(ROOT.parent))
from run import compare_case, materialize_command

# These directives are outside the pinned census. Real-case mutants use run.sh.
CASES = {
    'declaration-byte': ('synthetic-declaration', 'out/emit.d.ts'),
    'buildinfo-byte': ('synthetic-incremental', 'out/tsconfig.tsbuildinfo'),
}

def main():
    output = Path(sys.argv[1]).resolve()
    output.mkdir()
    os.environ['STAGE3_VERDICT_SYNTHETIC_FIXTURE'] = '1'
    reference = ['node', str(Path(os.environ['STAGE3_VERDICT_NODE_TSC']).resolve())]
    for mode, (case, path) in CASES.items():
        raw = (b'// @declaration: true\n// @sourceMap: true\n// @outDir: out\n'
               b'// @incremental: true\nexport const n = 1;\n')
        source, options = materialize_command(raw)
        folder = output / mode
        control = compare_case(ROOT.parent / 'standins/node.sh', reference, folder / 'control',
                               {'emit.ts': source}, options)
        if control:
            raise RuntimeError(f'{case}: Node control failed: {control}')
        differences = compare_case(ROOT / (mode + '.sh'), reference, folder / 'mutant',
                                   {'emit.ts': source}, options)
        if set(differences) != {'files'} or differences['files'].get(path, {}).get('kind') != 'bytes':
            raise RuntimeError(f'{case}: wrong failure: {differences}')
        print(json.dumps({'fixture': 'synthetic', 'case': case, 'standin': mode, 'exit': 1, 'differences': differences}),
              flush=True)

if __name__ == '__main__':
    main()
