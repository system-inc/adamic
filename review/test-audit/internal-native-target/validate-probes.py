import pathlib,json,difflib,subprocess,time
p=pathlib.Path('review/test-audit/internal-native-target');probes=json.loads((p/'probe-plan.json').read_text());helper=pathlib.Path('internal/native/audit_switch.go');gohelper='package native\nimport "os"\nfunc auditSelected(id string) bool {return os.Getenv("ADAMIC_MUTANT")==id}\n';chelp='\n#include <string.h>\nstatic bool audit_selected(const char *id) {const char *selected=getenv("ADAMIC_MUTANT");return selected != NULL && strcmp(selected,id)==0;}\n';flags=['-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all'];vals=[]
def diff(f,old,new):return ''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+f,tofile='b/'+f))
for q in probes+[dict(id='W01',file='internal/native/tsgo_features_test.go',signature='compile := func(flags []string) error {',empty='return nil')]:
 id=q['id'];f=q['file'];path=pathlib.Path(f);old=path.read_text();sig=q['signature'];ret=q['empty'];assert sig in old,id
 if f.endswith('.go'):
  new=old.replace(sig,sig+'\n if auditSelected("'+id+'") { '+ret+' }',1);helper.write_text(gohelper);text=diff(f,old,new)+''.join(difflib.unified_diff([],gohelper.splitlines(True),fromfile='/dev/null',tofile='b/'+str(helper)));cmd=['go','vet','./internal/native/']
 else:
  new=old.replace(sig,sig+' if (audit_selected("'+id+'")) { '+ret+' }',1).replace('#include <stdlib.h>','#include <stdlib.h>'+chelp,1);text=diff(f,old,new);cmd=['clang',*flags,'-I','internal/native/runtime','-c',f,'-o','/tmp/u054/'+id+'.o']
 (p/'diffs'/(id+'.diff')).write_text(text);path.write_text(new);start=time.monotonic()
 with (p/('validate-'+id+'.log')).open('w') as out:r=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT)
 path.write_text(old)
 if helper.exists():helper.unlink()
 vals.append(dict(id=id,status=r.returncode,seconds=time.monotonic()-start,command=' '.join(cmd),file=f,line=old[:old.index(sig)].count('\n')+1));(p/'probe-validation.json').write_text(json.dumps(vals,indent=2));assert r.returncode==0,id
for f in (p/'diffs').glob('*.diff'):subprocess.run(['git','apply','--check',str(f)],check=True,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
print('validated',len(vals),'probes/witness; all diffs apply')
