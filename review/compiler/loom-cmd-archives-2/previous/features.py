import json, os, pathlib, re, subprocess
out=pathlib.Path('review/compiler/cmd-adamic-archives')
paths=['internal/lower/testdata/predicates/'+x+'.a' for x in ['overload_some_empty','overload_some_false_read','overload_some_true_only','overload_erased','overload_assertion']]
paths+=['internal/oracle/testdata/non_null_'+('' if x.startswith('refuse_') else 'checked_')+x+'.ts' for x in ['narrowed','present','undefined','null','initializer','let_initializer','refuse_undefined','refuse_null']]
paths+=['cmd/adamic/testdata/wasi/request.a','cmd/adamic/testdata/wasi/request_alias.a']
rows=[]
for i,path in enumerate(paths):
    env=dict(os.environ, ADAMIC_EXPLAIN_TEST_DRIVER='1')
    with (out/f'features-{i}.c.txt').open('w') as source, (out/f'features-{i}.log').open('w') as log:
        subprocess.run(['timeout','55','/tmp/adamic-before.test','-test.run=^TestExplainChecksDriver$','--','c',path],env=env,stdout=source,stderr=log,check=True)
    rows.append(dict(path=path,features=re.findall(r'^#define (ADAMIC_\w+) 1$',(out/f'features-{i}.c.txt').read_text(),re.M)))
(out/'features.json').write_text(json.dumps(rows,indent=2)+'\n')
print(json.dumps(rows,indent=2))
