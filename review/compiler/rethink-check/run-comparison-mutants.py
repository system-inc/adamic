from pathlib import Path
import json,subprocess
p=Path('review/compiler/rethink-check'); rows=json.loads((p/'results.json').read_text()); r=next(x for x in rows if x['classification']=='FIXED'); n=r['name']; original=Path('/tmp/rethink-check')/(n+'.mjs'); expected=(r['node']['exit'],r['node']['stdout'],r['node']['stderr']); results=[]
for name,change in [('stdout',"process.stdout.write('!');"),('stderr',"process.stderr.write('!');"),('exit','process.exit(1);')]:
 source=original.read_text()+'\n'+change+'\n'; (p/('mutant-'+name+'.mjs.txt')).write_text(source); target=Path('/tmp/rethink-check')/('mutant-'+name+'.mjs'); target.write_text(source)
 with (p/('mutant-'+name+'.out.log')).open('wb') as out,(p/('mutant-'+name+'.err.log')).open('wb') as err:
  code=subprocess.run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(target)],stdout=out,stderr=err,timeout=15).returncode
 observed=(code,(p/('mutant-'+name+'.out.log')).read_text(),(p/('mutant-'+name+'.err.log')).read_text()); changed=[label for label,a,b in zip(['exit','stdout','stderr'],expected,observed) if a!=b]; assert changed==[name],(name,changed)
 results.append(dict(mutant=name,control=n,expected=expected,observed=observed,caught_by=changed)); print(name,'caught only by',name)
(p/'comparison-mutants-executed.json').write_text(json.dumps(results,indent=2)+'\n')
