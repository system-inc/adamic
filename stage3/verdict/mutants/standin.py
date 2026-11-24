"""Target exactly one real project, or the explicitly synthetic emit fixture."""
import os
from pathlib import Path
import subprocess
import sys

mode, *args = sys.argv[1:]
command = ['node', os.environ['STAGE3_VERDICT_NODE_TSC'], *args]
synthetic_modes = ('declaration-byte', 'buildinfo-byte', 'no-emit-incremental-js',
                   'no-emit-incremental-missing')
synthetic = (mode in synthetic_modes
             and os.environ.get('STAGE3_VERDICT_SYNTHETIC_FIXTURE') == '1'
             and Path('emit.ts').is_file())
real_source = ('expandoFunctionExpressionsWithDynamicNames2.ts' if mode == 'no-emit-write'
               else 'varianceCantBeStrictWhileStructureIsnt.ts')
target = ('012_expandoFunctionExpressionsWithDynamicNames2' if mode == 'no-emit-write'
          else '001_varianceCantBeStrictWhileStructureIsnt')
real_target = (mode not in synthetic_modes
               and os.environ.get('STAGE3_VERDICT_SUITE') == 'acceptance'
               and os.environ.get('STAGE3_VERDICT_CASE') == target
               and os.environ.get('STAGE3_VERDICT_SIDE') == 'candidate')
if not synthetic and not real_target:
    os.execvp(command[0], command)
if mode == 'no-emit-write' and '--noEmit' not in args:
    os.execvp(command[0], command)
if mode == 'timeout':
    os.execlp('sleep', 'sleep', '195')
result = subprocess.run(command)
javascript = Path('out/emit.js' if synthetic else Path(real_source).with_suffix('.js'))
if mode == 'no-emit-incremental-js':
    Path('unexpected.js').write_bytes(b'// forbidden under noEmit\n')
elif mode == 'no-emit-incremental-missing':
    Path('tsconfig.tsbuildinfo').unlink()
elif mode in ('extra-file', 'no-emit-write'):
    Path('unexpected.js').write_bytes(b'// unexpected\n')
elif mode == 'missing-file':
    javascript.unlink()
else:
    products = {'declaration-byte': Path('out/emit.d.ts'),
                'buildinfo-byte': Path('tsconfig.tsbuildinfo' if '--noEmit' in args else 'out/tsconfig.tsbuildinfo')}
    path = products.get(mode, javascript)
    data = path.read_bytes()
    path.write_bytes(bytes([data[0] ^ 1]) + data[1:])
raise SystemExit(result.returncode)
