import json, pathlib, subprocess
root = pathlib.Path.cwd()
evidence = root / 'review/compiler/after-chain-fix-fresh'
env = 'source /workspace/adamic-tools/env.sh; '
results = []
def run(name, command, catcher):
    log = evidence / (name + '.log')
    with log.open('w') as output:
        result = subprocess.run(['bash', '-lc', env + 'timeout 90s ' + command], stdout=output, stderr=subprocess.STDOUT, timeout=95)
    text = log.read_text()
    caught = result.returncode == 1 and catcher in text
    results.append(dict(mutant=name, exit=result.returncode, catcher=catcher, caught=caught, command=command))
    print(name, result.returncode, 'caught' if caught else 'FAILED', flush=True)
def overlay(name, file, before, after):
    path = root / file
    source = path.read_text()
    assert before in source
    changed = evidence / (name + '.go.txt')
    changed.write_text(source.replace(before, after))
    description = evidence / (name + '.json')
    description.write_text(json.dumps({'Replace': {str(path): str(changed)}}))
    return '-overlay=' + str(description)
site = overlay('missing-source-site', 'internal/lower/class.go', 'Site: l.writeSite(target.AsPropertyAccessExpression().Expression)', 'Site: 0')
run('missing-source-site', f"go test {site} ./internal/fresh -run 'TestFreshCorpusRemainder/../oracle/testdata/step21_saved_error.a' -count=1 -timeout 90s", "a write lowering didn't record")
push = overlay('drop-second-push', 'internal/lower/array_call_arguments.go', 'Elements: arguments, Spread: spread', 'Elements: arguments[:1], Spread: spread')
run('markdown-drop-second-push', f"go test {push} ./stage1/cohere/markdownblocks -run '^TestParserRepresentationProbes$/^gaps$/^10_multiple_push.ts$' -count=1 -timeout 90s", 'first byte difference')
run('scanner-drop-second-push', f"go test {push} ./stage1/typescript/scanner -run '^TestGapStandsWhereGapsMdSays$' -count=1 -timeout 90s", 'native push:')
def temporary(name, file, before, after, command, catcher):
    path = root / file
    original = path.read_text()
    assert original.count(before) == 1
    (evidence / (name + '.patch')).write_text(f'{file}\n- {before}\n+ {after}\n')
    try:
        path.write_text(original.replace(before, after))
        run(name, command, catcher)
    finally:
        path.write_text(original)
temporary('wrong-supplementary-offset', 'stage1/typescript/scanner/main.ts', 'offsets.push(bytes, bytes + 4);', 'offsets.push(bytes, bytes + 3);', "go test ./stage1/typescript/scanner -run '^TestGapStandsWhereGapsMdSays$' -count=1 -timeout 90s", 'Go')
temporary('missing-directory-unit', 'stage3/fixtures/lowering_chain_units_test.go', 'func TestFixturesSelfCompare(', 'func DisabledFixturesSelfCompare(', "go test ./stage3/fixtures -run '^TestFixtureDirectoriesHaveTopLevelTests$' -count=1 -timeout 90s", 'fixture directory self-compare has no top-level test')
temporary('wrong-self-comparison', 'stage3/fixtures/self-compare/closure.a', 'String(f === f)', 'String(f !== f)', "go test ./stage3/fixtures -run '^TestFixturesSelfCompare$/^self-compare$/^closure.a$' -count=1 -timeout 90s", 'recorded Node byte comparison failed')
(evidence / 'mutants.json').write_text(json.dumps(results, indent=2) + '\n')
assert all(result['caught'] for result in results), results
