"""Render complete paired evidence; partial rendering is explicitly marked."""
import collections,json,pathlib,re,sys
scratch=pathlib.Path(sys.argv[1]).resolve();data=json.loads((scratch/'analysis/paired.json').read_text());callers=json.loads((scratch/'callers.json').read_text());out=scratch/'analysis'
clean=lambda s:str(s).replace(str(scratch/'adapted')+'/','')
escape=lambda s:clean(s).replace('|','\\|').replace('\n','<br>')
def words(item):
 if not item:return '—'
 f=item.get('finding',item);return f.get('reason','')+'\n'+f.get('text','')
def overload(item):
 if not item:return None
 f=item.get('finding',item);m=re.search(r'overload (\d+) of (\S+)',f.get('reason',''));return (m.group(2),int(m.group(1))) if m else None
# Every changed event, including administrative recovery boundaries.
lines=['# Complete finding grades','','Fingerprints are `(kind, phase, reason, full message)` with complete source coordinates replaced by `<position>`. Same-source unit identity is correspondence, not part of the diagnostic tuple. Reached means an additional non-scan stop in a previously blocked unit; it is not a regression verdict or successful compilation. Raw positions and full original messages remain in paired.json and the streams.','']
for p in data['pairs']:
 lines += ['## '+p['mode']+' — '+p['file'],'']
 if 'scan' in p:
  lines += ['Scan-all base boundary: '+p['scan']['baseBoundary'], '','Head boundary: '+p['scan']['headBoundary'],'']
  for grade in ['gone','new']:
   for r in p['scan'][grade]:lines += [f'- {grade}, {r["count"]} event(s): '+escape(r['tuple'])]
  continue
 lines += [f'Before {p["baseFindings"]}; after {p["headFindings"]}; unchanged {p["unchanged"]}; grades {p["counts"]}.','']
 for i,c in enumerate(p['changes'],1):
  lines += [f'### {i}. {c["grade"]} — `{clean(c["unitKey"][-1])}`','']
  for side in ['base','head']:
   if side in c:
    f=c[side]['finding'];lines += [side+': '+f['kind']+' / '+f['phase']+' / `'+clean(f.get('where',''))+'`','','```text',f.get('reason',''),f.get('text',''),'```','']
 if not p['changes']:lines+=['No finding tuple changes.','']
(out/'FINDINGS.md').write_text('\n'.join(lines)+'\n')
# The 17 compiler/services mapped calls, separately from all-src extras.
lines=['# The 17 mapped compiler/services calls','','Counts are static call sites, never the number of repeated observations. A missing storage refusal is not proof of call success: a generic body may be deferred or another stop may precede it. Entry closure contains 11 of these calls; the other six are service calls measured under their own full-file entry roots.','',f'Observed storage-refusal sites: **{data["storageStopCalls17"]} of 17**. Required coverage complete: {data["coverage"]["allRequiredComplete"]}.','']
for row in data['mapped17']:
 c=row['call'];lines += [f'## {c["set"]} — {c["file"]}:{c["line"]}:{c["column"]}','','```ts',c['call'],'```','']
 if row['words']:
  for message in row['words']:lines += ['```text',message,'```','']
  lines += ['Observed in: '+', '.join(sorted({f["mode"]+' '+f['selectedFile'] for f in row['observations']})), '']
 else:lines += ['No certified-storage refusal observed in the completed required streams. This does not establish that the call passed.','']
(out/'MAPPED-CALLS.md').write_text('\n'.join(lines)+'\n')
# Name every moving non-target overload, with the unit's full before/after stops.
other=[]
for p in data['pairs']:
 if 'scan'in p:continue
 context={tuple(r['unitKey']):r for r in p.get('changedUnitFindings',[])}
 for c in p['changes']:
  names={o[0] for o in [overload(c.get('base')),overload(c.get('head'))] if o and o[0] not in ['arrayFrom','arrayToMap']}
  for name in names:other.append({'set':name,'mode':p['mode'],'selectedFile':p['file'],'grade':c['grade'],'change':c,'unitContext':context.get(tuple(c['unitKey']))})
(out/'other-overload-moves.json').write_text(json.dumps(other,indent=2)+'\n')
lines=['# Other overload-result/parameter findings','','This ledger includes every non-target overload movement found by the required entry/full pairs; supplemental compiler-directory core data is identified separately. `overload-ledger.json` retains all ordinal overload findings, including unchanged ones. Administrative Boundary messages are not counted as separate overload claims.','']
if not other:lines+=['No moving non-target ordinal overload reason in the completed pairs. Coverage complete: '+str(data['coverage']['allRequiredComplete'])+'.','']
for i,r in enumerate(other,1):
 lines += [f'## {i}. {r["set"]} — {r["grade"]}', '', r['mode']+' / '+r['selectedFile']+' / `'+clean(r['change']['unitKey'][-1])+'`','']
 for side in ['base','head']:
  values=[v['finding'] for v in (r.get('unitContext') or {}).get(side,[]) if v['finding'].get('kind')!='Boundary']
  lines += [side+':','']
  for f in values:lines += ['```text',clean(f.get('where','')),f.get('reason',''),f.get('text',''),'```','']
  if not values:lines+=['No diagnostic in that unit.','']
(out/'OTHER-OVERLOADS.md').write_text('\n'.join(lines)+'\n')
# Compact per-scope grade table; full details stay above.
lines=['# Paired run totals','','| Scope | File | Before | After | New | Gone | Changed | Reached |','|---|---|---:|---:|---:|---:|---:|---:|']
for p in data['pairs']:
 if 'scan' in p:continue
 c=p['counts'];lines.append('| '+p['mode']+' | '+p['file']+' | '+str(p['baseFindings'])+' | '+str(p['headFindings'])+' | '+' | '.join(str(c.get(k,0)) for k in ['new','gone','changed','reached'])+' |')
(out/'TOTALS.md').write_text('\n'.join(lines)+'\n')
print(json.dumps({'coverage':data['coverage'],'otherMovingSets':sorted(set(r['set'] for r in other)),'otherMoveEvents':len(other),'storageStopCalls17':data['storageStopCalls17']}))
