#!/usr/bin/env python3
"""Run independent proof mutants through scratch Go overlays; never edit production."""
import json, subprocess, sys, tempfile
from pathlib import Path
repo = Path(__file__).resolve().parents[4]
root = Path(tempfile.mkdtemp(prefix='adamic-delegation-mutants-'))
proof = repo / 'internal/lower/predicates_proof.go'
delegated = repo / 'internal/lower/predicates_delegated.go'
seed = '''for _,declaration := range declarations {
 if declaration == node.Parent || declaration.Type()==nil || declaration.Type().Kind!=ast.KindTypePredicate {continue}
 target:=declaration.Type().AsTypePredicateNode().Type;if target==nil {continue}
 values:=make([]predicateTruth,len(v.cells));valid:=true
 for index,cell:=range v.cells {yes,known:=v.wanted(l.checker.GetTypeAtLocation(target),cell);if !known {valid=false;break};if yes {values[index]=predicateTrue}else{values[index]=predicateFalse}}
 if valid {v.summaries[declaration]=values}
}
'''
mutants = [
 ('helper-annotation', proof, '\tvar last error', seed+'\tvar last error', 'TestDelegatedProofLyingHelper'),
 ('seedless-cycle', proof, '\tvar last error', seed+'\tvar last error', 'TestDelegatedProofSeedlessCycle'),
 ('negation-direction', proof, 'return (t&predicateTrue)<<1 | (t&predicateFalse)>>1', 'return t', 'TestDelegatedProofNegation'),
 ('ignore-effect', delegated, 'call := node.AsCallExpression()', 'call := node.AsCallExpression()\n if ast.IsIdentifier(call.Expression) && call.Expression.Text()=="change" {return predicateEither,nil}', 'TestDelegatedProofEffect'),
]
for name, path, old, replacement, test in mutants:
    source = path.read_text()
    assert source.count(old) == 1, name
    copy = root / (name+'.go')
    copy.write_text(source.replace(old,replacement,1))
    overlay = root / (name+'.json')
    overlay.write_text(json.dumps({'Replace': {str(path):str(copy)}}))
    log = root / (name+'.log')
    with log.open('w') as output:
        result = subprocess.run(['go','test','-overlay',str(overlay),'./internal/lower','-run','^'+test+'$','-count=1','-v'],cwd=repo,stdout=output,stderr=subprocess.STDOUT)
    text = log.read_text()
    assert result.returncode != 0 and '--- FAIL: '+test in text and '[build failed]' not in text, (name,text)
    print(name+': caught by '+test+'; '+str(log),flush=True)
print('all four proof mutants caught; logs '+str(root))
