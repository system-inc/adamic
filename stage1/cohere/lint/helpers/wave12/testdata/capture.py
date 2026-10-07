"""Run all six consumer fixture families with actual Go helpers through overlays."""
import json, os, pathlib, subprocess, tempfile, itertools, random
HERE = pathlib.Path(__file__).resolve().parent
ROOT = HERE.parents[5]
COHERE = ROOT / 'cohere'
HELPER = 'github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.normalizeValueFunctionArgument'
ledger = json.loads((HERE.parents[1] / 'readiness.json').read_text())
consumers = {row['rule'] for row in ledger['remaining'] if HELPER in row['remaining_helpers']}
with tempfile.TemporaryDirectory(prefix='wave12-helper-capture-') as temporary:
    scratch = pathlib.Path(temporary)
    replacements = {}
    def overlay(original, content):
        side = scratch / ('side-' + str(len(replacements)) + '.go')
        side.write_text(content)
        replacements[str(original)] = str(side)
    utility = COHERE / 'internal/lint/rules/tailwind/collapse/utility_nodes.go'
    source = utility.read_text()
    anchor = 'func normalizeValueFunctionArgument(argument string) string {'
    assert source.count(anchor) == 1
    overlay(utility, source.replace(anchor, 'func adamicOriginalNormalizeArgument(argument string) string {'))
    overlay(utility.parent / 'adamic_capture.go', '''package tailwind
import("encoding/json";"os";"sync")
var adamicCaptureLock sync.Mutex
func normalizeValueFunctionArgument(argument string) string {
 result := adamicOriginalNormalizeArgument(argument)
 if path := os.Getenv("ADAMIC_HELPER_CAPTURE"); path != "" {
  adamicCaptureLock.Lock(); defer adamicCaptureLock.Unlock()
  file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); if err != nil {panic(err)}
  if err := json.NewEncoder(file).Encode(argument); err != nil {panic(err)}; file.Close()
 }
 return result
}
''')
    harness = COHERE / 'internal/lint/testing/rule_testing.go'
    source = harness.read_text()
    anchor = 'return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
    assert source.count(anchor) == 1
    overlay(harness, source.replace(anchor, 'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\nRecordAssertedCase(t, result)\nreturn result'))
    install = os.environ.get('ADAMIC_TAILWIND_ROOT', '/workspace/scratch/wave12-tailwind')
    assert (pathlib.Path(install) / 'node_modules/tailwindcss/index.css').is_file()
    for fixtures in (COHERE / 'internal/lint/rules/tailwind').glob('*_test.go'):
        source = fixtures.read_text()
        old = '"/Users/kirkouimet/Projects/ahra/app/_theme/styles"'
        if old in source:
            overlay(fixtures, source.replace(old, json.dumps(install)))
    mapping = scratch / 'overlay.json'; mapping.write_text(json.dumps({'Replace':replacements}))
    environment = os.environ | {'ADAMIC_HELPER_CAPTURE':str(scratch/'arguments.jsonl'), 'COHERE_DOCS_CAPTURE':str(scratch/'fixtures')}
    import re
    names = []
    for file in ['enforce_canonical_classes_test.go','enforce_consistent_class_order_test.go','enforce_consistent_variant_order_test.go','enforce_shorthand_classes_test.go','no_conflicting_classes_test.go','no_unknown_classes_test.go']:
        names.extend(re.findall(r'^func (Test\w+)\(', (COHERE/'internal/lint/rules/tailwind'/file).read_text(), re.MULTILINE))
    pattern = '^(' + '|'.join(names) + ')$' 
    with (HERE.parent/'evidence/consumers.log').open('w') as log:
        subprocess.run(['go','test','-overlay='+str(mapping),'./internal/lint/rules/tailwind','-run',pattern,'-count=1','-v','-timeout=15m'],cwd=COHERE,env=environment,stdout=log,stderr=subprocess.STDOUT,check=True)
    with (HERE.parent/'evidence/utility.log').open('w') as log:
        subprocess.run(['go','test','-overlay='+str(mapping),'./internal/lint/rules/tailwind/collapse','-run','^TestUtility','-count=1','-v','-timeout=15m'],cwd=COHERE,env=environment,stdout=log,stderr=subprocess.STDOUT,check=True)
    fixtures = {}
    for file in sorted((scratch/'fixtures').glob('*.jsonl')):
        for line in file.read_text().splitlines():
            row = json.loads(line)
            if row['rule'] in consumers:
                fixtures[(row['rule'],row['file'],row['source'])] = row
    seen = {row['rule'] for row in fixtures.values()}
    assert seen == consumers, sorted(consumers-seen)
    rows = [fixtures[key] for key in sorted(fixtures)]
    (HERE/'consumers.json').write_text(json.dumps(rows,ensure_ascii=True,indent=2)+'\n')
    arguments = sorted(set(json.loads(line) for line in (scratch/'arguments.jsonl').read_text().splitlines()))
    controls = ['', '--spacing', '--text --line-height', '--a --b --c', '--a\n--b', '--a\nX --b', '--a\r\n--b', r'--a\*', r'\\*', '-*-*-*', '--default(4)', '--a\v--b', '--a\u00a0--b', '--a\u2028--b', '😀 --名称 --二', '\ufeff--a']
    for scalar in range(256):
        c = chr(scalar)
        controls.extend([c, '--a'+c+'--b', '--a'+c+'x --b', '\\'+c+'-*', '--a'+c])
    for length in range(6):
        controls.extend(''.join(s) for s in itertools.product('-* \\\n',repeat=length))
    randomizer=random.Random(1212)
    for _ in range(4000):
        controls.append(''.join(randomizer.choice(['--','-*','\\*',' ','\t','\r','\n','\f','\v','x','(',')','😀','名称']) for _ in range(randomizer.randrange(30))))
    inputs = arguments + [row['source'] for row in rows] + controls
    (HERE/'cases.json').write_text(json.dumps(inputs,ensure_ascii=True,indent=2)+'\n')
    metadata={'helper':HELPER,'consumer_rules':sorted(consumers),'fixtures':len(rows),'observed_distinct_helper_arguments':len(arguments),'controls':len(controls),'cases':len(inputs),'final_blockers_removed':0}
    (HERE/'coverage.json').write_text(json.dumps(metadata,indent=2)+'\n')
    print(json.dumps(metadata,indent=2))
