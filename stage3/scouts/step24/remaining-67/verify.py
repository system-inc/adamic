#!/usr/bin/env python3
"""Compare full emitted artifacts byte for byte, with a real byte mutant."""
import hashlib,json,sys
from pathlib import Path
unit=Path(__file__).resolve().parent
before,after=map(Path,sys.argv[1:3])
def artifacts(tree):
 base=tree/'built/local'
 return {str(p.relative_to(base)):p.read_bytes() for p in base.rglob("*") if p.is_file() and (p.suffix=='.js' or p.suffix=='.ts' and p.name.endswith('.d.ts'))}
b,a=artifacts(before),artifacts(after)
assert b and b.keys()==a.keys(),(list(b),list(a))
internalNames={n for n in b if n.endswith('.internal.d.ts') or '/' in n and n.endswith('.d.ts')}
allowed={'typescript.internal.d.ts','compiler/types.d.ts','compiler/resolutionCache.d.ts','compiler/builder.d.ts'}
assert {n for n in b if b[n]!=a[n]}==allowed, 'unexpected internal declaration delta'
internal = {name: {'beforeSha256':hashlib.sha256(b[name]).hexdigest(),'afterSha256':hashlib.sha256(a[name]).hexdigest(),'identical':b[name]==a[name]} for name in sorted(allowed)}
for name in b:
 if name not in allowed: assert b[name]==a[name],'emitted bytes differ: '+name
first=next(name for name in b if name.endswith('.js'))
mutated=a.copy();mutated[first]+=b'\nconsole.log("mutant");\n'
try:
 assert all(b[n]==mutated[n] for n in b if n not in allowed),'emitted bytes differ'
except AssertionError:caught=True
else:raise AssertionError('byte mutant survived')
api='typescript.d.ts'
assert api in b
api_mutated=a.copy();api_mutated[api]+=b'\ninterface UnreviewedMutant { value: number; }\n'
try:
 assert all(b[n]==api_mutated[n] for n in b if n not in allowed),'published API bytes differ'
except AssertionError:api_caught=True
else:raise AssertionError('API byte mutant survived')
internal_mutated=a.copy();internal_mutated['tsserverlibrary.internal.d.ts']+=b'\ninterface UnreviewedInternalMutant {}\n'
try:
 assert all(b[n]==internal_mutated[n] for n in b if n not in allowed),'unexpected internal API bytes differ'
except AssertionError:internal_caught=True
else:raise AssertionError('unexpected internal API byte mutant survived')
report={'javascriptFiles':sum(n.endswith('.js') for n in b),'publicAndLibraryDeclarationFiles':sum(n.endswith('.d.ts') and n not in internalNames for n in b),'internalDeclarationFiles':len(internalNames),'bytes':sum(map(len,b.values())),'identicalExceptRecordedInternalDeclarations':True,'internalDeclarations':internal,'javascriptMutantCaught':caught,'publicApiMutantCaught':api_caught,'unreviewedInternalApiMutantCaught':internal_caught,'artifacts':{n:{'bytes':len(v),'sha256':hashlib.sha256(v).hexdigest()} for n,v in b.items()}}
(unit/'emitted-identity.json').write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps({k:v for k,v in report.items() if k!='artifacts'}))
