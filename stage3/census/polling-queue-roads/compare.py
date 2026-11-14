import json,re,difflib,pathlib,collections
p=pathlib.Path(__file__).parent/'evidence'
base=(p/'sys.original.ts').read_text().splitlines()
def findings(road):
 rows=[json.loads(x)for x in (p/f'census.{road}.jsonl').read_text().splitlines()]
 assert rows[0]['latent_mode']=='full' and rows[-1]['status']=='relation_report'
 assert rows[0]['module_format']=='CommonJSBundle',rows[0]['module_format']
 allf=[f for r in rows for f in r.get('findings',[]) if '/src/compiler/sys.ts:'in f.get('where','')]
 uniq={tuple(f.get(k,'')for k in ['kind','where','reason','text']):f for f in allf}
 return list(uniq.values())
original=findings('original')
assert any(f['where'].endswith('sys.ts:272:23') and f['reason']=="a cast the runtime can't check" for f in original)
summary={}
for road in 'abc':
 lines=(p/f'sys.{road}.ts').read_text().splitlines()
 mapping={j+1:i+1 for tag,i1,i2,j1,j2 in difflib.SequenceMatcher(None,base,lines,autojunk=False).get_opcodes()if tag=='equal' for i,j in zip(range(i1,i2),range(j1,j2))}
 def normalize(f,shift):
  def sub(m):
   line=int(m[1]);return 'sys.ts:'+str(mapping.get(line,'edited'+str(line)) if shift else line)+':'+m[2]
  return tuple(re.sub(r'[^\s]*?/src/compiler/sys.ts:(\d+):(\d+)',sub,f.get(k,'')) for k in ['kind','where','reason','text'])
 current=findings(road);assert not any(f['reason']=="a cast the runtime can't check" and 'PollingIntervalQueue' in (p/f'sys.{road}.ts').read_text().splitlines()[int(f['where'].rsplit(':',2)[1])-1] for f in current)
 old={normalize(f,False):f for f in original};new={normalize(f,True):f for f in current}
 gone=[v for k,v in old.items()if k not in new];added=[v for k,v in new.items()if k not in old]
 delta={'original_unique':len(old),'road_unique':len(new),'gone':gone,'new':added,'unchanged_after_line_mapping':len(old.keys()&new.keys()),'note':'Kind/where/reason/text comparison, matching unchanged source lines. Edited source sites appear as gone/new; the report groups changed sites explicitly.'}
 (p/f'delta.{road}.json').write_text(json.dumps(delta,indent=2)+'\n')
 js0=(p/'sys.original.js').read_text().splitlines(True);js1=(p/f'sys.{road}.js').read_text().splitlines(True)
 diff=''.join(difflib.unified_diff(js0,js1,fromfile='sys.original.js',tofile=f'sys.{road}.js'))
 (p/f'javascript.{road}.diff').write_text(diff)
 added_js=sum(l.startswith('+')and not l.startswith('+++')for l in diff.splitlines());gone_js=sum(l.startswith('-')and not l.startswith('---')for l in diff.splitlines())
 d0=json.loads((p/'typecheck.original.json').read_text())['diagnostics'];d1=json.loads((p/f'typecheck.{road}.json').read_text())['diagnostics']
 def dk(d):return json.dumps([d['file'],d['code'],d['message']])
 c0=collections.Counter(map(dk,d0));c1=collections.Counter(map(dk,d1))
 summary[road]={'diagnostics':len(d1),'new_diagnostics':list((c1-c0).elements()),'gone_findings':len(gone),'new_findings':len(added),'js_added_lines':added_js,'js_removed_lines':gone_js,'js_diff_bytes':len(diff.encode())}
(p/'summary.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary,indent=2))
