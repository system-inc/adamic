#!/usr/bin/env python3
"""Render the checked-in report from recorded measurements."""
import gzip
import json
import statistics
from pathlib import Path

HERE = Path(__file__).resolve().parent
RESULTS = HERE / 'results'
sizes = json.loads((RESULTS / 'sizes.json').read_text())
startup = json.loads((RESULTS / 'startup.json').read_text())
versions = json.loads((RESULTS / 'versions.json').read_text())
lookup = {(row['program'], row['variant']): row for row in sizes}
lines = ['# Wasm size and startup against JavaScript', '',
    f"Measured at `{versions['revision']}` on Node {versions['node']}, WASI SDK 27 (clang 20.1.8), "
    f"Binaryen 126, esbuild 0.25.12, gzip 1.13 and {versions.get('brotli', 'brotli unknown')} CLI at `-q 11`. "
    'No compiler changes were authored and this unit changed no production flags. Final gates and measurements follow the '
    'merge of origin/main at f8013f0; premerge oracle logs are retained separately.', '',
    '`hello` is `internal/load/testdata/0.1/compile/01_hello.ts`; `dedication` is '
    '`dedication/dedication.a`. These sources contain the same dedication text and are intentional identical controls. '
    '`collections` is `internal/oracle/testdata/collections.a` (6,396 source bytes), the largest existing '
    'oracle fixture containing `new Map` in the source-byte census in `results/map-candidates.json`. '
    'It exercises dynamically built strings, map keys and values, arrays, live collection iteration and closures together. '
    '`request` is `internal/native/wasm/request.a`, selected automatically as a reactor by its exported '
    '`handleRequest(request: string): string`.', '',
    '## Flags and instruments', '',
    'Every Wasm compilation, including every runtime translation unit, uses '
    '`-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable '
    '-Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off '
    '-fno-optimize-sibling-calls --target=wasm32-wasi --sysroot=$WASI_SYSROOT '
    '-DADAMIC_TARGET_WASI=1 -mno-atomics`, followed by the optimization in the row. '
    'Linking adds `-I <runtime-cache-directory> -lm -Wl,-z,stack-size=131072`. '
    'Commands add `-mexec-model=command`. The reactor adds `-mexec-model=reactor '
    '-Wl,--export-memory` and `-Wl,--export=<name>` for each of '
    '`adamic_request, adamic_response_bytes, adamic_response_length, malloc, free, adamic_release`. '
    'No count or sanitizer flags are used. The runtime is linked with '
    '`-Xlinker --whole-archive runtime.a -Xlinker --no-whole-archive` unless the row says `no-whole`.', '',
    '| Configuration | Optimization | Additional flags or operation |',
    '|---|---|---|',
    '| O2 | `-O2` | Current driver |',
    '| Oz | `-Oz` | Replace `-O2` for generated C and runtime |',
    '| gc | `-O2` | Link: `-Wl,--gc-sections` |',
    '| sections | `-O2` | Compile and link: `-ffunction-sections -fdata-sections` |',
    '| lto | `-O2` | Compile and link: `-flto` |',
    '| no-whole | `-O2` | Remove both whole-archive switches and their `-Xlinker` prefixes |',
    '| O2-opt | `-O2` | Post-link `wasm-opt -Oz` |',
    '| Oz-opt | `-Oz` | Post-link `wasm-opt -Oz` |',
    '| Oz-strip-opt | `-Oz` | Post-link `llvm-strip`, then `wasm-opt -Oz` |', '',
    'Suffix `raw` means no further treatment; `strip` runs SDK `llvm-strip`; `opt` runs '
    'Binaryen `wasm-opt -Oz` on the raw module; `strip-opt` runs `llvm-strip` then '
    '`wasm-opt -Oz`. Columns below are `bytes / gzip -9 -n / Brotli quality 11`, independently '
    'compressing the named artifact. Brotli uses `brotli -q 11 -c`; Node independently decodes its '
    'output and checks the original bytes. Gzip and Brotli decompression round trips are checked. '
    'Complete rows, including treatments of every flag experiment and SHA-256 hashes, are in '
    '[sizes.csv](results/sizes.csv). Exact compiler argv, including SDK and cache paths, are in '
    '`results/clang-*.jsonl.gz`.', '',
    '## Bytes', '', '| Program | Configuration and treatment | Bytes / gzip9 / Brotli11 |', '|---|---|---:|']
for name in ['hello', 'dedication', 'collections', 'request']:
    for variant in ['O2-raw', 'O2-strip', 'O2-opt', 'O2-strip-opt', 'Oz-raw', 'Oz-strip', 'Oz-opt', 'Oz-strip-opt']:
        row = lookup[name, variant]
        lines.append(f"| {name} | {variant} | {row['raw']:,} / {row['gzip9']:,} / {row['brotli11']:,} |")
lines += ['', '| Program | Form | Bytes / gzip9 / Brotli11 |', '|---|---|---:|']
for name in ['hello', 'dedication', 'collections', 'request', 'shared']:
    for variant in ['source', 'js', 'js-min', 'js-runtime']:
        if (name, variant) in lookup:
            row = lookup[name, variant]
            lines.append(f"| {name} | {variant} | {row['raw']:,} / {row['gzip9']:,} / {row['brotli11']:,} |")
lines += ['', '`js` is the unchanged `adamic js` output and requires the separately listed 6,040-byte '
    'shared `oracle/adamic.mjs` runtime (provided as a local `adamic` package). `js-min` is a standalone '
    'bundle made by `esbuild --minify --bundle --platform=node --format=esm`; it includes the '
    'reachable runtime. Node built-ins are external. Source bytes are unmodified; scratch `.mts` '
    'copies let Node use its built-in type stripping. Raw source has no bundled runtime requirement '
    'for these four programs.', '',
    '| Program | O2 stripped | gc stripped | sections stripped | lto stripped | no-whole stripped (raw) |',
    '|---|---:|---:|---:|---:|---:|']
for name in ['hello', 'dedication', 'collections', 'request']:
    values = [lookup[name, variant + '-strip']['raw'] for variant in ['O2', 'gc', 'sections', 'lto', 'no-whole']]
    lines.append(f"| {name} | " + ' | '.join(f'{v:,}' for v in values) + f" ({lookup[name, 'no-whole-raw']['raw']:,}) |")
lines += ['', '## Startup', '',
    f"`nproc={startup['before']['nproc']}`, cgroup `cpu.max={startup['before']['cpu_max']}`. "
    f"Load average before: `{startup['before']['loadavg']}` at {startup['before']['time']}; "
    f"after: `{startup['after']['loadavg']}` at {startup['after']['time']}. "
    'No benchmark builds or oracle tests ran concurrently with this timing phase. The cloud host '
    'is shared and cannot be certified idle. Hyperfine was unavailable.', '',
    'All cells are milliseconds, **best / median of 30**, with raw samples in '
    '[startup.json](results/startup.json). Process runs have one untimed warmup per form, then '
    'rotate Wasm, JS and source order each round. Python `perf_counter_ns` measures spawn through '
    'process exit, with stdout/stderr redirected to `/dev/null`. Wasm uses current `O2-raw`; '
    'commands run `node --disable-warning=ExperimentalWarning oracle/wasi.mjs program.wasm`, '
    'JS runs `node --disable-warning=ExperimentalWarning program.mjs`, and source runs '
    '`node --disable-warning=ExperimentalWarning source.mts`. A reactor has no `_start`, so '
    '`reactor.mjs` instantiates and calls `_initialize` once instead of the command-only WASI runner. '
    'The process timings include host construction, file reads and program initialization.', '',
    '| Program | Wasm process | JS process | Source process |', '|---|---:|---:|---:|']
def cell(values):
    return f'{min(values):.3f} / {statistics.median(values):.3f}'
for name, data in startup['programs'].items():
    lines.append(f'| {name} | ' + ' | '.join(cell(data['process_ms'][key]) for key in ['wasm', 'js', 'source']) + ' |')
lines += ['', '| Program | WebAssembly.compile | instantiate + entry | import(JS) | import(source) |', '|---|---:|---:|---:|---:|']
for name, data in startup['programs'].items():
    lines.append(f'| {name} | ' + ' | '.join(cell(data['stages_ms'][key]) for key in ['compile', 'initialize', 'importJS', 'importSource']) + ' |')
lines += ['', 'In-process timings use Node `performance.now`, 30 rotating rounds in one Node process per '
    'program. Wasm bytes are read before timing. A fresh WASI object is constructed outside the '
    'instantiate timer, then a new instance calls `_start` or `_initialize`. JS imports use a unique '
    'query string per round to execute the entry again; its shared runtime remains cached. Console '
    'output is suppressed; WASI output goes to `/dev/null`. Repeated Wasm compilation can use V8\'s '
    'code cache. These are warm-process measurements, not 30 independent cold compilations. '
    'No forced garbage collection is performed. Entry timings include each program\'s top-level work. The request row measures initialization, '
    'not a served request.', '', '## Oracle evidence', '',
    'Every measured artifact was checked against the source on Node for exact stdout, stderr and '
    'exit. Reactor artifacts additionally agreed with the source handler on Node for ASCII, Unicode, '
    'NUL and 1 KiB inputs, releasing input and response allocations. The full WASI agreement '
    'test was run uncached with the same isolated SDK wrapper for each configuration below. '
    'Counts are Go fixture subtests, including expected lowering refusals, not a claim that every '
    'subtest executed Wasm. For checked fixtures the existing test uses the JS backend as witness.', '',
    '| Configuration | Pass | Fail | Skip | Exit |', '|---|---:|---:|---:|---:|']
for path in sorted(RESULTS.glob('oracle-*.json')):
    data = json.loads(path.read_text())
    if 'PASS' in data:
        name = path.stem[len('oracle-'):]
        lines.append(f"| {name} | {data['PASS']} | {data['FAIL']} | {data['SKIP']} | {data['exit']} |")
lines += ['', 'The existing `TestWASIOracleCatchesMutants` and `TestWASIRunnerCatchesMutants` passed: '
    'a changed output byte and exit 23 each triggered their intended disagreement. The benchmark '
    'also changed `Kenneth Lane Thompson` to `Kenneth Lane Thompson!` in emitted JS, appended '
    '`process.exitCode = 23`, and appended a stderr write, independently. Each compiled/executed '
    'and was rejected by the exact comparison. A real Go no-tests-selected green run was rejected '
    'by the nonzero subtest guard. Compressing input plus `!` was rejected by both decompressed-byte '
    'equality checks; adding `!` to the observed reactor response was rejected by the source response '
    'comparison. These are recorded in `results/mutants.json`. No mutant '
    'modified compiler or runtime sources.', '', '## Reproduce', '', '```sh',
    'bash cloud/setup.sh --wasi-sdk > /tmp/wasm-size-setup.log 2>&1',
    'source /workspace/adamic-tools/env.sh',
    '# Install tools in the workspace, not the repository.',
    'mkdir -p /workspace/wasm-size-tools',
    'curl -fsSL https://github.com/WebAssembly/binaryen/releases/download/version_126/binaryen-version_126-x86_64-linux.tar.gz -o /workspace/wasm-size-tools/binaryen.tar.gz',
    'tar -xzf /workspace/wasm-size-tools/binaryen.tar.gz -C /workspace/wasm-size-tools',
    'npm install --prefix /workspace/wasm-size-tools esbuild@0.25.12 > /tmp/wasm-size-npm.log 2>&1',
    'mkdir -p /workspace/wasm-size-tools/brotli-1.1.0 /workspace/wasm-size-tools/brotli-build',
    'curl -fsSL https://github.com/google/brotli/archive/refs/tags/v1.1.0.tar.gz -o /workspace/wasm-size-tools/brotli.tar.gz',
    'tar -xzf /workspace/wasm-size-tools/brotli.tar.gz -C /workspace/wasm-size-tools/brotli-1.1.0 --strip-components=1',
    'cc -O2 -DBROTLI_HAVE_LOG2=1 -DOS_LINUX -I /workspace/wasm-size-tools/brotli-1.1.0/c/include /workspace/wasm-size-tools/brotli-1.1.0/c/common/*.c /workspace/wasm-size-tools/brotli-1.1.0/c/dec/*.c /workspace/wasm-size-tools/brotli-1.1.0/c/enc/*.c /workspace/wasm-size-tools/brotli-1.1.0/c/tools/brotli.c -lm -o /workspace/wasm-size-tools/brotli-build/brotli > /tmp/wasm-size-brotli-cc.log 2>&1',
    'bash bench/wasm/run.sh measure > /tmp/wasm-size-measure.log 2>&1',
    'bash bench/wasm/run.sh strip-opt > /tmp/wasm-size-strip-opt.log 2>&1',
    'for configuration in O2 Oz O2-opt Oz-opt Oz-strip-opt no-whole lto; do',
    '  bash bench/wasm/run.sh oracle --configuration "$configuration" > "/tmp/wasm-size-oracle-$configuration-status.log" 2>&1',
    'done',
    '# Stop building/testing before this timing phase.',
    'bash bench/wasm/run.sh startup > /tmp/wasm-size-startup.log 2>&1',
    'ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run "^TestWASI(OracleCatchesMutants|RunnerCatchesMutants)$" -count=1 -v -timeout 10m > bench/wasm/results/oracle-mutants.log 2>&1',
    'python3 bench/wasm/render.py',
    'bash bench/wasm/run.sh archive', '```', '',
    'The oracle mode runs `ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle '
    "-run '^TestWASIAgreesWithNode$' -count=1 -v -timeout 30m` into a log file. "
    'An isolated SDK directory symlinks the real sysroot and llvm-ar and supplies `clang.py` as '
    'its compiler. The wrapper replaces `-O2`, appends only the named flags, or postprocesses '
    'successful links. Its `--version` includes the configuration so the unchanged runtime cache '
    'cannot reuse an archive built with different hidden flags. Compiler argv and all build '
    'commands are recorded. It never edits `internal/native`.', '',
    'Setup printed Go/clang/Node ready at 0s, WASI SDK and submodules ready at 9s, '
    'build cache warm and done at 117s, 5 processors and four CPUs of quota. CMake was unavailable '
    '(`cmake: command not found`), so Brotli was built directly with the recorded C command.', '',
    '## Reading the numbers', '',
    'Stripping removes far more bytes than switching optimization levels. Explicit section GC '
    'and function/data sections do not reduce stripped sizes here. Function/data sections can change '
    'raw metadata bytes; the complete raw and compressed rows remain in the CSV. Dropping whole-archive '
    'changes raw metadata size but has no stripped-size benefit for these programs; retain '
    'the current constructor-preserving link policy. The reactor is 16 stripped bytes larger '
    'without whole-archive. LTO increases stripped sizes for all four measured programs; '
    'this sample gives no size reason to enable it.', '',
    'For a size-oriented build, the observed candidate is `-Oz` followed by `llvm-strip` and '
    '`wasm-opt -Oz`, provided the configuration\'s full oracle is green. This is a recommendation '
    'supported by the recorded agreement run, not a production flag change. Compare compressed '
    'rows as well as raw rows: smaller uncompressed output does not guarantee smaller gzip or '
    'Brotli output.', '',
    'In these runs, Wasm process medians were 29.4 to 31.3 ms, JS medians were 27.6 to 30.0 ms, '
    'and source medians were 52.3 to 60.6 ms. JS processes finished sooner than Wasm processes; '
    'warm in-process compile/import timings measure a different boundary and can benefit from '
    'engine caches.', '',
    'The standalone minified JS bundles are much smaller than these Wasm modules. Process '
    'startup includes Node and differs from warm compile/import timings. These data do not '
    'measure browser downloads, Worker cold starts, HTTP throughput, deployed isolates, sustained '
    'execution speed, memory use or new correctness properties. No full repository gate was '
    'run for this benchmark-only change. No size optimization was landed in the compiler.', '']
(HERE / 'README.md').write_text('\n'.join(lines))
