"""Generate input fixtures from the pinned upstream tests without editing their harness."""
import json
import re
from pathlib import Path

QUOTED = r'"(?:\\.|[^"\\])*"|`[^`]*`'

def strings(text):
    return [value[1:-1] if value.startswith('`') else json.loads(value) for value in re.findall(QUOTED, text)]

def source_constant(text, name):
    tail = text.split('var ' + name + ' = ', 1)[1]
    if 'strings.Join([]string{' in tail[:100]:
        block = tail.split('strings.Join([]string{', 1)[1].split('}, "\\n")', 1)[0]
    else:
        block = tail.split('(', 1)[1].split('\n)', 1)[0]
    block = re.sub(r'(?m)^\s*//.*$', '', block)
    return '\n'.join(strings(block)) + '\n'

def controls(repository, directory):
    roots = []
    descriptions = []
    nexus = repository / 'cohere/internal/lint/rules/nexus'
    for stem, prefix, wrap in [
        ('correctness_require_child_process_error_listener', 'correctnessRequireChildProcessErrorListener', False),
        ('correctness_require_response_status_check', 'correctnessRequireResponseStatusCheck', False),
        ('performance_no_independent_await_in_loop', 'performanceNoIndependentAwaitInLoop', True),
    ]:
        text = (nexus / (stem + '_test.go')).read_text()
        prelude = source_constant(text, prefix + 'Prelude')
        if stem.startswith('correctness_require_child'):
            (directory / 'node.d.ts').write_text(source_constant(text, prefix + 'NodeTypes'))
        if stem.startswith('correctness_require_response'):
            service = directory / 'source/services/network/NetworkService.ts'
            service.parent.mkdir(parents=True, exist_ok=True)
            service.write_text(source_constant(text, prefix + 'NetworkService'))
            prelude = prelude.replace('../libraries/structure/source/services/network/NetworkService', './source/services/network/NetworkService')
        matches = list(re.finditer(r'\{(' + QUOTED + r'), \[\]string\{(.*?)\n\s*\}', text, re.S))
        if not matches:
            raise RuntimeError('no imported cases: ' + stem)
        for match in matches:
            name = strings(match.group(1))[0]
            body = '\n'.join(strings(match.group(2)))
            source = prelude + ('export async function subject():Promise<unknown>{\n' + body + '\nreturn undefined;\n}\n' if wrap else body) + '\nexport {};\n'
            path = directory / ('control-%03d.a' % len(roots))
            path.write_text(source)
            roots.append(path)
            descriptions.append({'path': str(path), 'rule': stem, 'case': name})
    # Additional structural, Unicode and exceptional-flow controls.
    extra = [
        "export async function f(){const r=await fetch('/');try{return await r.json();}finally{if(!r.ok)throw Error();}}",
        "export async function f(){const r=await fetch('/');const g=()=>r.ok;return r.json();}",
        "export async function f(){const r=await fetch('/');const text=await r.text();throw Error(text);}",
        "export async function f(){const r=await fetch('/');while(Math.random()){const data=await r.json();if(Math.random())break;}return 1;}",
        "/* 世界 🌍 */\r\nexport async function f(items:string[]){for(const é of items) {await Promise.resolve(é);}}\r\n",
    ]
    for source in extra:
        path = directory / ('control-%03d.a' % len(roots)); path.write_text(source); roots.append(path)
    (directory / 'cases.json').write_text(json.dumps(descriptions, indent=2)+'\n')
    (directory / 'controls.manifest').write_text('\n'.join(map(str, roots))+'\n')
    (directory / 'tsconfig.json').write_text(json.dumps({'compilerOptions': {'strict': True, 'target': 'ES2022', 'module': 'NodeNext', 'lib': ['ES2022', 'DOM', 'ESNext.Disposable'], 'noEmit': True, 'allowImportingTsExtensions': True}, 'files': ['node.d.ts', roots[0].name], 'sourceExtensions': ['.a'], 'include': ['*.a']}))
    return roots

if __name__ == '__main__':
    import sys
    root = Path(sys.argv[1]); output = Path(sys.argv[2]); output.mkdir(parents=True, exist_ok=True)
    print('generated', len(controls(root, output)), 'controls')
