import pathlib,json,subprocess,os,time,difflib
D=pathlib.Path('review/test-audit/internal-regexp-matcher_oracle');family=['TestMatcherNodeControls','TestMatcherRandomNode','TestMatcherTest262Executions','TestMatcherCanonicalizeNode','TestMatcherUTF16PatternsNode','TestMatcherOct6Node'];regex='^('+'|'.join(family)+')$';commands=[]
def run(label,regex):
 args=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/regexp/','-run',regex];t=time.monotonic()
 with (D/(label+'.log')).open('w') as out:rc=subprocess.call(args,stdout=out,stderr=subprocess.STDOUT)
 commands.append(dict(label=label,command=' '.join(args),wall_seconds=round(time.monotonic()-t,3),exit=rc));(D/'finish-commands.json').write_text(json.dumps(commands,indent=2))
for i in range(3):run('time-family-'+str(i+1),regex)
# Empty compilation probes run each caller alone because nil programs panic.
probes=[('PCompile','matcher.go','func Compile(pattern, flags string) (*Program, error) {',['TestMatcherNodeControls','TestMatcherRandomNode','TestMatcherTest262Executions','TestMatcherStepLimit','TestMatcherCanonicalizeNode','TestMatcherStepLimitBoundary','TestMatcherOct6Node','TestMatcherOct6LoopsNode']),('PCompileUTF16','matcher.go','func CompileUTF16(pattern []uint16, flags string) (*Program, error) {',['TestMatcherUTF16PatternsNode']),('PCompileProperties','matcher.go','func CompileWithProperties(pattern, flags string, properties PropertyProvider) (*Program, error) {',['TestMatcherPropertyProviderStrings','TestMatcherProviderSnapshot'])]
for id,file,entry,rows in probes:
 p=pathlib.Path('internal/regexp')/file;s=p.read_text();new=s.replace(entry,entry+'\nif auditProbe("'+id+'") {return nil,nil}')
 helper=pathlib.Path('internal/regexp/audit_probe.go');helper.write_text('package regexp\nimport "os"\nfunc auditProbe(id string) bool{return os.Getenv("ADAMIC_MUTANT")==id}\n')
 # Standalone entry probe includes selector solely to keep vet from rejecting unreachable code.
 diff=''.join(difflib.unified_diff(s.splitlines(True),new.splitlines(True),fromfile='a/'+str(p),tofile='b/'+str(p)))+''.join(difflib.unified_diff([],helper.read_text().splitlines(True),fromfile='/dev/null',tofile='b/'+str(helper)))
 (D/(id+'.diff')).write_text(diff);p.write_text(new)
 os.environ['ADAMIC_MUTANT']=id
 for row in rows:run(id+'-'+row,'^'+row+'$')
 p.write_text(s);helper.unlink();del os.environ['ADAMIC_MUTANT']
run('restored-baseline','.')
