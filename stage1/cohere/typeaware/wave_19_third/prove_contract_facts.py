"""Compile raw-fact mutants under Go overlays; checker snapshots must reject each."""
import json, pathlib, subprocess, sys
root=pathlib.Path(__file__).resolve().parents[4]
work=pathlib.Path(sys.argv[1]);work.mkdir(parents=True,exist_ok=True)
changes=[('signature','wave19_type_signatures.go','result := g.add(c.GetReturnTypeOfSignature(signature))','result := g.add(checker.Checker_numberType(c))'),('generic','wave19_generic_call.go','parameters = append(parameters, g.add(t))','_ = t'),('member','wave19_type_members.go','property = g.add(checker.Checker_getTypeOfSymbol(c, symbol))','_ = symbol'),('heritage','wave19_heritage_members.go','roots = append(roots, g.add(c.GetTypeOfSymbolAtLocation(property, node)))','_ = property')]
for name,file,before,after in changes:
    original=root/'bridge/tsgo/checker'/file
    source=original.read_text();assert source.count(before)==1
    mutant=work/(name+'.go');mutant.write_text(source.replace(before,after,1))
    overlay=work/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(original):str(mutant)}}))
    log=work/(name+'.log')
    with log.open('wb') as out:
        result=subprocess.run(['go','test','-overlay='+str(overlay),'./bridge/tsgo/checker','-run','^TestWave19ContractFacts$','-count=1','-v'],cwd=root,stdout=out,stderr=subprocess.STDOUT)
    text=log.read_text();assert result.returncode!=0 and 'wrong ' in text and 'build failed' not in text,(name,text)
    print(name+' mutant compiled; direct checker snapshot rejected incorrect facts')
