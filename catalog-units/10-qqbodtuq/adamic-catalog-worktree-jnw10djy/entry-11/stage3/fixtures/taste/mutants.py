#!/usr/bin/env python3
"""Run explicit source rewrites on main; retain masked mutants as observations."""
import argparse
import json
from pathlib import Path
import shutil
import subprocess

p = argparse.ArgumentParser()
p.add_argument('--main', type=Path, required=True)
p.add_argument('--output', type=Path, required=True)
args = p.parse_args()
bucket = Path(__file__).resolve().parent
args.output.mkdir(parents=True, exist_ok=True)
shutil.copytree(bucket / 'support', args.output / 'support', dirs_exist_ok=True)
status = {x['file']: x for x in json.loads((bucket / 'status.json').read_text())}
rewrites = [
    ('operand', '01_diagnostic_code.a', 'return d.canonicalHead?.code || d.code;', 'const code = d.canonicalHead?.code;\n    return code === undefined || code === 0 || Number.isNaN(code) ? d.code : code;'),
    ('object_condition', '03_diagnostic_file.a', 'diagnostic.file ?', 'diagnostic.file !== undefined ?'),
    ('string_condition', '04_jsx_runtime.a', 'return base ?', 'return base !== undefined && base.length !== 0 ?'),
    ('optional_double_not', '05_this_type.a', 'return !!(type.flags & TypeFlags.TypeParameter && (type as TypeParameter).isThisType);', 'return (type.flags & TypeFlags.TypeParameter) !== 0 && (type as TypeParameter).isThisType === true;'),
    ('logical_assignment', '17_binder_flow.a', 'hasFlowEffects ||= saveHasFlowEffects;', 'if (!hasFlowEffects) hasFlowEffects = saveHasFlowEffects;'),
    ('comma', '16_scan_exclamation.a', None, None),
    ('void', '13_void_callback.a', 'diag => void diagnostics.push(diag)', 'diag => { diagnostics.push(diag); }'),
    ('label', '14_relative_complement.a', 'loopB:', ''),
    ('double_not', '19_deprecated_flags.a', '!!(getCombinedNodeFlagsCached(declaration) & NodeFlags.Deprecated)', '(getCombinedNodeFlagsCached(declaration) & NodeFlags.Deprecated) !== 0'),
    ('label_break', '20_jsdoc_terminate.a', 'terminate:', ''),
    ('named_export', '18_named_export.a', None, None),
    ('assignment_once', '22_assignment_once.a', None, None),
    ('finally_jumps', '23_labels_finally.a', None, None),
    ('truthy_loops', '21_truthy_loops.a', None, None),
    ('star_export', '15_barrel.a', 'export * from "./support/core.a";', ''),
]
results = []
for form, file, before, after in rewrites:
    text = (bucket / file).read_text()
    if form == 'assignment_once':
        text = text.replace('receiver().value &&= right();', '{ const slot = receiver(); const value = slot.value; if (value !== undefined && value !== 0 && !Number.isNaN(value)) slot.value = right(); }')
        text = text.replace('arrayReceiver()[index()] &&= right();', '{ const array = arrayReceiver(); const key = index(); const value = array[key]; if (value !== undefined && value !== 0 && !Number.isNaN(value)) array[key] = right(); }')
        text = text.replace('receiver().value ??= right();', '{ const slot = receiver(); if (slot.value === undefined) slot.value = right(); }')
        text = text.replace('arrayReceiver()[index()] ??= right();', '{ const array = arrayReceiver(); const key = index(); if (array[key] === undefined) array[key] = right(); }')
        text = text.replace('expressionMayContainStrings &&= mayContainStrings;', 'if (expressionMayContainStrings) expressionMayContainStrings = mayContainStrings;')
    elif form == 'truthy_loops':
        text = text.replace('while (node)', 'while (node !== undefined)').replace('; label; label =', '; label !== undefined; label =')
        text = text.replace('if (type.flags & TypeFlags.IndexedAccess)', 'if ((type.flags & TypeFlags.IndexedAccess) !== 0)').replace('while (type.flags & TypeFlags.IndexedAccess)', 'while ((type.flags & TypeFlags.IndexedAccess) !== 0)')
    elif form == 'finally_jumps':
        text = text.replace('loopB:', '').replace('loopA:', '')
        header = 'for (let offsetB = 0; offsetB < 3; offsetB += 1) {'
        text = text.replace(header, header + '\n        let stop = false; let nextOuter = false;')
        text = text.replace('continue loopA;', 'continue;').replace('continue loopB;', '{ nextOuter = true; break; }').replace('break loopA;', 'break;').replace('break loopB;', 'stop = true; break;')
        text = text.replace('        console.log("outer-tail:"', '        if (stop) break; if (nextOuter) continue;\n        console.log("outer-tail:"')
    elif form == 'named_export':
        support = args.output / 'support/performance.a'
        support.write_text(support.read_text().replace('const performance', 'export const performance').replace('export { performance };', ''))
    elif form == 'label_break':
        text = text.replace('terminate:', '').replace('break terminate;', 'scanner.setSkipJsDocLeadingAsterisks(false); return finishNode(moduleTag, pos);')
    elif form == 'comma':
        for width, token in [(3, 'ExclamationEqualsEqualsToken'), (2, 'ExclamationEqualsToken')]:
            before = f'return pos += {width}, token = SyntaxKind.{token};'
            assert before in text
            text = text.replace(before, f'pos += {width}; return token = SyntaxKind.{token};')
    elif form == 'label':
        text = text.replace('loopB:', '').replace('loopA:', '')
        header = 'for (let offsetA = 0, offsetB = 0; offsetB < arrayB.length; offsetB++) {'
        text = text.replace(header, header + '\n        let nextB = false;')
        text = text.replace('continue loopB;', 'nextB = true; break;')
        text = text.replace('continue loopA;', 'continue;')
        text = text.replace('continue;\n            }\n        }', 'continue;\n            }\n            if (nextB) break;\n        }')

    else:
        assert text.count(before) == 1, (form, before)
        text = text.replace(before, after)
    target = args.output / file
    target.write_text(text)
    node_command = ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(target)]
    node = subprocess.run(node_command, cwd=args.main, capture_output=True, timeout=60)
    binary = args.output / (file + '.bin')
    command = ['go', 'run', './cmd/adamic', 'build', str(target), '-o', str(binary)]
    build = subprocess.run(command, cwd=args.main, capture_output=True, timeout=180)
    diagnostic = (build.stdout + build.stderr).decode()
    outcome = 'Compiles' if build.returncode == 0 else 'NotYet' if "stage 0 can't lower" in diagnostic else 'Refused' if 'Adamic 0.1 refuses' in diagnostic else 'Checker'
    native = subprocess.run([str(binary)], capture_output=True, timeout=60) if build.returncode == 0 else None
    node_observation = {'stdout': node.stdout.decode(), 'stderr': node.stderr.decode(), 'exit': node.returncode}
    result = {'form': form, 'file': file, 'before': status[file]['stage0']['outcome'], 'after': outcome, 'what': diagnostic, 'command': command, 'node': node_observation, 'same_node': node_observation == status[file]['node'], 'flips_outcome': status[file]['stage0']['outcome'] != outcome}
    if native:
        result['native'] = {'stdout': native.stdout.decode(), 'stderr': native.stderr.decode(), 'exit': native.returncode}
        result['matches_node'] = result['native'] == node_observation
    (args.output / (form + '.json')).write_text(json.dumps(result, indent=2) + '\n')
    results.append(result)
    print(form, result['before'], '->', outcome, 'same Node:', result['same_node'], flush=True)
(bucket / 'mutants.json').write_text(json.dumps(results, indent=2) + '\n')
