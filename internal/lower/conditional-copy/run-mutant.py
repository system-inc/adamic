from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[3]
source = root / 'internal/load/rules.go'
saved = source.read_text()
invocation = 'return rule_runner.RunRule(p.compiler, typeChecker, files, name)'
assert saved.count(invocation) == 1
logs = Path('/workspace/scratch/conditional-mutant')
logs.mkdir(exist_ok=True)
try:
    source.write_text(saved.replace(invocation, 'return nil, nil'))
    with (logs / 'refusal.log').open('w') as output:
        result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^TestConditionalCopyRefusedByCohere$', '-v', '-count=1', '-timeout', '30m'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
    text = (logs / 'refusal.log').read_text()
    assert result.returncode != 0 and 'want cohere conditional refusal at 4:27, got <nil>' in text and '[build failed]' not in text, text
    print('Removed cohere invocation: caught by conditional refusal assertion, got nil; no build failure.')
    environment = os.environ.copy()
    environment.update(ADAMIC_CONDITIONAL_COPY_ORACLE='1', ADAMIC_GATE_UNCACHED='1')
    with (logs / 'unguarded-oracle.log').open('w') as output:
        result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^TestNativeAgreesWithNode/internal/oracle/testdata/type_rules_conditional_copy.a$', '-v', '-count=1', '-timeout', '30m'], cwd=root, env=environment, stdout=output, stderr=subprocess.STDOUT)
    assert result.returncode == 0, (logs / 'unguarded-oracle.log').read_text()
    print('Unguarded conditional copy: normal native/JavaScript/Node and sanitizer oracle passes.')
finally:
    source.write_text(saved)
    print('Invocation restored.')
