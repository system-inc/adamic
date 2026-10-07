import importlib.util,json,re
from pathlib import Path
base=Path('/workspace/scratch/stack-check-scc/cache')
spec=importlib.util.spec_from_file_location('cache_summary',base/'summarize.py')
module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
results={}
mutants=[]
for name in ['parse-before','parse-natural','parse-after','sieve-before','sieve-after']:
    path=base/(name+'.callgrind')
    result=module.summarize(path)
    results[name]={k:result[k] for k in ['events','summary','self_totals','summary_minus_self']}
    source=path.read_text()
    footer=re.search(r'^totals: (.*)$',source,re.M)
    original=list(map(int,footer[1].split()))
    for index,event in enumerate(result['events']):
        changed=original.copy();changed[index]+=1
        mutant=source[:footer.start(1)]+' '.join(map(str,changed))+source[footer.end(1):]
        candidate=base/'event-mutant.callgrind'
        candidate.write_text(mutant)
        try:
            module.summarize(candidate)
        except ValueError as error:
            mutants.append({'profile':name,'event':event,'error':str(error),'original':original[index],'mutant':changed[index]})
        else:
            raise RuntimeError(name+' '+event+' mutant escaped')
    stdout=(base/(name+'.stdout')).read_bytes()
    want=Path('/workspace/scratch/stack-check-scc/parse-before.stdout').read_bytes() if name.startswith('parse') else Path('/workspace/scratch/stack-check-scc/handler/before-check.stdout').read_bytes()
    if stdout!=want:
        raise RuntimeError(name+' output differs')
    print(name,json.dumps(result['self_totals']))
(base/'metrics.json').write_text(json.dumps(results,indent=2)+'\n')
(base/'event-mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
print(len(mutants),'footer event mutants caught')
