"""Capture real Go normalization calls and all four consumer fixture families."""
import itertools, json, os, pathlib, random, re, subprocess, tempfile
HERE = pathlib.Path(__file__).resolve().parent
COHERE = HERE.parents[5] / 'cohere'
HELPER = 'github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph.normalizeBigIntLiteral'
CONSUMERS = ['array-callback-return', 'consistent-return', 'no-unreachable-loop', 'react-hooks/rules-of-hooks']
with tempfile.TemporaryDirectory(prefix='wave12-cfg-capture-') as temporary:
    scratch = pathlib.Path(temporary)
    replacements = {}
    def overlay(original, content):
        side = scratch / ('side-' + str(len(replacements)) + '.go')
        side.write_text(content)
        replacements[str(original)] = str(side)
    helper = COHERE / 'internal/lint/ecmascript/control_flow_graph/helpers.go'
    source = helper.read_text()
    anchor = 'func normalizeBigIntLiteral(text string) string {'
    assert source.count(anchor) == 1
    overlay(helper, source.replace(anchor, 'func adamicOriginalNormalizeBigIntLiteral(text string) string {'))
    overlay(helper.parent / 'adamic_capture.go', r'''package control_flow_graph
import("encoding/json";"os";"sync")
var adamicCaptureLock sync.Mutex
func normalizeBigIntLiteral(text string) string {
 result:=adamicOriginalNormalizeBigIntLiteral(text)
 if path:=os.Getenv("ADAMIC_CFG_CAPTURE");path!="" {
  adamicCaptureLock.Lock();defer adamicCaptureLock.Unlock()
  file,err:=os.OpenFile(path,os.O_CREATE|os.O_WRONLY|os.O_APPEND,0644);if err!=nil{panic(err)}
  if err:=json.NewEncoder(file).Encode(text);err!=nil{panic(err)};file.Close()
 }
 return result
}
''')
    harness = COHERE / 'internal/lint/testing/rule_testing.go'
    source = harness.read_text()
    anchor = 'return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
    assert source.count(anchor) == 1
    overlay(harness, source.replace(anchor, 'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\nRecordAssertedCase(t, result)\nreturn result'))
    mapping = scratch/'overlay.json';mapping.write_text(json.dumps({'Replace':replacements}))
    env = os.environ | {'ADAMIC_CFG_CAPTURE':str(scratch/'arguments.jsonl'),'COHERE_DOCS_CAPTURE':str(scratch/'fixtures')}
    for package, files in [('core',['array_callback_return_test.go','consistent_return_test.go','no_unreachable_loop_test.go']),('react',['rules_of_hooks_test.go'])]:
        names=[]
        for file in files:
            names.extend(re.findall(r'^func (Test\w+)\(', (COHERE/'internal/lint/rules'/package/file).read_text(), re.MULTILINE))
        assert names
        with (HERE.parent/('evidence/cfg-consumers-'+package+'.log')).open('w') as log:
            subprocess.run(['go','test','-overlay='+str(mapping),'./internal/lint/rules/'+package,'-run','^('+'|'.join(names)+')$','-count=1','-v','-timeout=15m'],cwd=COHERE,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
    fixtures={}
    for file in sorted((scratch/'fixtures').glob('*.jsonl')):
        for line in file.read_text().splitlines():
            row=json.loads(line)
            if row['rule'] in CONSUMERS:fixtures[(row['rule'],row['file'],row['source'])]=row
    rows=[fixtures[key] for key in sorted(fixtures)]
    assert set(row['rule'] for row in rows)==set(CONSUMERS)
    arguments_path=scratch/'arguments.jsonl'
    arguments=sorted(set(json.loads(line) for line in arguments_path.read_text().splitlines())) if arguments_path.exists() else []
    controls=['','n','nn','0n','-0n','+0n','0x0n','0b0n','0o0n','0X_FFn','0_7n','077n','08n','09n','-0x10000000000000000n','0x10000000000000000n','0b'+('1'*257)+'n','0x'+('a'*257)+'n','9'*1024+'n',' 1n','1 n','１n','😀n']
    for base,prefix,alphabet in [(2,'0b','01'),(8,'0o','01234567'),(10,'','0123456789'),(16,'0x','0123456789abcdef')]:
        for value in [0,1,7,8,15,16,2**53-1,2**53,2**64-1,2**64,2**127,2**256-1]:
            spelling = (bin(value)[2:] if base==2 else oct(value)[2:] if base==8 else str(value) if base==10 else hex(value)[2:])
            for sign in ['','+','-']:
                for content in [spelling,'_'.join(spelling),'_'+spelling,spelling+'_',spelling+'__0']:
                    for suffix in ['','n','nn']:controls.append(sign+prefix+content+suffix)
    for length in range(5):controls.extend(''.join(chars) for chars in itertools.product('01n_x+- ',repeat=length))
    rng=random.Random(1214)
    for _ in range(2000):
        prefix=rng.choice(['','0','0x','0X','0b','0B','0o','0O']); sign=rng.choice(['','+','-'])
        controls.append(sign+prefix+''.join(rng.choice('0123456789abcdef_') for _ in range(rng.randrange(100)))+rng.choice(['','n']))
    inputs=arguments+[row['source'] for row in rows]+controls
    (HERE/'cfg_consumers.json').write_text(json.dumps(rows,ensure_ascii=True,indent=2)+'\n')
    (HERE/'bigint_cases.json').write_text(json.dumps(inputs,ensure_ascii=True,indent=2)+'\n')
    metadata={'helper':HELPER,'consumer_rules':CONSUMERS,'fixtures':len(rows),'observed_distinct_helper_arguments':len(arguments),'controls':len(controls),'cases':len(inputs),'final_blockers_removed':0}
    (HERE/'bigint_coverage.json').write_text(json.dumps(metadata,indent=2)+'\n')
    print(json.dumps(metadata,indent=2))
