import json,re,sys
from pathlib import Path

def read(name, shift=False):
    records=[json.loads(l) for l in Path(name).read_text().splitlines()]
    assert len(records)==2
    def normalized(record):
        root=str(Path(record['roots'][0]).parents[2])+'/'
        result=[]
        for diagnostic in record['diagnostics']:
            diagnostic=diagnostic.replace(root,'')
            if shift:
                diagnostic=re.sub(r'(src/compiler/checker.ts:)(\d+)(:)',lambda m:m[1]+str(int(m[2])-10*int(int(m[2])>50153))+m[3],diagnostic)
            result.append(diagnostic)
        return sorted(result)
    assert normalized(records[0])==normalized(records[1])
    return normalized(records[0])

before=read(sys.argv[1]);after=read(sys.argv[2],True);mutant=read(sys.argv[3])
removed=sorted(set(before)-set(after));assert not set(after)-set(before)
assert len(removed)==1 and removed[0].startswith('src/compiler/checker.ts:50153:121: error TS2532:')
assert before==mutant
own=lambda rows:sum(bool(re.match(r'src/compiler/checker.ts:\d+:\d+: error TS',d)) for d in rows)
assert (len(before),len(after),len(mutant))==(131,130,131)
assert (own(before),own(after),own(mutant))==(17,16,17)
print(json.dumps(dict(before=dict(wholeProgram=len(before),checkerOwnFile=own(before)),after=dict(wholeProgram=len(after),checkerOwnFile=own(after)),mutant=dict(wholeProgram=len(mutant),checkerOwnFile=own(mutant)),removed=removed,added=[],coordinateMapping='The scoped branch replacement adds ten CRLFs and shifts later checker.ts primary locations by ten; map them back before comparing full diagnostic sets.'),indent=2))
