#!/usr/bin/env python3
"""Private typed-runner comparison; the shared syntax context has no checker lease."""
import argparse
import gzip
import json
from pathlib import Path
import re
import subprocess
import statistics
import time

parser = argparse.ArgumentParser()
parser.add_argument('--artifacts', type=Path, required=True)
parser.add_argument('--compiler', type=Path, required=True)
parser.add_argument('--checker', type=Path, required=True)
parser.add_argument('--sanitized-checker', type=Path, required=True)
args = parser.parse_args()
root = Path(__file__).resolve().parents[4]
own = Path(__file__).resolve().parent
out = args.artifacts.resolve()
out.mkdir(parents=True, exist_ok=True)
sequence = 0

def run(label, command, cwd=root, expected=0):
    global sequence
    sequence += 1
    stem = out / f'{sequence:03d}-{label}'
    started = time.monotonic()
    with stem.with_suffix('.stdout').open('wb') as stdout, stem.with_suffix('.stderr').open('wb') as stderr:
        result = subprocess.run(list(map(str, command)), cwd=cwd, stdout=stdout, stderr=stderr, timeout=600)
    assert result.returncode == expected, (label, result.returncode, expected)
    return stem.with_suffix('.stdout').read_bytes(), stem.with_suffix('.stderr').read_bytes(), time.monotonic() - started

virtual = root / 'cohere/wave20_jsx_constructed_oracle.go'
overlay = out / 'overlay.json'
overlay.write_text(json.dumps({'Replace': {str(virtual): str(own / 'testdata/oracle.go')}}))
oracle = out / 'oracle'
run('oracle-build', ['go', 'build', '-overlay', overlay, '-o', oracle, virtual], root / 'cohere')
entry = own / 'suite.a'
native = out / 'suite'
san = out / 'suite-asan'
run('native-build', [args.compiler, 'build', entry, '-o', native, '--tsgo', args.checker])
run('asan-build', [args.compiler, 'build', entry, '-o', san, '--tsgo', args.sanitized_checker, '--sanitize'])
source = (root / 'cohere/internal/lint/rules/react/jsx_no_constructed_context_values_test.go').read_text()
literal = r'"(?:[^"\\]|\\.)*"'
rows = []
for match in re.finditer(r'\{\s*(' + literal + r'),\s*`([^`]*)`', source[:source.index('func TestJsxNoConstructedContextValuesRequiresTheTypedHarness')]):
    rows.append((json.loads(match[1]), match[2], '.tsx', []))
assert len(rows) == 41, len(rows)
for value in ['{a:1}', '[1]', '() => {}', 'function() {}', 'class {}', 'new Foo()', '/x/', '<div />', '<></>', '{...rest}', 'c ? {a:1} : {b:2}', 'c ? x : {b:2}', 'x || {a:1}', '{a:1} && x', 'x ?? {a:1}', '{a:1} as any', '({a:1})', '(({a:1}))', 'x = {a:1}', 'x += {a:1}', 'makeIt()', '"s"', '1', '`t`', 'a+b']:
    rows.append(('construction ' + value, 'function Component() { return <Ctx.Provider value={' + value + '} />; }', '.tsx', []))
provider = '<Ctx.Provider value={{a:1}} />'
for prefix, suffix in [('function Component(){return ', ';}'), ('function component(){return ', ';}'), ('const Component = () => ', ';'), ('const component = () => ', ';'), ('const thing = function Component(){return ', ';}'), ('const Component = function thing(){return ', ';}'), ('class C extends React.Component { render(){return ', ';}}'), ('class C extends React.PureComponent { other(){return ', ';}}'), ('class c extends Component { f=()=> ', ';}'), ('class C { Render(){return ', ';}}'), ('class C extends Foo.Component { render(){return ', ';}}'), ('const Component = React.memo(()=> ', ');'), ('const Component = React.forwardRef(()=> ', ');'), ('const Component = arbitrary(()=> ', ');'), ('function Émile(){return ', ';}'), ('function Ωmega(){return ', ';}'), ('function outer(){function Component(){return ', ';}}'), ('function Component(){function helper(){return ', ';}}')]:
    rows.append(('component '+prefix, prefix+provider+suffix, '.tsx', []))
for label,text in [
 ('unicode and trivia', 'const note="😀";\nfunction Component(){\n/*foo*/ function v(){};\nreturn <Ctx.Provider value={v} />;}'),
 ('member usage', 'function Component(){ const v={a:{}}; return <Ctx.Provider value={v.a}/>;}'),
 ('module object', 'const v={}; function Component(){return <Ctx.Provider value={v}/>;}'),
 ('outer object', 'function outer(){const v={}; return function Component(){return <Ctx.Provider value={v}/>;};}'),
 ('object then number', 'function Component(){var v={};var v=1;return <Ctx.Provider value={v}/>;}'),
 ('number then object', 'function Component(){var v=1;var v={};return <Ctx.Provider value={v}/>;}'),
 ('cycle', 'function Component(){const a=b;const b=a;return <Ctx.Provider value={a}/>;}'),
 ('parenthesized createContext', 'const C=(React.createContext)();function Component(){return <C value={{}}/>;}'),
 ('other factory', 'const C=other.createContext();function Component(){return <C value={{}}/>;}'),
 ('no value', 'function Component(){return <Ctx.Provider/>;}'),
 ('boolean value', 'function Component(){return <Ctx.Provider value/>;}'),
 ('duplicate first value', 'function Component(){return <Ctx.Provider value="x" value={{}}/>;}'),
 ('multi provider', 'function Component(){const v={};return <><Ctx.Provider value={v}/><Ctx.Provider value={v}/></>;}')]:
    rows.append((label,text,'.tsx',[]))
for extension in ['.jsx','.js']:
    rows.append(('suffix '+extension,'function Component(){return <Ctx.Provider value={{}}/>;}',extension,[]))
rows.append(('unstable memo must refuse', 'function Component(){const dep={};const v=useMemo(()=>({dep}),[dep]);return <Ctx.Provider value={v}/>;}', '.tsx', []))

# Every literal fixture in the production memo/escape table is independent of the native source.
stability_source = (root / 'cohere/internal/lint/rules/react/jsx_no_constructed_context_values_stability_test.go').read_text()
for match in re.finditer(r'\{\s*(' + literal + r'),\s*`([^`]*)`', stability_source):
    rows.append(('memo '+json.loads(match[1]), match[2], '.tsx', []))

for match in re.finditer(r'source := `([^`]*)`', stability_source):
    rows.append(('production single memo', match[1], '.tsx', []))
rows += [
 ('no dependencies callback', 'function Component(){const v=useCallback(()=>1);return <Ctx.Provider value={v}/>;}', '.tsx', []),
 ('memo with no arguments', 'function Component(){return <Ctx.Provider value={useMemo()}/>;}', '.tsx', []),
 ('nonfunction memo factory', 'function Component(){return <Ctx.Provider value={useMemo(factory)}/>;}', '.tsx', []),
 ('foreign helper builds object', 'import {build} from "./helper";function Component(){const x=build();const v=useMemo(()=>({x}),[x]);return <Ctx.Provider value={v}/>;}', '.tsx', []),
 ('foreign helper caches object', 'import {cached} from "./helper";function Component(){const x=cached();const v=useMemo(()=>({x}),[x]);return <Ctx.Provider value={v}/>;}', '.tsx', []),
 ('foreign async helper', 'import {load} from "./helper";function Component(){const x=load();const v=useMemo(()=>({x}),[x]);return <Ctx.Provider value={v}/>;}', '.tsx', []),
 ('foreign context factory', 'function Component(){return <ExternalContext value={{}}/>;}', '.tsx', []),
 ('source unicode CRLF', 'function Émile(){\r\nconst o={};\r\nconst v=useMemo(()=>({o}),[o]);\r\nreturn <Ctx.Provider value={v}/>;}', '.tsx', []),
]

findings = 0
truths = []
fixture_records = []
for index, (label, text, extension, flags) in enumerate(rows):
    project = out / f'case-{index:03d}'
    project.mkdir(exist_ok=True)
    input_path = project / ('input' + extension)
    input_path.write_text(text)
    (project / 'helper.ts').write_text('export function build(){return {a:1};}\nlet old: {a:number}|undefined;\nexport function cached(){if(old)return old;const v={a:1};old=v;return v;}\nexport async function load(){return old;}\n')
    (project / 'globals.d.ts').write_text('// 😀 foreign declaration positions\ndeclare var Text: any;\ndeclare const ExternalF: any;\ndeclare const ExternalFragment = React.Fragment;\ndeclare const ExternalContext = React.createContext();\n')
    config = project / 'tsconfig.json'
    config.write_text(json.dumps({'compilerOptions': {'strict': True, 'allowJs': True, 'jsx': 'preserve', 'target': 'ES2022', 'types': []}, 'files': [input_path.name, 'globals.d.ts']}))
    manifest = project / 'manifest'
    manifest.write_text(str(input_path) + '\n')
    command = [config, manifest, *flags]
    truth = run(f'{index}-go', [oracle, *command])[0]
    for kind, binary in [('native', native), ('asan', san)]:
        got, errors, _ = run(f'{index}-{kind}', [binary, *command])
        assert not errors and got == truth, (label, kind, got, truth)
    findings += int(truth.splitlines()[-1].split()[-1])
    truths.append((command, truth))
    fixture_records.append({'label': label, 'source': text, 'extension': extension, 'flags': flags})
print('PASS', len(rows), 'controls', findings, 'findings; complete Go/native/sanitizer bytes', flush=True)
(out / 'fixtures.json.gz').write_bytes(gzip.compress(json.dumps(fixture_records, indent=2).encode(), mtime=0))
mutated = own / 'analysis.a'
original = mutated.read_text()
assert original.count("kind = 'object'") == 1
mutation = out / 'analysis-mutant.a'
mutation.write_text(re.sub(r"from '([^']+)'", lambda match: "from '" + str((own / match[1]).resolve()) + "'" if match[1].startswith('.') else match[0], original.replace("kind = 'object'", "kind = 'array'")))
changed = out / 'suite-mutant.a'
changed.write_text(re.sub(r"from '([^']+)'", lambda match: "from '" + str(mutation if match[1] == './analysis.a' else (own / match[1]).resolve()) + "'" if match[1].startswith('.') else match[0], entry.read_text()))
binary = out / 'mutant'
run('mutant-build', [args.compiler, 'build', changed, '-o', binary, '--tsgo', args.checker])
for command, truth in truths:
    got, errors, _ = run('mutant', [binary, *command])
    assert not errors
    if got != truth:
        print('KILLED construction-kind mutant: compiled, exit 0, empty stderr, Go bytes differ', flush=True)
        break
else:
    raise AssertionError('semantic mutant survived')
binary.unlink()
corpora = []
for name, config in [('repository', root / 'tsconfig.json'), ('compiler', Path('/workspace/wave20-typescript/src/compiler/tsconfig.json'))]:
    manifest = Path('/workspace/wave20-validation/harness-third') / (name + '.manifest')
    truth = run(name + '-go', [oracle, config, manifest])[0]
    for kind, binary in [('native', native), ('asan', san)]:
        got, errors, _ = run(name + '-' + kind, [binary, config, manifest])
        assert not errors and got == truth, (name, kind)
    print('PASS', name, len(truth), 'complete corpus bytes normal/sanitized', flush=True)
    corpora.append((name, config, manifest))
released_source = own / 'released.a'
released = out / 'released'
run('released-build', [args.compiler, 'build', released_source, '-o', released, '--tsgo', args.checker])
probe = out / 'probe.a'
probe.write_text('const x = make();\n')
config = out / 'tsconfig.json'
config.write_text('{"compilerOptions":{"strict":true,"target":"ES2022","types":[]},"sourceExtensions":[".a"]}')
for question in ['binding-origin', 'type-shape', 'resolved-callee']:
    _, errors, _ = run('released-' + question, [released, config, probe, question], expected=70)
    assert b'invalid or released checker handle' in errors
    print('PASS released MemoView', question, 'query panics 70', flush=True)
retained_source = out / 'retained.a'
retained_source.write_text(re.sub(r"from '([^']+)'", lambda match: "from '" + str((own / match[1]).resolve()) + "'" if match[1].startswith('.') else match[0], released_source.read_text().replace('tsgoRelease(program);', '')))
retained = out / 'retained'
run('retained-build', [args.compiler, 'build', retained_source, '-o', retained, '--tsgo', args.checker])
for question in ['binding-origin', 'type-shape', 'resolved-callee']:
    _, errors, _ = run('retained-' + question, [retained, config, probe, question])
    assert not errors
    print('KILLED retained-handle mutant', question, 'compiled, exit 0; required panic 70 catches it', flush=True)
retained.unlink()
bench = {}
for name, config, manifest in corpora:
    times = {'native': [], 'go': []}
    for round_index in range(3):
        outputs = {}
        for kind in (['native', 'go'] if round_index % 2 == 0 else ['go', 'native']):
            output, errors, elapsed = run(name + '-' + kind + '-timed', [native if kind == 'native' else oracle, config, manifest])
            if kind == 'native':
                assert not errors
            outputs[kind] = output
            times[kind].append(elapsed)
        assert outputs['native'] == outputs['go']
    bench[name] = {kind: statistics.median(values) for kind, values in times.items()}
    print('TIMING', name, bench[name], flush=True)
(out / 'summary.json').write_text(json.dumps({'controls': len(rows), 'findings': findings, 'semantic_mutant': 'caught by bytes', 'corpora': 'both normal/sanitized PASS', 'released_question': 'panic 70', 'bench': bench, 'shared_registration': False}, indent=2) + '\n')
print('PASS complete analysis controls, both corpora, semantic mutant, sanitizer and released question; shared checker integration pending', flush=True)
