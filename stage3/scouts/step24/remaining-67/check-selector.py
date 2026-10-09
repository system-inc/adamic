#!/usr/bin/env python3
"""Validate the exact seven nullable-owner edits and its wrong-contract mutant."""
import json,sys
from pathlib import Path
u=Path(__file__).resolve().parent
wanted={('src/compiler/types.ts','aliasesToMakeVisible'),('src/compiler/types.ts','moduleResolverHost'),('src/compiler/checker.ts','questionToken'),('src/compiler/checker.ts','modifiers'),('src/compiler/checker.ts','errors'),('src/compiler/checker.ts','value'),('src/compiler/resolutionCache.ts','nonRecursive')}
def check(data):
 actual={(d['file'],d['name']) for d in data['declarations']}
 assert len(data['declarations'])==7 and actual==wanted,('nullable owner ledger mismatch',actual,wanted)
check(json.loads((u/'nullable-experiment.json').read_text()))
mutant=json.loads(Path(sys.argv[1]).read_text())
try:check(mutant)
except AssertionError as error:
 assert ('src/compiler/checker.ts','questionToken') not in {(d['file'],d['name']) for d in mutant['declarations']}
 (u/'nullable-mutant.json').write_text(json.dumps({'caught':'exact nullable owner ledger','message':str(error),'declarationsChanged':mutant['declarations_changed'],'mutation':'SignatureToSignatureDeclarationOptions.questionToken?: number, on isolated real source','missingExpectedOwner':'questionToken'},indent=2)+'\n')
else:raise AssertionError('wrong-contract mutant survived')
print('seven exact nullable owners; wrong-contract mutant caught')
