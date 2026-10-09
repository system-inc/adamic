import pathlib,json,collections
p=pathlib.Path('/workspace/adamic/review/test-defend/internal-native-radix');names=['TestRecordBenchmark','TestRuntimeStringEquality','TestRuntimeReleasePaths','TestRegExpIteratorResultShape'];go={};c={}
for n in names:
 go[n]={l.split()[0] for l in (p/(n+'.cover')).read_text().splitlines()[1:] if int(l.split()[-1])>0};covered=set()
 data=json.loads((p/'c-coverage'/n/'export.json').read_text())
 for d in data['data']:
  for f in d['files']:
   name=pathlib.Path(f['filename']).name
   if not (pathlib.Path('/workspace/adamic/internal/native/runtime')/name).exists():continue
   # LLVM segments delimit executable source regions, including zero-count nested blocks.
   for i,segment in enumerate(f['segments']):
    if not segment[3] or segment[2]<=0 or (len(segment)>5 and segment[5]):continue
    following=f['segments'][i+1] if i+1<len(f['segments']) else segment
    end=following[0]+(following[1]>1)
    for line in range(segment[0],max(segment[0]+1,end)):covered.add('internal/native/runtime/'+name+':'+str(line))
 c[n]=covered
result=[]
for a,b in [(names[0],names[1]),(names[2],names[3])]:
 result.append(dict(test=a,subsumer=b,go_exclusive_blocks=sorted(go[a]-go[b]),c_exclusive_lines=sorted(c[a]-c[b]),c_covered_lines=sorted(c[a]),subsumer_c_covered_lines=sorted(c[b])))
(p/'coverage-diffs.json').write_text(json.dumps(result,indent=2))
for r in result:
 print(r['test'],'go',len(r['go_exclusive_blocks']),'C',len(r['c_exclusive_lines']));print([s for s in r['c_exclusive_lines'] if any(f in s for f in ['heap.c','map.c','record.c','count.h','class_inheritance.c'])][:90])
