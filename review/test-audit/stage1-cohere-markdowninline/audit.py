import pathlib,subprocess,json,time,os,difflib,re
root=pathlib.Path('/workspace/adamic'); ev=pathlib.Path('/tmp/u131/evidence'); pkg='./stage1/cohere/markdowninline/'; files=['stage1/cohere/markdowninline/inline.ts','stage1/cohere/markdowninline/classes.ts','internal/native/runtime/regexp.c']; bases={f:(root/f).read_text() for f in files}; harness='stage1/cohere/markdowninline/inline_products_shards_test.go'; hb=(root/harness).read_text(); runs=[]
shards=[f'TestMarkdownInline_{i:04d}' for i in range(0,2048,64)]; alltests=[l.strip() for l in (ev/'list.log').read_text().splitlines() if l.startswith('Test')]; selected=[t for t in alltests if not re.fullmatch(r'TestMarkdownInline_\d{4}',t) or t in shards]; pattern='^('+'|'.join(selected)+')$'; (ev/'matrix-test-names.json').write_text(json.dumps(selected,indent=2)); products=[t for t in alltests if t.startswith('TestProduct_')]
rows={'delimiter':'^TestDelimiterExpressionMatchesNode$','products':'^TestProduct_','inline':'^('+'|'.join(shards+['TestMarkdownInlineUnion'])+')$','union':'^TestMarkdownInlineShardUnion$','selector':'^TestMarkdownInlineShardSelector$','sources':'^TestMarkdownInlineProductSourceInputs$','assignment':'^TestMarkdownInlineShardAssignmentStable$'}
def run(id,pat=pattern,extra=[],env={}):
 c=['timeout','120','go','test','-json','-count=1','-timeout','90s',*extra,pkg,'-run',pat]; environment={**os.environ,'ADAMIC_MARKDOWNINLINE_LIBRARY':'/tmp/u131/prettier/node_modules/prettier',**env}; start=time.monotonic()
 with (ev/(id+'.log')).open('w') as f:p=subprocess.run(c,cwd=root,env=environment,stdout=f,stderr=subprocess.STDOUT)
 rec={'id':id,'command':c,'environment':{k:v for k,v in environment.items() if k in ['ADAMIC_MARKDOWNINLINE_LIBRARY','ADAMIC_BUILD_CACHE_DIR','ADAMIC_AUDIT_WEAK']},'exit':p.returncode,'wall':time.monotonic()-start}; runs.append(rec); (ev/'runs.json').write_text(json.dumps(runs,indent=2)); print(id,p.returncode,round(rec['wall'],3),flush=True); return p.returncode
for row,pat in rows.items():
 for i in range(1,4):assert run(f'timing-{row}-{i}',pat)==0
mutants=[('M1',files[0],'preferred > alternate ? other : quote','preferred < alternate ? other : quote','flip condition'),('M2',files[1],'else if(code > end)','else if(code >= end)','off-by-one endpoint'),('M3',files[0],'marker === 61 || marker === 45','marker === 62 || marker === 45','change constant'),('M4',files[2],"next == '$' ? replacement : input","next != '$' ? replacement : input",'flip condition'),('P1',files[0],'export function formatLeaf(mode: string, text: string): string {','export function formatLeaf(mode: string, text: string): string { return \"\";','empty semantic entry probe'),('P2',files[2],'adamic_string *replacement, bool require_global) {','adamic_string *replacement, bool require_global) { return &adamic_string_empty;','empty RegExp replacement entry probe')]
catalog=[]
def diff(id,file,base,changed): (ev/(id+'.diff')).write_text(''.join(difflib.unified_diff(base.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
for id,f,old,new,kind in mutants:
 base=bases[f]; assert base.count(old)==1,(id,base.count(old)); changed=base.replace(old,new,1); diff(id,f,base,changed);catalog.append({'id':id,'file':f,'line':base[:base.index(old)].count('\n')+1,'before':old,'after':new,'kind':kind})
(ev/'catalog.json').write_text(json.dumps(catalog,indent=2))
(ev/'scope.md').write_text('Unit u131, origin/main d054e3578d4e53bef8cb8c9ab82d6ace69cdafeb, nproc=5. CODE UNDER TEST: TypeScript leaf printer and native RegExp replacement for the delimiter gap. ORACLES: protected Go cohere overlay and pinned Prettier 3.9.6 plugin, plus protected source Node and self-written construction assertions. Port functions: lineTerminator, rest, match, whitespace, canOpenOrClose, escapeDelimiterRuns, printWord, replaceCharacter, printInlineCode, printWikiLink, printURL, printTitle, printReference, printImage, formatLeaf, punctuation, driver decode/encode and top-level argument/batch/raw-output driver. The gap reaches regex_substitution via adamic_regex_replace, plus regex_piece, regex_program, regex_require_global, regex_input, regex_execute, regex_advance, regex_to_length and string/array ownership primitives. Full transitive runtime function coverage was not recorded. Function bodies for every mutated target were read whole. The source gap and Node oracle were not mutated.\nFixed production plan before mutant results: M1 title quote condition flip, M2 punctuation upper-bound off-by-one, M3 pseudo-setext constant change, M4 regex_substitution input/replacement condition flip. All mutants have separate native caches; M4 changes the runtime of every native product. P1 formatLeaf empty string; P2 adamic_regex_replace empty string. No probe is a production kill. Separate standalone rebuilds are used, at most four primary mutants.\nGrouping: all 2048 wrapper bodies were checked whole and are exactly the same checker with different ordinals; their corpus union joins the family. Their checker also carries raw-input and planted/built-in mutant witnesses, which are weakened separately. All 20 TestProduct wrappers share inlineBuild/buildcache.Product over different recipes and are grouped as one product family. Four distinct construction tests stay separate. The delimiter-expression test is separate. Whole package timed out at 90.049s. Matrix is bounded to the 32 ordinal-multiple-of-64 shards plus all 26 non-shard tests; matrix-test-names.json lists all 58. Kills outside that set are unknown. Pinned Prettier enabled; bounded baseline passed in 6.853s.\n')
try:
 for id,f,old,new,kind in mutants:
  (root/f).write_text(bases[f].replace(old,new,1)); run(id,env={'ADAMIC_BUILD_CACHE_DIR':'/tmp/u131/cache/'+id}); (root/f).write_text(bases[f])
finally:
 for f,base in bases.items():(root/f).write_text(base)
def body(text,name):
 start=text.index('func '+name+'('); brace=text.index('{',start); depth=1; i=brace+1
 # These target bodies contain string braces only in fmt text; use explicit next-function boundary below instead.
 end=text.index('\n}',brace)+2; return text[start:end]
s1=body(hb,'inlineUnion'); s1new=s1[:s1.index('{')+1]+'\n\treturn nil\n}'
s3=body(hb,'inlineProductSources'); s3new=s3[:s3.index('{')+1]+'\n\tt.Helper()\n\treturn in\n}'
s4=body(hb,'enumerateInlineShards'); s4new=s4.replace('key = "corpus:" + names[text]','key = fmt.Sprintf("corpus:%d", text)')
s5new=s4.replace('for mode := range len(inlineModes) {','for mode := range len(inlineModes)-1 {')
w1=body(hb,'inlineDifference'); w1new=w1[:w1.index('{')+1]+'\n\treturn nil\n}'
checks=[('S1',s1,s1new,'construction'),('S2','n < 1 || i < 0','n < 0 || i < 0','construction'),('S3',s3,s3new,'construction'),('S4',s4,s4new,'construction'),('S5',s4,s5new,'construction'),('W1',w1,w1new,'weakened witness comparison')]
for id,old,new,kind in checks:
 assert hb.count(old)==1,(id,hb.count(old)); changed=hb.replace(old,new,1);diff(id,harness,hb,changed);catalog.append({'id':id,'file':harness,'line':hb[:hb.index(old)].count('\n')+1,'before':old,'after':new,'kind':kind})
 target=ev/(id+'.go.txt');target.write_text(changed);overlay=ev/(id+'.overlay.json');overlay.write_text(json.dumps({'Replace':{str(root/harness):str(target)}}));run(id,extra=['-overlay='+str(overlay)])
 with (ev/(id+'-vet.log')).open('w') as f:p=subprocess.run(['go','vet','-overlay='+str(overlay),pkg],cwd=root,stdout=f,stderr=subprocess.STDOUT)
 runs.append({'id':id+'-vet','exit':p.returncode});(ev/'runs.json').write_text(json.dumps(runs,indent=2));assert p.returncode==0
(ev/'catalog.json').write_text(json.dumps(catalog,indent=2))
with (ev/'package-vet.log').open('w') as f:p=subprocess.run(['go','vet',pkg],cwd=root,stdout=f,stderr=subprocess.STDOUT)
assert p.returncode==0
