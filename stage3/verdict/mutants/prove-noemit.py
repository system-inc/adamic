"""Compare stock Node with mutations of the same no-emit project."""
import hashlib
import json
import os
from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parent
sys.path.insert(0, str(ROOT.parent))
from run import compare_case, materialize_command, output_tree

CASE = 'synthetic-noemit-incremental'
OUTCOMES = (
    ('stock', None, None),
    ('no-emit-incremental-js', 'unexpected.js', 'extra'),
    ('no-emit-incremental-missing', 'tsconfig.tsbuildinfo', 'missing'),
    ('buildinfo-byte', 'tsconfig.tsbuildinfo', 'bytes'),
)


def outputs(folder, side):
    tree = output_tree(folder / side / 'project')
    return {name: {'bytes': len(content), 'sha256': hashlib.sha256(content).hexdigest()}
            for name, (kind, content) in tree.items() if name not in ('emit.ts', 'tsconfig.json')}


def prove_build_info_layout(output, reference):
    stock = ROOT.parent / 'standins/node.sh'
    correct = 'out/MixedCase/project/tsconfig.tsbuildinfo'
    moved = 'out/MixedCase/reference/project/tsconfig.tsbuildinfo'
    custom = 'custom/state.bin'
    base = {'noEmit': True, 'incremental': True, 'outDir': 'out', 'rootDir': '../..'}
    for mode, candidate, reference_path, candidate_path, expected_reference, expected_candidate in (
        ('stock-layout', stock, None, None, correct, correct),
        ('moved-layout', ROOT / 'move-build-info.py', None, None, correct, moved),
        ('custom-both', stock, custom, custom, custom, custom),
        ('custom-candidate-only', stock, None, custom, correct, custom),
    ):
        folder = output / mode / 'MixedCase'
        options = dict(base, tsBuildInfoFile=reference_path) if reference_path else base
        candidate_options = dict(base, tsBuildInfoFile=candidate_path) if candidate_path else base
        differences = compare_case(candidate, reference, folder, {'emit.ts': 'export const n = 1;\n'},
                                   options, candidate_options=candidate_options)
        node_outputs = outputs(folder, 'reference')
        candidate_outputs = outputs(folder, 'candidate')
        exits = [(folder / side / 'actual.exit').read_text().strip() for side in ('reference', 'candidate')]
        print(json.dumps({'case': mode, 'compiler_exits': exits, 'node_outputs': node_outputs,
                          'candidate_outputs': candidate_outputs, 'differences': differences}), flush=True)
        assert exits == ['0', '0'], exits
        assert list(node_outputs) == [expected_reference], node_outputs
        assert list(candidate_outputs) == [expected_candidate], candidate_outputs
        if expected_candidate == expected_reference:
            assert not differences and node_outputs == candidate_outputs, differences
        else:
            assert {path: diff['kind'] for path, diff in differences['files'].items()} == {
                expected_reference: 'missing', expected_candidate: 'extra',
            }, differences
            if mode == 'moved-layout':
                assert node_outputs[correct] == candidate_outputs[moved]
                assert differences['candidate_build_info']['path'] == correct
            else:
                assert set(differences) == {'files'}, differences


def main():
    output = Path(sys.argv[1]).resolve()
    output.mkdir()
    os.environ['STAGE3_VERDICT_SYNTHETIC_FIXTURE'] = '1'
    reference = ['node', str(Path(os.environ['STAGE3_VERDICT_NODE_TSC']).resolve())]
    content, options = materialize_command(
        b'// @noEmit: true\n// @incremental: true\nexport const n = 1;\n')
    for mode, path, kind in OUTCOMES:
        folder = output / mode / CASE
        candidate = ROOT.parent / 'standins/node.sh' if path is None else ROOT / (mode + '.sh')
        differences = compare_case(candidate, reference, folder, {'emit.ts': content}, options)
        node_outputs = outputs(folder, 'reference')
        candidate_outputs = outputs(folder, 'candidate')
        assert list(node_outputs) == ['tsconfig.tsbuildinfo'], node_outputs
        assert (folder / 'reference/actual.exit').read_text() == '0\n'
        if path is None:
            assert not differences and node_outputs == candidate_outputs, differences
        else:
            checks = {'files'}
            if kind == 'extra':
                checks.add('candidate_files')
            elif kind == 'missing':
                checks.add('candidate_build_info')
            assert set(differences) == checks and set(differences['files']) == {path}, differences
            assert differences['files'][path]['kind'] == kind, differences
        print(json.dumps({'case': CASE, 'standin': mode, 'exit': int(bool(differences)),
                          'node_outputs': node_outputs, 'candidate_outputs': candidate_outputs,
                          'differences': differences}), flush=True)

    for case, directives, expected_path in (
        ('synthetic-noemit-composite-custom', b'// @composite: true\n// @tsBuildInfoFile: cache/state.bin\n',
         'cache/state.bin'),
        ('synthetic-noemit-configdir', b'// @incremental: true\n// @tsBuildInfoFile: ${configDir}/cache/state.bin\n',
         'cache/state.bin'),
        ('synthetic-noemit-dot-outfile', b'// @incremental: true\n// @module: amd\n// @outFile: .js\n',
         '.tsbuildinfo'),
        ('synthetic-noemit-outfile', b'// @incremental: true\n// @module: amd\n// @outFile: out/bundle.js\n',
         'out/bundle.tsbuildinfo'),
        ('synthetic-noemit-outfile-trailing-separator',
         b'// @incremental: true\n// @module: system\n// @outFile: out/bundle.js/\n',
         'out/bundle.tsbuildinfo'),
        ('synthetic-noemit-trailing-separator', b'// @incremental: true\n// @outDir: out/\n',
         'out/tsconfig.tsbuildinfo'),
    ):
        content, options = materialize_command(b'// @noEmit: true\n' + directives + b'export const n = 1;\n')
        folder = output / case
        differences = compare_case(ROOT.parent / 'standins/node.sh', reference, folder,
                                   {'emit.ts': content}, options)
        node_outputs = outputs(folder, 'reference')
        candidate_outputs = outputs(folder, 'candidate')
        actual_exits = [(folder / side / 'actual.exit').read_text().strip() for side in ('reference', 'candidate')]
        print(json.dumps({'case': case, 'standin': 'stock', 'exit': int(bool(differences)),
                          'compiler_exits': actual_exits, 'node_outputs': node_outputs,
                          'candidate_outputs': candidate_outputs, 'differences': differences}), flush=True)
        assert actual_exits == ['0', '0'] and list(node_outputs) == [expected_path], (actual_exits, node_outputs)
        assert node_outputs == candidate_outputs and not differences, differences

    for side in ('reference', 'candidate'):
        folder = output / f'{side}-noemit'
        differences = compare_case(ROOT.parent / 'standins/node.sh', reference, folder,
                                   {'input.d.ts': 'declare const value: number;\n'},
                                   {'noEmit': side == 'reference'}, candidate_options={'noEmit': side == 'candidate'})
        exits = [(folder / item / 'actual.exit').read_text().strip() for item in ('reference', 'candidate')]
        print(json.dumps({'case': folder.name, 'compiler_exits': exits, 'differences': differences}), flush=True)
        assert exits == ['0', '0'] and not differences, differences

    folder = output / 'mixed-noemit-real-output'
    differences = compare_case(ROOT.parent / 'standins/node.sh', reference, folder,
                               {'emit.ts': 'export const value = 1;\n'},
                               {'noEmit': True}, candidate_options={})
    exits = [(folder / side / 'actual.exit').read_text().strip() for side in ('reference', 'candidate')]
    print(json.dumps({'case': folder.name, 'compiler_exits': exits, 'differences': differences}), flush=True)
    assert exits == ['0', '0'] and set(differences) == {'files'}, (exits, differences)
    assert set(differences['files']) == {'emit.js'} and differences['files']['emit.js']['kind'] == 'extra', differences

    prove_build_info_layout(output, reference)


if __name__ == '__main__':
    main()
