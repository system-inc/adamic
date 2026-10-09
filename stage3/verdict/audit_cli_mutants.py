#!/usr/bin/env python3
"""Run real stock cases with each CLI bridge deliberately broken."""
import argparse
import json
from pathlib import Path
from unittest.mock import patch
from run import ROOT, baseline_suite
from prove_groups import subset


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('manifest', type=Path)
    parser.add_argument('tree', type=Path)
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    manifest = json.loads(args.manifest.read_text())
    args.output.mkdir()
    mutants = [('literal-null bridge omitted', 'jsxFactoryAndJsxFragmentFactoryNull',
                'run.config_options', {}, 'stdout'),
               ('declaration outputs-skipped exit becomes 2', 'jsDeclarationsCrossfileMerge',
                'run.expected_exit', 2, 'exit'),
               ('provided type packages force explicit typeRoots', 'nodeNextImportModeImplicitIndexResolution2',
                'run.has_type_packages', False, 'stdout')]
    results = []
    for index, (name, stem, function, value, stream) in enumerate(mutants):
        row = next(row for row in manifest['cases'] if Path(row['source']).stem == stem)
        with patch(function, return_value=value):
            report = baseline_suite(ROOT / 'standins/node.sh', args.tree.resolve(),
                                    args.output / str(index), None, subset(manifest, [row]))
        if report['failed'] != 1 or set(report['failures'][0]['differences']) != {stream}:
            raise RuntimeError(f'{name}: mutant escaped or hit the wrong check')
        results.append({'mutant': name, 'case': row['source'], 'configuration': row['configuration'],
                        'caught_only_by': stream})
        print(name + ': caught only by ' + stream, flush=True)
    (args.output / 'proof.json').write_text(json.dumps({'proved': True, 'mutants': results}, indent=2) + '\n')


if __name__ == '__main__':
    main()
