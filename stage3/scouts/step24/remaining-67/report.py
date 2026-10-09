#!/usr/bin/env python3
import json,re
from pathlib import Path
from collections import Counter
u=Path(__file__).resolve().parent
rows=json.loads((u/'owners.json').read_text()); replay=json.loads((u/'replay20.json').read_text())
nullable=json.loads((u/'nullable-experiment.json').read_text())
nullable_sites={(r['file'].removeprefix('src/compiler/'),r['line'],r['code']) for r in nullable['resolved_findings'] if r.get('code')}
method_families={'checker.ts:54261','program.ts:1876','program.ts:2599','program.ts:472','program.ts:5107','watch.ts:756','watch.ts:845'}
public_sites={'moduleNameResolver.ts:132','moduleNameResolver.ts:627','moduleNameResolver.ts:635','commandLineParser.ts:2663'}
for r in rows:
 key=r['file']+':'+str(r['line']); owners=r['ownerCandidates']
 r['adaptations']={
 '20':'',
 '30-33':'No indexed expression at this diagnostic site. Required indexed-read assertions do not widen an optional declaration.',
 '70':'This diagnostic is about undefined/presence, not writable-view variance. Readonly alone preserves the incompatible value contract.',
 '75':'No diagnostic chain here targets the reviewed Type caches, JSON configuration metadata or errorModuleName slots. Existing optional fields are not missing metadata.',
 }
 if r['file']=='builder.ts':
  r.update(disposition='covered',owner='adaptation 20',witness='optional-object.a')
  r['adaptations']['20']='Covered. A prior structural mismatch masks the optional-code diagnostic at numeric step 20; the bounded assignment pattern resolves each reviewed field separately.'
 elif (r['file'],r['line'],r['code']) in nullable_sites:
  r.update(disposition='covered',owner='adaptation 20',witness='nullable-context.a')
  r['adaptations']['20']='Covered internal owner. Contextual target includes undefined; getPropertiesOfType returned no common properties. The new pattern examines its non-null object constituent with unchanged value checks.'
 elif r['code']==2420:
  r.update(disposition='incidental candidate',owner='compiler',witness='class-optional-presence.a')
  r['adaptations']['20']='TS2420 is outside the seed codes. This is the same SymbolTracker.moduleResolverHost slot selected by independent TS2379 evidence; the after measurement determines whether it clears incidentally.'
 elif r['line']==17890 and r['file']=='checker.ts':
  r.update(disposition='declined',owner='adaptation declaration review',witness='indexed-optional-write.a')
  r['adaptations']['20']='Deliberately excluded: namedMemberDeclarations?.[i] is an optional indexed read, not a present-undefined declaration seed.'
  r['adaptations']['30-33']='Partition 31 covers checker.ts. Tuple labels are intentionally optional, including namedMemberDeclarations?.[i]; asserting this read invents a required label.'
 elif r['file']=='resolutionCache.ts' and r['line']==1386:
  r.update(disposition='declined',owner='compiler generic relation',witness='generic-reset.a')
  r['adaptations']['20']='Deliberately excluded generic receiver T. An arbitrary specialization may narrow files; widening the base declaration does not prove the write through T.'
 elif any(o['kind']=='MethodSignature' for o in owners) or key in method_families:
  r.update(disposition='declined',owner='compiler optional method presence',witness='optional-method.a' if r['code']==2412 else 'optional-method-object.a')
  r['adaptations']['20']='Deliberately excluded optional method. Adding undefined to its result does not widen the callable; converting it to a property changes method variance and API shape.'
 elif r['file']=='watchPublic.ts' and r['line']==568:
  r.update(disposition='declined',owner='adaptation tagged-state review',witness='optional-write.a')
  r['adaptations']['20']='Receiver is FilePresentOnHost | FilePresenceUnknownOnHost. One fileWatcher declaration is required, so the all-owners-optional rule declines the union. Closing the watcher does not itself change that tagged view.'
 elif r['code'] in (2322,2345):
  r.update(disposition='outside seed codes',owner='compiler stricter-options relations',witness='nullable-call.a' if r['code']==2345 else 'nested-optional.a' if 'exactOptionalPropertyTypes' in r['message'] else 'nullable-assignment.a')
  r['adaptations']['20']='TS'+str(r['code'])+' is outside adaptation 20\'s three seed codes. '+('Nested present-undefined incompatibility requires a separately reviewed relation/owner pattern.' if 'exactOptionalPropertyTypes' in r['message'] else 'Required argument/result accepts no undefined; no optional declaration is the immediate target.')
 elif key in public_sites:
  r.update(disposition='API constraint',owner='adaptation public declaration review',witness='optional-object.a')
  r['adaptations']['20']='The nullable contextual target hides a compatible optional property. Its declaration is public; adding undefined changes emitted API bytes, which this unit requires identical. The internal-only extension declines it.'
 elif r['file']=='resolutionCache.ts' and r['line'] in (990,1019):
  r.update(disposition='declined',owner='compiler generic relation',witness='generic-nullable-contract.a')
  r['adaptations']['20']='reusedNames belongs to ResolveNamesWithLocalCacheInput<Entry,...>. The source specialization and uninstantiated declaration annotation are not interchangeable; the existing per-constituent relation check declines it. No blanket generic widening.'
 elif r['file']=='watchPublic.ts' and r['line'] in (371,383):
  r.update(disposition='declined',owner='compiler generic relation',witness='generic-nullable-contract.a')
  r['adaptations']['20']='createProgram belongs to internal CreateWatchCompilerHostInput<T>. Source and declaration T are distinct binders: relating the actual callable to the uninstantiated owner annotation fails the existing compatibility check. The nullable traversal alone does not establish arbitrary specialization safety.'
 else:
  r.update(disposition='nested contract review',owner='compiler structural presence',witness='nested-optional.a')
  r['adaptations']['20']='The existing top-level contextual-property selector does not traverse this nested instantiated contract. It does not provide a compatible optional-owner seed as written; recursive structural selection needs its own reviewed rule.'
comparison=json.loads((u/'parser-comparison.json').read_text())
cleared=set(comparison['cleared'])
for r in rows:
 r['afterStatus']='cleared' if r['id'] in cleared else 'remaining'
 if r['afterStatus']=='cleared' and r['disposition'] in ('outside seed codes','incidental candidate'):
  r['disposition']='incidental clearance'
  r['owner']='adaptation 20 shared owner'
  r['adaptations']['20']+=' The actual after build clears it through the independently selected shared slot; this code is not a new seed.'
(rfile:=u/'dispositions.json').write_text(json.dumps(rows,indent=2)+'\n')
text='# Every pinned stop\n\nCoordinates refer to the adapted tree measured by f053ef44, not to pristine source. Each witness is a structural minimal program; the full Node compiler suite exercises the real sites.\n\n'
text+='| Stop | After | Disposition | Owner | Why 20 missed it | Node witness |\n|---|---|---|---|---|---|\n'
for r in rows:text+='| '+r['id']+' | '+r['afterStatus']+' | '+r['disposition']+' | '+r['owner']+' | '+r['adaptations']['20'].replace('|','\\|')+' | [fixture](fixtures/'+r['witness']+') |\n'
(u/'STOPS.md').write_text(text)
(u/'summary.json').write_text(json.dumps({'pinned':len(rows),'after':len(comparison['after']),'cleared':len(cleared),'firstAfter':comparison['firstAfter'],'codes':dict(Counter(r['code'] for r in rows)),'dispositions':dict(Counter(r['disposition'] for r in rows))},indent=2)+'\n')
print((u/'summary.json').read_text())
