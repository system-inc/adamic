import pathlib, subprocess, os, json, time, re
root=pathlib.Path('/workspace/adamic'); out=root/'review/test-audit/internal-native-element_borrow'
rows='TestNbodyIndexedElementsBorrow TestThrowElementBorrowPlan TestCallTargetsElementBorrowPlan TestDevirtualizeBorrowDocClaim TestUniformFieldsMatchNode TestRuntimeFieldLayoutsAreIncluded TestRegexProgramsKeepCheckedFieldReads TestOptionalWriteMissingSlotRemainsChecked TestFreedValuesAreCaughtWithSlabs TestSizeClassesShareTheirChunks TestResidentSetUnits TestIeee754MatchesNodeBitForBit TestCEndsInNewline TestLibraryMapSetIteratorResources'.split()
(out/'rows.json').write_text(json.dumps(rows,indent=2)); pattern='^('+'|'.join(rows)+')$'; (out/'scope.regex').write_text(pattern)
def run(name,args,env=None):
 st=time.monotonic()
 with (out/(name+'.log')).open('w') as f: p=subprocess.run(args,cwd=root,env=env or os.environ,stdout=f,stderr=subprocess.STDOUT)
 with (out/'commands.jsonl').open('a') as f:f.write(json.dumps({'name':name,'command':args,'seconds':time.monotonic()-st,'exit':p.returncode})+'\n')
 return p.returncode
if __name__=='__main__':
 discovered=(out/'list.log').read_text().splitlines(); assert all(r in discovered for r in rows)
 run('scope-baseline',['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverprofile='+str(out/'scope.cover'),'./internal/native/','-run',pattern])
 run('coverage-functions',['go','tool','cover','-func='+str(out/'scope.cover')])
 for r in rows:
  for n in range(1,4):run('timing-'+r+'-'+str(n),['timeout','120','go','test','-count=1','-timeout','90s','./internal/native/','-run','^'+r+'$'])
