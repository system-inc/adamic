"""Ordered, witnessed discovery in a private copy; never a native correctness pass."""
import json
from pathlib import Path
import re
import shutil

HERE = Path(__file__).resolve().parent


def diagnostic(text):
    match = re.search(r'adamic: (.+?):(\d+):(\d+): (.*)', text)
    if not match:
        match = re.search(r'([^\n]+?main\.c):(\d+):(\d+): error: (.*)', text)
    if not match:
        return None
    file, line, column, message = match.groups()
    return {'path': file, 'file': file.split('/src/compiler/')[-1] if '/src/compiler/' in file else Path(file).name,
            'line': int(line), 'column': int(column), 'message': message}


def signature(message):
    # Locations and particular enum slots differ between scanner and small witnesses.
    return re.sub(r'slot \S+;', 'slot <member>;', message)


def catalogue():
    rows = []
    for directory in ['native3', 'land-area-next', 'combined-records-library']:
        source = HERE / 'evidence' / directory
        if not (source / 'witnesses/results.json').exists():
            continue
        for item in json.loads((source / 'witnesses/results.json').read_text()):
            stop = diagnostic(item.get('build_stderr', ''))
            if stop and item.get('node_exit') == 0 and item.get('build_exit') == 1:
                rows.append({'source': item['source'], 'message': stop['message'], 'provenance': directory + '/' + item['name']})
    return rows


def walk(out, raw_slice, compiler, compiler_tree, cache, run, split=0):
    root = out / 'discovery'
    shutil.copytree(raw_slice, root / 'adapted')
    for name in ['main.a', 'token-names.a']:
        shutil.copyfile(out / ('split-' + str(split)) / name, root / name)
    originals = {p.relative_to(root).as_posix(): p.read_text() for p in root.rglob('*') if p.suffix in {'.ts', '.a'} and p.is_file()}
    rows = []
    candidates = catalogue()
    api = cache / 'api/node_modules/typescript/lib/typescript.js'
    env = {'SCANNER_TYPESCRIPT': str(api), 'ADAMIC_NATIVE_SPLIT': str(split)}
    seen = set()
    for order in range(1, 16):
        stem = f'discovery-{order:02}'
        code = run([compiler, 'build', root / 'main.a', '-o', root / 'scanner'], stem, compiler_tree, env)
        if code == 0:
            (root / 'completion.json').write_text(json.dumps({'status': 'build-with-placeholders', 'native_correctness_claim': False}) + '\n')
            break
        stop = diagnostic((out / (stem + '.stderr')).read_text())
        if not stop:
            rows.append({'order': order, 'build_exit': code, 'unsupported_diagnostic': stem + '.stderr', 'witness_unavailable': True})
            break
        stop.update(order=order, build_exit=code, scope='first unchanged stop' if order == 1 else 'behind throwing discovery placeholders')
        # Witnesses are rerun on this compiler; old evidence never substitutes for a fresh result.
        matches = [c for c in candidates if signature(c['message']) == signature(stop['message'])]
        witness = None
        for number, candidate in enumerate(matches):
            name = f'witness-{order:02}-{number:02}'
            file = root / (name + '.a')
            file.write_text(candidate['source'])
            helper = HERE / 'scratch-witness.cjs'
            node = run(['node', helper, file], name + '-node', compiler_tree, env)
            built = run([compiler, 'build', file, '-o', root / (name + '-native')], name + '-build', compiler_tree, env)
            result = diagnostic((out / (name + '-build.stderr')).read_text())
            if node == 0 and built != 0 and result and signature(result['message']) == signature(stop['message']):
                witness = {'file': str(file.relative_to(out)), 'node_exit': node, 'build_exit': built,
                           'message': result['message'], 'catalogue_origin': candidate['provenance']}
                break
        stop['witness'] = witness
        rows.append(stop)
        (root / 'stops.json').write_text(json.dumps(rows, indent=2) + '\n')
        if not witness:
            stop['continuation'] = 'stopped: no freshly verified Node witness for this diagnostic'
            break
        key = (stop['file'], stop['line'], stop['column'], stop['message'])
        if key in seen:
            stop['continuation'] = 'stopped: diagnostic repeated after placeholder'
            break
        seen.add(key)
        if order == 15:
            stop['continuation'] = 'fifteen-stop bound reached'
            break
        file = Path(stop['path'])
        # Only source inside this discovery copy may be modified.
        if not file.is_relative_to(root):
            stop['continuation'] = 'stopped: diagnostic is outside discovery source'
            break
        text = file.read_text()
        placeholder = None
        if 'index signature' in stop['message'] and '[index: string]: T;' in text:
            text = text.replace('[index: string]: T;', '') + '\nexport function discoveryMapLike(): never { throw new Error("discovery placeholder: MapLike"); }\n'
            placeholder = 'index-signature removal with throwing marker; later record observations are dependent'
        elif 'refuses a namespace' in stop['message'] and 'export namespace Debug' in text:
            text = text[:text.index('export namespace Debug')] + '''export const Debug = {
 isDebugging: false,
 fail: (message?: string, stackCrawlMark?: {}): never => { throw new Error("discovery placeholder: Debug.fail"); },
 assertEqual: <T>(a: T, b: T, msg?: string, msg2?: string, stackCrawlMark?: {}): void => { throw new Error("discovery placeholder: Debug.assertEqual"); }
};
'''
            placeholder = 'namespace replacement with throwing functions; later Debug observations are dependent'
        elif 'mutable namespace export' in stop['message'] and 'export let isDebugging = false;' in text:
            text = text.replace('export let isDebugging = false;', 'export const isDebugging = false;') + '\nexport function discoveryDebug(): never { throw new Error("discovery placeholder: Debug state"); }\n'
            placeholder = 'immutable namespace marker with dormant throw; namespace signatures preserved'
        elif 'both null and undefined' in stop['message'] and 'export type CompilerOptionsValue =' in text:
            original = next(line for line in text.splitlines() if line.startswith('export type CompilerOptionsValue ='))
            text = text.replace(original, 'export type CompilerOptionsValue = string;\nexport function discoveryCompilerOptionsValue(): never { throw new Error("discovery placeholder: CompilerOptionsValue"); }')
            placeholder = 'type-only nullable union placeholder with throwing marker'
        elif 'utf16EncodeAsStringWorker' in text.splitlines()[stop['line']-1]:
            original = next(line for line in text.splitlines() if line.startswith('const utf16EncodeAsStringWorker:'))
            text = text.replace(original, 'const utf16EncodeAsStringWorker: (codePoint: number) => string = (codePoint: number): string => { throw new Error("discovery placeholder: UTF16 worker"); };')
            placeholder = 'typed throwing UTF16 worker'
        elif 'Script_Extensions: undefined! as Set<string>' in text.splitlines()[stop['line']-1]:
            text = text.replace('Script_Extensions: undefined! as Set<string>,', 'Script_Extensions: (() : Set<string> => { throw new Error("discovery placeholder: Script_Extensions"); })(),')
            placeholder = 'typed throwing initializer'
        elif file.name == 'diagnosticInformationMap.generated.ts' and 'adamic/no-unchecked-cast' in stop['message']:
            owner_line = next(i for i, line in enumerate(text.splitlines(), 1) if line.startswith('function diag('))
            code = run(['node', HERE / 'scratch-stub.cjs', file, str(owner_line), '1', 'DiagnosticMessage'], stem + '-placeholder', compiler_tree, env)
            placeholder = 'typed throwing diagnostic factory; later casts are placeholder-dependent' if code == 0 else None
        elif 'function returning' in stop['message']:
            code = run(['node', HERE / 'scratch-stub.cjs', file, str(stop['line']), str(stop['column']), 'never'], stem + '-placeholder', compiler_tree, env)
            placeholder = 'throwing function with never return; later return typing is placeholder-dependent' if code == 0 else None
        else:
            code = run(['node', HERE / 'scratch-stub.cjs', file, str(stop['line']), str(stop['column'])], stem + '-placeholder', compiler_tree, env)
            placeholder = 'smallest enclosing function body throws; signature retained/inferred' if code == 0 else None
        if not placeholder:
            stop['continuation'] = 'stopped: no verified placeholder for this source owner'
            break
        if file.read_text() != text and 'smallest enclosing' not in placeholder and 'throwing function with never' not in placeholder and 'typed throwing diagnostic factory' not in placeholder:
            file.write_text(text)
        stop['placeholder'] = placeholder
    changes = []
    for relative, before in originals.items():
        after = (root / relative).read_text()
        if before != after:
            changes.append({'file': relative, 'before': before, 'after': after})
    (root / 'placeholders.json').write_text(json.dumps({'label': 'uncommitted discovery only; no native correctness claim', 'changes': changes}, indent=2) + '\n')
    (root / 'stops.json').write_text(json.dumps(rows, indent=2) + '\n')
    return rows
