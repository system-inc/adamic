import pathlib,subprocess,time,json,os
root=pathlib.Path('/workspace/adamic'); pkg='stage1/cohere/lint/helpers/comments'; p=root/pkg; out=root/'review/test-audit/stage1-cohere-lint-helpers-comments';out.mkdir(parents=True,exist_ok=True)
base={str(f.relative_to(root)):f.read_text() for f in list(p.glob('*.ts'))+list(p.glob('*_test.go'))}
mut=[('M1','can_begin_at.ts','code > 13','code > 12','change constant'),('M2','all.ts','point > 65535 ? 2 : 1','point > 65535 ? 1 : 1','change constant'),('M3','sort_by_position.ts','let outer = 1;','let outer = 2;','off-by-one bound'),('M4','main.ts',"        panic('NotYet: stage-1 JSX parser adapter');\n",'','drop statement')]
(out/'plan.json').write_text(json.dumps({'base':subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip(),'code_under_test':'TypeScript comments port and module driver compiled natively','oracle':'live Go cohere; JSX exit 70 and NotYet label are self-written refusal contract','functions':['main module entry','read','all','Comment.constructor','FileCommentCache.constructor','FileCommentCache.get','forFile','canBeginAt','sortByPosition','tokenStart','isParameterListKind','isBracedListKind','collectListInteriors'],'menu':mut},indent=2))
meta=[]
def restore():
 for f,s in base.items(): (root/f).write_text(s)
def run(id,regex):
 log=out/(id+'.log');cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./'+pkg+'/','-run',regex];t=time.monotonic()
 with log.open('w') as fd:r=subprocess.run(cmd,cwd=root,stdout=fd,stderr=subprocess.STDOUT)
 meta.append({'id':id,'command':' '.join(cmd)+' > '+str(log)+' 2>&1','exit':r.returncode,'wall_seconds':time.monotonic()-t});(out/'runs.json').write_text(json.dumps(meta,indent=2));print(id,r.returncode,round(meta[-1]['wall_seconds'],2),flush=True)
def diff(id):
 (out/(id+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',pkg],cwd=root))
rows=['TestCommentMutants_[0-9]+','TestCommentMutantsUnion','TestCommentMutantsPlantedFailure','TestCommentsMatchCohere','TestConsumerCommentHelpers','TestJsxParserGapIsExplicit','TestJsxAdapterGuardMutant']
try:
 for row in rows:
  for n in range(1,4):run('timing-'+row+'-'+str(n),'^'+row+'$')
 # Bounded matrix excludes witnesses whose production failures are not proof.
 matrix='^(TestCommentsMatchCohere|TestConsumerCommentHelpers|TestJsxParserGapIsExplicit)$'
 for id,file,old,new,kind in mut:
  restore();f=p/file;s=f.read_text();assert s.count(old)==1,(id,s.count(old));f.write_text(s.replace(old,new));diff(id)
  # native build exercised by every positive row, and by the refusal row
  run(id,matrix)
 restore();f=p/'main.ts';s=f.read_text();f.write_text(s[:s.index('const args = programArguments();')]);diff('P1');run('P1',matrix)
 restore();f=p/'comment_mutants_test.go';s=f.read_text();f.write_text(s.replace('if bytes.Equal(got, want) {','if true {'));diff('W1');run('W1','^TestCommentMutants_[0-9]+$')
 restore();f=p/'comment_mutants_test.go';s=f.read_text();f.write_text(s.replace('if bytes.Equal(got, want) {','if false {'));diff('W2');run('W2','^TestCommentMutantsPlantedFailure$')
 restore();f=p/'comments_test.go';s=f.read_text();start=s.index('func matchesGapRefusal(');end=s.index('\n}',start)+2;f.write_text(s[:start]+'func matchesGapRefusal(err error, stderr string) bool { return true }'+s[end:]);diff('W3');run('W3','^TestJsxAdapterGuardMutant$')
 restore();f=p/'comment_mutants_test.go';s=f.read_text();old='for shard := 0; shard < testCommentMutantsShards; shard++ {';assert s.count(old)==2;f.write_text(s.replace(old,'for shard := 0; shard < testCommentMutantsShards-1; shard++ {',1));diff('S1');run('S1','^TestCommentMutantsUnion$')
 restore();f=p/'comment_mutants_test.go';s=f.read_text();start=s.index('func commentMutants(');body=s.index('{',start);# return type anonymous struct braces precede body; locate return statement instead
 pos=s.index('\n\treturn ',start);f.write_text(s[:pos]+s[pos:].replace('\n\treturn ', '\n\treturn nil\n\treturn ',1));diff('P2');run('P2','^TestCommentMutantsUnion$')
finally:restore()
