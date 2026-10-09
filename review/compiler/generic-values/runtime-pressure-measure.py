from pathlib import Path
import subprocess
root = Path('/workspace/generic-values-proof-tmp')
for name in ['base', 'head']:
 header = Path('internal/native/runtime/adamic.h').read_bytes() if name == 'head' else subprocess.check_output(['git', 'show', '50654a40:internal/native/runtime/adamic.h'])
 jsoncode = Path('internal/native/runtime/library_language.c').read_bytes() if name == 'head' else subprocess.check_output(['git', 'show', '50654a40:internal/native/runtime/library_language.c'])
 mapcode = Path('internal/native/runtime/map.c').read_bytes() if name == 'head' else subprocess.check_output(['git', 'show', '50654a40:internal/native/runtime/map.c'])
 libraries = []
 for path in (Path.home() / '.cache/adamic/runtime').iterdir():
  if not (path / 'runtime.a').exists() or (path / 'adamic.h').read_bytes() != header or (path / 'map.c').read_bytes() != mapcode or (path / 'library_language.c').read_bytes() != jsoncode: continue
  symbols = subprocess.check_output(['nm', '-g', str(path / 'features.o')])
  if b'adamic_runtime_features\n' not in symbols or b'__asan' in symbols: continue
  # A sanitized archive has allocator instrumentation; release does not.
  heap_symbols = subprocess.check_output(['nm', '-g', str(path / 'heap.o')])
  if b'__asan' in heap_symbols or b'adamic_count_allocation' in heap_symbols: continue
  libraries.append(path)
 assert len(libraries) == 1, libraries
 lib = libraries[0]
 source = root / ('runtime-base-pressure.c' if name == 'base' else 'runtime-pressure.c')
 binary = root / ('runtime-' + name + '-measured-pressure')
 command = ['clang', '-O2', '-std=c11', '-ffp-contract=off', '-fno-optimize-sibling-calls', '-I', str(lib), str(source), 'review/compiler/generic-values/runtime-pressure-probe.c', '-Wl,--whole-archive', str(lib / 'runtime.a'), '-Wl,--no-whole-archive', '-Wl,--wrap=adamic_allocate', '-Wl,--wrap=adamic_weak_forget', '-lm', '-o', str(binary)]
 subprocess.run(command, check=True, timeout=90)
 result = subprocess.run([str(binary)], capture_output=True, check=True, timeout=20)
 assert result.stdout == b'199990000\n', result
 print(name, ' '.join(command)); print(result.stdout.decode() + result.stderr.decode())
