import html
import collections,gzip,json,pathlib
out=pathlib.Path(__file__).parent
root=pathlib.Path('/workspace/overload-adapted')
declarations=json.loads((out/'declarations.json').read_text())
def load(name):
 with gzip.open(out/(name+'.jsonl.gz'),'rt') as f:return [json.loads(x) for x in f]
def loc(where):return 'src/'+where.split('/src/',1)[1] if '/src/' in where else where
def position(where):
 file,line,col=loc(where).rsplit(':',2);lines=(root/file).read_bytes().decode().splitlines(keepends=True)
 return file,len(''.join(lines[:int(line)-1]).encode())+len(lines[int(line)-1].encode('utf-16-le')[:2*(int(col)-1)].decode('utf-16-le').encode())
def owner(where,assertion=False):
 file,pos=position(where)
 rows=[d for d in declarations.get(file,[]) if d['start']<=pos<d['end'] and (not assertion or d['name'].startswith('assert'))]
 return min(rows,key=lambda d:d['end']-d['start'])['name'] if rows else ''
def esc(s):return html.escape(s,quote=False).replace('|','\\|').replace('\n','<br>')
def reasons(fs):return '<br>'.join('"'+esc(f['kind']+': '+f['reason'])+'"' for f in fs) or 'None'
base,head=load('debug-parent'),load('debug-head')
findings=lambda rows:[f for r in rows for f in r.get('findings',[])]
bf,hf=findings(base),findings(head)
def sites(fs):
 result=collections.defaultdict(dict)
 for f in fs:
  if '/src/compiler/debug.ts:' in f['where']:result[f['where']][(f['kind'],f['reason'],f['text'])]=f
 return {k:list(v.values()) for k,v in result.items()}
bs,hs=sites(bf),sites(hf)
lines=['## debug.ts, full mode','','Both streams load the same checker-clean tsc entry graph (81 source files), then select only debug.ts for full-mode observation. Both end with a relation trailer and child exit 0. There are 81 debug units on each side. At locations in debug.ts itself, raw events are 20 -> 12 and unique findings are 17 -> 11; the remaining two unique findings are in core.ts. The full event streams, including two unchanged dependency findings, contain 22 -> 14 findings; unique raw findings and distinct reads are both 19 -> 13, including Boundary markers. Refused findings are 3 -> 1, NotYet findings remain 5, and Boundary markers are 11 -> 7. Two callback refusal locations and four continuation Boundary markers clear. There are no moved reasons, new findings, new sites, panics or regressions in this full-mode pair. Four attempted units become finding-free.','','The predicted direction holds: assertNode and assertOptionalNode clear their callback refusal, while assertEachNode keeps it (its body passes test on to every). The observed callback locations are 273:82 and 299:78, each emitted in two attempting-unit contexts. This full census therefore observes two distinct callback locations clearing, rather than three. The source has three narrowing callback overload signatures across these two functions (one assertNode and two assertOptionalNode), but the overload-group census returns at the first failed overload, so the baseline does not emit a separate overload-2 diagnostic for assertOptionalNode. The source evidence is retained in parent-census-small.go.txt and head-census-small.go.txt. Generic signature declarations remain deferred and are not counted as independently lowered bodies.','','Every assert-family finding site follows. Boundary rows are the meter\'s continuation markers, retained to make the raw pairing fully accountable. Locations are relative to src/compiler/debug.ts; reason strings are quoted in full.','','| Function | Site | Parent | Head |','| --- | --- | --- | --- |']
table=[]
for where in sorted(set(bs)|set(hs),key=lambda w:position(w)[1]):
 name=owner(where,True)
 if not name:continue
 row=dict(function=name,site=loc(where),base=bs.get(where,[]),head=hs.get(where,[]))
 table.append(row);lines.append('| '+name+' | '+loc(where).split('debug.ts:')[1]+' | '+reasons(row['base'])+' | '+reasons(row['head'])+' |')
lines+=['','Every assert-family declaration, including finding-free siblings and deferred signatures:','','| Function | Declaration | Eligibility, parent -> head | First non-Boundary stop, parent | First non-Boundary stop, head |','| --- | --- | --- | --- | --- |']
bu={u['where']:u for r in base for u in r.get('units',[]) if u.get('name','').startswith('assert')};hu={u['where']:u for r in head for u in r.get('units',[]) if u.get('name','').startswith('assert')}
for where,u in sorted(bu.items(),key=lambda kv:position(kv[0])[1]):
 def stops(fs):
  seen={}
  for f in fs:
   if f.get('unit')==where and f['kind']!='Boundary':seen[(f['kind'],f['reason'])]=f
  return list(seen.values())
 lines.append('| '+u['name']+' | '+loc(where).split('debug.ts:')[1]+' | '+u['status']+' -> '+hu[where]['status']+' | '+reasons(stops(bf))+' | '+reasons(stops(hf))+' |')
lines+=['','Other debug.ts sites that move: **none**. The unchanged non-assert debug findings are listed here for completeness.','','| Function | Site | Parent | Head |','| --- | --- | --- | --- |']
for where in sorted(set(bs)|set(hs),key=lambda w:position(w)[1]):
 if owner(where,True):continue
 lines.append('| '+owner(where)+' | '+loc(where).split('debug.ts:')[1]+' | '+reasons(bs.get(where,[]))+' | '+reasons(hs.get(where,[]))+' |')
external=[f for f in bf if '/src/compiler/debug.ts:' not in f['where']]
lines+=['','Full-mode dependency findings: the comparator finding at core.ts:1049:82 and its Boundary at core.ts:1049:5 are unchanged. A debug-only observation can report a dependency stop; these are not additional changed debug.ts sites.','','The exact access-chain pair output is in [debug-paired.json](debug-paired.json) and [debug-paired.md](debug-paired.md).','']
(out/'debug-table.json').write_text(json.dumps(table,indent=2)+'\n')
(out/'debug-section.md').write_text('\n'.join(lines))
print('assert-family finding sites',len(table),'assert-family declarations',len(bu))
