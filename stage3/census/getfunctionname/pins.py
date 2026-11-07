from pathlib import Path
import json,hashlib,re
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/getfunctionname');root=repo/'stage3/adapt'
def line(file,needle):
 lines=Path(file).read_text().splitlines();hits=[i+1 for i,l in enumerate(lines) if needle in l];assert hits,(file,needle);return hits[0]
entries=[]
def add(adaptation,file,needle,status,scope,detail):
 p=root/file if not file.startswith('/') else Path(file);entries.append({'adaptation':adaptation,'file':str(p.relative_to(repo)) if p.is_relative_to(repo) else str(p),'line':line(p,needle),'status':status,'scope':scope,'detail':detail,'sha256':hashlib.sha256(p.read_bytes()).hexdigest()})
add('30-indexed-reads','30-indexed-reads/sites.json','"function": "getFunctionName"','active','match[1] element read','debug.ts match[1], occurrence 1 of total 1, action assert, U-regex. adapt.cjs groups all matching expressions in the file and requires this site; function/line fields document the survey rather than locate edits.')
add('41-explicit-any-remaining','41-explicit-any-remaining/rules.json','return (func as any).name;','active','exact return expression in getFunctionName','Rule 6 requires exactly one old/new spelling: return (func as any).name; becomes return func.name;. The adapter reads rules.json and rejects a missing/duplicate reviewed site.')
add('45-regex-captures','45-regex-captures/adapt.cjs','const debug = func("debug", "getFunctionName");','active','named function, match initializer and capture 1','Requires a unique getFunctionName implementation in debug.ts, a literal regexp exec initializer for match, and its match[1] read. The regexp proof must establish capture 1 is mandatory.')
add('99-correctness-core-debug-messages (accepted road b overlay)',str(s/'road-b/pins.json'),'"owner": "getFunctionName"','active in requested main-plus-road-b tree','whole unchanged getFunctionName owner','The accepted adapter guards array pins the complete getFunctionName implementation byte for byte. inspect() requires the owner even on an already-adapted tree. Adapter and pins blobs were checked against accepted cut 2a95448b8fadbcd30549a148bd28da5f8b75f907 via GitHub.')
add('99-correctness-axolotl-n','99-correctness-axolotl-n/proposed-pins.json','"owner": "Debug.getFunctionName"','inactive proposal','proposed whole owner and func parameter','This proposal was dropped with the four H06 assertion-call sites. Actual adapt.cjs imports pins.json, whose only owner is Debug.formatControlFlowGraph.hasNode, not proposed-pins.json.')
add('99-correctness-core-multimap','99-correctness-core-multimap/owners.json','80e948812b0d1fba4daafa211a8576f14a0cc67dc87bff66720ca64888a6f030','active, indirect import pins','whole import statements containing cast','Four exact import pins contain the imported caller name cast: src/compiler/checker.ts, src/services/codeFixProvider.ts, src/services/codefixes/importFixes.ts and src/services/codefixes/inferFromUsage.ts. These pin imports, not the cast implementation or its getFunctionName expression.')
# Broad searches are retained to distinguish active pin inputs from historical evidence.
textHits=[];ownerHits=[]
for p in sorted(root.rglob('*')):
 if not p.is_file():continue
 if p.suffix=='.cjs' and (p.name=='adapt.cjs' or p.parent==root/'41-explicit-any-remaining'):
  text=p.read_text()
  if re.search(r'\b(getFunctionName|deprecate)\b',text) or re.search(r"['\"]cast['\"]|function cast|\bcast<",text):textHits.append(str(p.relative_to(repo)))
 if p.suffix=='.json' and any(t in p.name.lower() for t in ['owner','pin','rule','site']):
  try:data=json.loads(p.read_text())
  except json.JSONDecodeError:continue
  def walk(x):
   if isinstance(x,dict):
    for k in ['owner','function','name']:
     if str(x.get(k,'')) in ['cast','deprecate','getFunctionName','Debug.getFunctionName']:ownerHits.append({'file':str(p.relative_to(repo)),'field':k,'value':x[k],'fields':list(x)})
    for v in x.values():walk(v)
   elif isinstance(x,list):
    for v in x:walk(v)
  walk(data)
result={'activeAndProposedPins':entries,'callerPins':{'src/compiler/core.ts:cast':'No targeted active owner or exact-message pin found in main or the accepted road-b overlay. 41 REMAINING.md lists its any default as unresolved survey evidence, not an applied rule.','src/deprecatedCompat/deprecate.ts:deprecate':'No targeted active owner or exact-call pin found in main or the accepted road-b overlay.'},'sourceSearchHits':textHits,'jsonOwnerSearchHits':ownerHits,'notActivePins':['40-explicit-any debug.ts evidence and later-debug-adapter lists are historical evidence, not whole-file source guards for getFunctionName or these callers.','99-cli-debug-assertion-level witnesses and README mention the former lazy messages; its active adapter pins helper callers and removed-name uses, not getFunctionName or cast/deprecate.','99-correctness-axolotl-r/source-manifest.json and other proof/scope manifests describe historical source snapshots; they are not imported as target-owner guards by their adapters.'],'getAllKeysMoveOnMain':'stage3/adapt/99-correctness-core-getallkeys/adapt.cjs and owners.json'}
(s/'raw/pins.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({'pins':entries,'callerPins':result['callerPins']},indent=2))
