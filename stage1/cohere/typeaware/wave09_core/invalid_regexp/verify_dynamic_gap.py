"""Prove the requested dynamic RegExp boundary and its refusal check."""
import pathlib,subprocess,json
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[4]
work=pathlib.Path('/workspace/wave-09-dynamic-gap');work.mkdir(exist_ok=True)
def run(name,args):
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:
  result=subprocess.run(list(map(str,args)),cwd=repo,stdout=out,stderr=err,timeout=120)
 print(name,result.returncode,flush=True)
 return result.returncode,(work/(name+'.stdout')).read_bytes(),(work/(name+'.stderr')).read_bytes()
probe=root/'dynamic_regexp_gap.a'
code,out,err=run('source',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',probe])
assert (code,out,err)==(0,b'true\n',b'')
code,out,err=run('native',['/workspace/wave-09-core/adamic','build',probe,'-o',work/'native'])
assert code!=0 and b'RegExp with a nonconstant pattern' in err
text=probe.read_text();assert text.count("new RegExp(pattern,'u')")==1
mutant=work/'mutant.a';mutant.write_text(text.replace("new RegExp(pattern,'u')","new RegExp('a','u')"))
code,out,err=run('mutant-build',['/workspace/wave-09-core/adamic','build',mutant,'-o',work/'mutant','--sanitize'])
assert code==0 and err==b''
code,out,err=run('mutant',[work/'mutant'])
assert (code,out,err)==(0,b'true\n',b'')
(work/'results.json').write_text(json.dumps({'source':'PASS','native':'REFUSED nonconstant pattern','mutant':'successful static literal removes required refusal','full_rule_parity':False},indent=2)+'\n')
print('PASS explicit dynamic constructor blocker, Node truth and clean sanitized static-pattern mutant; full rule remains blocked',flush=True)
