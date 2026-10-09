import pathlib,json,subprocess,gzip,re
p=pathlib.Path('review/test-defend/internal-native-radix');scope=json.loads((p/'scope.json').read_text())
def covered(label):
 out=set()
 for s in (p/'coverage'/(label+'.out')).read_text().splitlines()[1:]:
  span,_,count=s.split();file,bounds=span.rsplit(':',1);start,end=bounds.split(',')
  if int(count):out.update((file,n) for n in range(int(start.split('.')[0]),int(end.split('.')[0])+1))
 return out
sub=covered('TestRegExpSearchNode');out=[]
for label in ['TestRegExpBytecodeTest262','TestRegExpBytecodeRandomNode-family']:
 exclusive=sorted(covered(label)-sub);out.append({'test':label,'comparison':'TestRegExpSearchNode','exclusive_go_lines':[f'{file}:{line}' for file,line in exclusive],'c_coverage':None})
(p/'coverage-exclusive.json').write_text(json.dumps(out,indent=2))
cs=json.load(gzip.open('internal/regexp/testdata/matches.json.gz'));zero=[]
for index,c in enumerate(cs):
 for m in re.finditer(r'\{(\d+)(?:,(\d*))?\}',c['pattern']):
  bound=m[1] if m[2] is None else m[2]
  if bound and int(bound)==0:zero.append(index)
samples=[{'case_index':i,**c} for i,c in enumerate(cs) if c['pattern']=='^.$' and c['flags']=='mu' and c['input']==[0x2028]]
cmd=['node','-e',"const r=/^.$/mu;const m=r.exec(String.fromCharCode(0x2028));console.log(JSON.stringify({captures:m?Array.from(m.indices??[],x=>x):null,lastIndex:r.lastIndex}));"]
answer=subprocess.check_output(cmd,text=True)
(p/'input-distinction.json').write_text(json.dumps({'test262_cases':len(cs),'test262_u2028_cases':sum(0x2028 in c['input'] for c in cs),'zero_closed_bounds_found_in_recorded_patterns':zero,'test262_witness_cases':samples,'authority_command':cmd,'authority_output':answer,'random_generator_quantifiers':['','*','+','?','{0}','{1,}','{2,4}','*?','+?','??'],'random_generator_input_runes':'abc 12\\n\\réKſ🌍','search_subsumer_zero_quantifier':False,'search_subsumer_u2028_input':False},indent=2))
old=gzip.decompress(subprocess.check_output(['git','show','origin/test-audit/internal-native-radix:review/test-audit/internal-native-radix/u052-list.log.gz'])).decode().splitlines();old=[n for n in old if n.startswith('Test')]
scope['added_since_audit']=sorted(set(scope['all_package_tests'])-set(old));scope['removed_since_audit']=sorted(set(old)-set(scope['all_package_tests']));scope['missing_family_members']=[f'TestRegExpBytecodeRandomNodeUnit{i:02}' for i in range(40) if f'TestRegExpBytecodeRandomNodeUnit{i:02}' not in scope['all_package_tests']];(p/'scope.json').write_text(json.dumps(scope,indent=2));print('coverage exclusive counts',[(r['test'],len(r['exclusive_go_lines'])) for r in out]);print('added',scope['added_since_audit']);print('witness',samples,answer)
