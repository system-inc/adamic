import pathlib,json
p=pathlib.Path(__file__).resolve().parent
names=[s.strip() for s in (p/'list.log').read_text().splitlines() if s.startswith('Test')]
def read(log):
    return [json.loads(s) for s in log.read_text().splitlines() if s.startswith('{')]
def statuses(label):
    result={n:'unknown' for n in names}
    for r in json.loads((p/(label+'-runs.json')).read_text()):
        for e in read(p/(label+'-'+str(r['group'])+'.log')):
            n=e.get('Test')
            if n in result and e.get('Action') in ['pass','fail','skip']: result[n]=e['Action']
    logs=[p/'clean-performance-enabled.log'] if label=='clean' else [p/('D1-'+n+'-enabled.log') for n in ['TestPerformance','TestWholePerformance']]
    for log in logs:
        for e in read(log):
            n=e.get('Test')
            if n in result and e.get('Action') in ['pass','fail','skip']: result[n]=e['Action']
    return result
clean=statuses('clean');actual=statuses('D1')
assert set(clean.values())=={'pass'},clean
assert actual['TestWholeGeneratedAgrees']=='fail',actual
assert all(s=='pass' for n,s in actual.items() if n!='TestWholeGeneratedAgrees'),actual
failure=next(e['Output'].strip() for e in read(p/'D1-14.log') if 'whole_generated_test.go:66:' in e.get('Output',''))
passed=[n for n,s in actual.items() if s=='pass']
(p/'matrix.json').write_text(json.dumps({'base':json.loads((p/'metadata.json').read_text())['base'],'mutant':'D1','bounded':True,'all_current_top_level_tests_accounted_for':True,'clean':clean,'D1':actual,'rows_failed':['TestWholeGeneratedAgrees'],'rows_passed':passed,'unknown':[],'performance_enabled':True,'failure':failure},indent=2)+'\n')
rows=[{'test':'TestProduct_ParserOracle','package':'stage1/typescript/parser','prior_verdict':'untrue','subsumed_by':[],'defense':'cannot-judge','unique_mutant':None,'attempts':[],'evidence':"go test -run '^TestProduct_ParserOracle$' -coverpkg=github.com/system-inc/adamic/internal/buildcache -coverprofile=<test>.cold.cover; identical cold coverage to TestProduct_ExpressionsOracle (57 covered blocks, zero exclusive). No allowed production mutation target: oracle build recipe is _test.go harness; wrapper discards artifact path."},{'test':'TestWholeGeneratedAgrees','package':'stage1/typescript/parser','prior_verdict':'subsumed','subsumed_by':['TestEveryTypeNodeKindAgrees'],'defense':'defended','unique_mutant':'D1 stage1/typescript/parser/statements.ts:748','attempts':[{'mutant':'D1','file_line':'stage1/typescript/parser/statements.ts:748','change':"WithStatement constant -> WhileStatement in Statements.statement's WithKeyword branch",'rows_failed':['TestWholeGeneratedAgrees']}],'evidence':'python3 review/test-defend/stage1-typescript-parser-setup_deadline_products/run-matrix.py D1 (15 bounded timeout/go-test groups); '+failure}]
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
(p/'rows-passed.txt').write_text('\n'.join(passed)+'\n')
print(json.dumps({'failed':['TestWholeGeneratedAgrees'],'passed':len(passed),'failure':failure},indent=2))
