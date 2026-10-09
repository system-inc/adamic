#!/usr/bin/env python3
"""Run real matcher mutants sequentially, restore every source even on failure."""
import pathlib
import subprocess
root = pathlib.Path(__file__).resolve().parents[3]
mutants = [
    ('greedy-as-lazy', 'matcher.go', 'greedy: n.Greedy', 'greedy: false'),
    ('captures-not-reset', 'matcher.go', 'for _, id := range i.clear {', 'for _, id := range []int{} {'),
    ('lookbehind-left-to-right', 'matcher.go', 'direction = -1', 'direction = 1'),
    ('case-fold-without-u-distinction', 'canonicalize.go', 'if unicodeMode(f) {', 'if true {'),
    ('provider-string-not-snapshotted', 'sets.go', 'set.strings[index] = slices.Clone(text)', 'set.strings[index] = text'),
    ('step-limit-as-failed-match', 'matcher.go', 'if err := budget.step(); err != nil {\n\t\t\treturn s, false, err', 'if err := budget.step(); err != nil {\n\t\t\treturn s, false, nil'),
    ('unavailable-property-as-empty-set', 'sets.go', 'data, err := p.Lookup(property)\n\t\tif err != nil {\n\t\t\treturn set, err', 'data, err := p.Lookup(property)\n\t\tif err != nil {\n\t\t\treturn set, nil'),
]
for name, file, old, new in mutants:
    target = root / 'internal/regexp' / file
    original = target.read_text()
    if original.count(old) != 1:
        raise SystemExit(f'{name}: expected one mutation site, got {original.count(old)}')
    backups = {target: original}
    test_name = 'TestMatcherNodeControls'
    witness = 'DISAGREEMENT'
    if name == 'provider-string-not-snapshotted':
        test_name, witness = 'TestMatcherProviderSnapshot', 'property provider snapshot got='
    if name == 'step-limit-as-failed-match':
        test_name, witness = 'TestMatcherStepLimit', 'catastrophic backtracking returned <nil>'
    if name == 'unavailable-property-as-empty-set':
        test_name, witness = 'TestMatcherPropertyProviderStrings', 'missing property silently compiled'
    log = pathlib.Path('/tmp') / f'regex-mutant-{name}.log'
    try:
        target.write_text(original.replace(old, new))
        if name == 'case-fold-without-u-distinction':
            sets = root / 'internal/regexp/sets.go'
            backups[sets] = sets.read_text()
            site = 'if unicodeMode(f) {\n\t\ttable = simpleCaseFold[:]'
            if backups[sets].count(site) != 1:
                raise SystemExit('expected one compiled set folding mutation site')
            sets.write_text(backups[sets].replace(site, 'if true {\n\t\ttable = simpleCaseFold[:]'))
        with log.open('w') as output:
            result = subprocess.run(['go','test','-v','-count=1','-timeout','30s','-run','^'+test_name+'$','./internal/regexp'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
        text = log.read_text()
        differences = [line.strip() for line in text.splitlines() if witness in line]
        if result.returncode == 0 or not differences:
            raise SystemExit(f'{name}: NOT CAUGHT by result oracle; see {log}')
        print(f'{name}: caught, exit={result.returncode}, {len(differences)} oracle failures; {log}')
        print(differences[0])
    finally:
        for path, contents in backups.items():
            path.write_text(contents)
