from pathlib import Path
import subprocess, json, time, os
repository = Path(__file__).resolve().parents[6]
rule = Path(__file__).resolve().parents[1]
out = Path(os.environ.get('ADAMIC_WAVE15_INDIRECT_ARTIFACTS', '/tmp/wave15-indirect'))
out.mkdir(parents=True, exist_ok=True)
compiler = os.environ.get('ADAMIC_WAVE15_SIXTH_COMPILER', '/workspace/wave15-sixth-adamic')
(out/'indirect.tsx').write_bytes((rule/'testdata/indirect.jsx-source').read_bytes())
(out/'manifest').write_text(str(out/'indirect.tsx')+'\n')
(out/'tsconfig.json').write_text(json.dumps({'compilerOptions': {'target':'ES2022','module':'ESNext','strict':True,'jsx':'preserve'},'files':['indirect.tsx']}))
oracle_source = repository/'stage1/cohere/typeaware/rules/react-jsx-fragments/testdata/oracle.go'
virtual = repository/'cohere/adamic_wave15_indirect_oracle.go'
(out/'overlay.json').write_text(json.dumps({'Replace': {str(virtual): str(oracle_source)}}))
def run(command, label, cwd=None):
    start = time.perf_counter_ns()
    result = subprocess.run(command, capture_output=True, cwd=cwd)
    (out/(label+'.stdout')).write_bytes(result.stdout)
    (out/(label+'.stderr')).write_bytes(result.stderr)
    assert result.returncode == 0, (label, result.returncode, result.stderr.decode()[:2000])
    return result.stdout, time.perf_counter_ns()-start
run(['go','build','-overlay',str(out/'overlay.json'),'-o',str(out/'oracle'),str(virtual)], 'oracle-build', repository/'cohere')
truth, elapsed = run([str(out/'oracle'),str(out/'tsconfig.json'),str(out/'manifest')], 'oracle')
expected = b''.join(sorted(line for line in truth.splitlines(keepends=True) if len(line.split(b'\t')) >= 7 and line.split(b'\t')[2] == b'react/jsx-no-constructed-context-values'))
assert len(expected.splitlines()) == 14, truth.decode()
(out/'expected.stdout').write_bytes(expected)
print('Go source/checker/all rules ns', elapsed, flush=True)
for sanitize in [False, True]:
    label = 'indirect-asan' if sanitize else 'indirect'
    run([compiler,'build',str(rule/'indirect_controls.a'),'-o',str(out/label)] + (['--sanitize'] if sanitize else []), label+'-build')
    actual, elapsed = run([str(out/label)], label)
    assert actual == expected, (label, expected.decode(), actual.decode())
    print(label, 'PASS', len(actual.splitlines()), 'findings', len(actual), 'bytes; process ns', elapsed, flush=True)
# Keep an upstream panic distinct from successful byte agreement.
(out/'satisfies.tsx').write_bytes((rule/'testdata/satisfies-oracle-panic.jsx-source').read_bytes())
(out/'satisfies.manifest').write_text(str(out/'satisfies.tsx')+'\n')
(out/'satisfies.json').write_text(json.dumps({'compilerOptions':{'target':'ES2022','module':'ESNext','strict':True,'jsx':'preserve'},'files':['satisfies.tsx']}))
panic = subprocess.run([str(out/'oracle'),str(out/'satisfies.json'),str(out/'satisfies.manifest')],capture_output=True)
(out/'satisfies-go.stderr').write_bytes(panic.stderr)
assert panic.returncode == 2 and b'*ast.SatisfiesExpression, not *ast.AsExpression' in panic.stderr
refusal = subprocess.run([str(out/'indirect'),'--satisfies'],capture_output=True)
(out/'satisfies-native.stderr').write_bytes(refusal.stderr)
assert refusal.returncode == 70 and b'production Go constructed-context panics on SatisfiesExpression' in refusal.stderr
print('SatisfiesExpression Go panic exit 2; native explicit refusal exit 70 (not byte agreement)',flush=True)
mutations = [
    ('branch-order', 'return left === undefined ? constructionOf(values, node.second, depth + 1) : left;', 'return constructionOf(values, node.second, depth + 1);'),
    ('provider-factory', "callee.name === 'createContext'", "callee.name === 'unrelatedContext'"),
    ('function-kind', 'new Construction(262, last)', 'new Construction(210, last)'),
]
original = (rule/'numeric_constructions.a').read_text()
for label, before, after in mutations:
    assert before in original
    (rule/'mutant_tree.a').write_text(original.replace(before, after, 1))
    (rule/'mutant_rule.a').write_text((rule/'rule.a').read_text().replace("'./numeric_constructions.a'", "'./mutant_tree.a'"))
    (rule/'mutant_indirect_controls.a').write_text((rule/'indirect_controls.a').read_text().replace("'./rule.a'", "'./mutant_rule.a'").replace("'./numeric_constructions.a'", "'./mutant_tree.a'"))
    try:
        run([compiler,'build',str(rule/'mutant_indirect_controls.a'),'-o',str(out/label)], label+'-build')
        actual, elapsed = run([str(out/label)], label)
        assert actual != expected, label+' escaped'
        first = next((i for i,(a,b) in enumerate(zip(actual, expected)) if a != b), min(len(actual),len(expected)))
        print(label, 'compiled exit 0; Go byte comparison caught byte', first, flush=True)
    finally:
        for filename in ['mutant_tree.a','mutant_rule.a','mutant_indirect_controls.a']:
            (rule/filename).unlink()
