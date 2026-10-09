from pathlib import Path
import json, os, subprocess, time

root = Path.cwd()
evidence = root / 'review/compiler/lowering-chain-reds'

def overlay(name, path, change, packages, pattern, required):
    directory = evidence / 'mutants' / name
    directory.mkdir(parents=True, exist_ok=True)
    source = root / path
    original = source.read_text()
    changed = change(original)
    assert changed != original, name
    target = directory / (source.name + '.txt')
    target.write_text(changed)
    manifest = directory / 'overlay.json'
    manifest.write_text(json.dumps({'Replace': {str(source): str(target)}}))
    env = dict(os.environ, GOFLAGS='-overlay=' + str(manifest))
    command = ['go', 'test', *packages, '-run', pattern, '-count=1', '-v', '-timeout', '90s']
    started = time.monotonic()
    # go/types scans source files on disk, so reader mutants must be visible there.
    reader = name in {'sort-direct-reader', 'static-unapproved-reader', 'css-part-omission'}
    if reader:
        source.write_text(changed)
        env.pop('GOFLAGS', None)
    try:
        with (directory / 'test.log').open('w') as log:
            result = subprocess.run(command, env=env, stdout=log, stderr=subprocess.STDOUT, timeout=120)
    finally:
        if reader:
            source.write_text(original)
    output = (directory / 'test.log').read_text()
    assert result.returncode == 1 and all(text in output for text in required), (name, result.returncode, output[-2000:])
    print(f'{name}: caught, exit 1, {time.monotonic()-started:.2f}s', flush=True)

overlay('push-old-stop', 'internal/lower/library_array.go',
 lambda s: s.replace('if name == "push" || name == "unshift" {', 'if name == "push" { return nil, true, l.notYet(node, "push with other than one value") }\n\tif name == "push" || name == "unshift" {', 1),
 ['./stage1/typescript/parser', './stage1/cohere/cssstrings', './stage1/cohere/json'],
 '^(TestPushSpreadGap|TestMultiPushGap|TestDocumentedStageZeroGaps)$',
 ['--- FAIL: TestPushSpreadGap', '--- FAIL: TestMultiPushGap', '--- FAIL: TestDocumentedStageZeroGaps'])

guard = '''        if name == "get" && l.optionalMapStructuralReceiver(source) {
            return nil, l.notYet(node, "an optional Map get through a structural receiver")
        }'''
# Read gofmt's actual block, retaining the nearby explanation.
chain = (root / 'internal/lower/optional_chain.go').read_text()
start = chain.index('\t\tif name == "get" && l.optionalMapStructuralReceiver(source) {')
end = chain.index('\n\t\tkey, value, err := l.mapTypes(source)', start)
block = chain[start:end]
overlay('map-boundary-bypass', 'internal/lower/optional_chain.go', lambda s: s.replace(block, '', 1),
 ['./internal/lower'], '^TestOptionalIndexingMapShapeRefused$', ['got <nil>, want structural Map receiver boundary'])

overlay('sort-direct-reader', 'internal/lower/exceptions.go',
 lambda s: s.replace('l.result.ClosureMayThrow(node)', 'l.result.Functions[node.Comparator].MayThrow', 1),
 ['./internal/ir'], '^TestCallTargetReaders$', ['unapproved call-target read internal/lower/exceptions.go:throwsOutReadiness:ArraySort.Comparator'])
overlay('static-unapproved-reader', 'internal/lower/class_static_guard_test.go',
 lambda s: s.replace('func TestClassStaticInitializerCallIsEmitted(', 'func TestClassStaticInitializerCallIsEmittedMutant(', 1),
 ['./internal/ir'], '^TestCallTargetReaders$', ['unapproved call-target read internal/lower/class_static_guard_test.go:TestClassStaticInitializerCallIsEmittedMutant:Call.Function'])
overlay('stale-allowlist', 'internal/ir/call_targets_guard_test.go',
 lambda s: s.replace('var targetReaders = map[string]targetReader{', 'var targetReaders = map[string]targetReader{\n "internal/lower/exceptions.go:throwsOut:ArraySort.Comparator": {"compiler", "retired reader"},', 1),
 ['./internal/ir'], '^TestCallTargetReaders$', ['stale call-target allowlist entry internal/lower/exceptions.go:throwsOut:ArraySort.Comparator'])

overlay('css-part-omission', 'stage1/cohere/cssstrings/strings.ts',
 lambda s: s.replace('parts.push(value.slice(start, index), printString(value.slice(index, end), singleQuote));', 'parts.push(printString(value.slice(index, end), singleQuote));'),
 ['./stage1/cohere/cssstrings'], '^TestMultiPushPortMatchesGo$', ['first byte difference', '--- FAIL: TestMultiPushPortMatchesGo'])
