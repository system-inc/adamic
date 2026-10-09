import pathlib, json, subprocess, time, difflib
repo=pathlib.Path('/workspace/adamic')
evidence=repo/'stage1/cohere/lint/validation/compiler-agree-corpus-sources'
source=repo/'stage1/cohere/lint/corpus_sources_test.go'
original=source.read_text()
mutants={
 'fixture-leak': ('if fixture {','if false && fixture {','exclusion exceeded named directory contract'),
 'expanded-policy': ('{"testdata", "validation", "evidence"}','{"testdata", "validation", "evidence", "tests"}','exclusions grew beyond the named folders'),
 'prefix-exclusion': ('slices.Contains(compilerCorpusFixtureFolders, directory)','strings.HasPrefix(directory, "testdata") || slices.Contains(compilerCorpusFixtureFolders, directory)','exclusion exceeded named directory contract'),
}
summary=[]
for name,(before,after,message) in mutants.items():
 assert before in original, name
 changed=original.replace(before,after,1)
 scratch=pathlib.Path('/workspace/scratch')/('corpus-mutant-'+name+'.go')
 scratch.write_text(changed)
 overlay=evidence/(name+'-overlay.json')
 overlay.write_text(json.dumps({'Replace':{str(source):str(scratch)}},indent=2)+'\n')
 (evidence/(name+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='original',tofile=name)))
 cmd=['go','test','-overlay',str(overlay),'./stage1/cohere/lint','-run','^TestCompilerCorpusSourcesExcludeOnlyNamedFolders$','-count=1','-v','-failfast','-timeout=0']
 started=time.monotonic()
 with (evidence/(name+'.txt')).open('w') as log:
  result=subprocess.run(cmd,cwd=repo,stdout=log,stderr=subprocess.STDOUT)
 output=(evidence/(name+'.txt')).read_text()
 assert result.returncode==1 and message in output and 'build failed' not in output,(name,result.returncode,output)
 summary.append({'mutant':name,'exit':result.returncode,'expected_failure':message,'wall_seconds':time.monotonic()-started})
(evidence/'mutants-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary))
