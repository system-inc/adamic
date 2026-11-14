from pathlib import Path
import json
s=Path('/tmp/resolve-host-1612');d=s/'work/stage3/census/resolve-host-1612';audit=json.loads((s/'evidence/resolver-audit.json').read_text())
# Presence tests and optional calls are kept separate from ordinary callee reads.
rows=[
 {'member':'fileExists','modulePresenceSites':[],'additionalClosurePresenceSites':[],'feasibleRequirePresenceSites':0},
 {'member':'readFile','modulePresenceSites':[],'additionalClosurePresenceSites':[],'feasibleRequirePresenceSites':0},
 {'member':'directoryExists','modulePresenceSites':[820],'additionalClosurePresenceSites':['utilities.ts:7819'],'feasibleRequirePresenceSites':1},
 {'member':'realpath','modulePresenceSites':[1975],'additionalClosurePresenceSites':[],'feasibleRequirePresenceSites':1},
 {'member':'getDirectories','modulePresenceSites':[820],'additionalClosurePresenceSites':[],'feasibleRequirePresenceSites':0},
 {'member':'trace','modulePresenceSites':[124],'additionalClosurePresenceSites':[],'feasibleRequirePresenceSites':0},
 {'member':'getCurrentDirectory','modulePresenceSites':[488,2331,2582,2663,2870,2898,2924,2974],'additionalClosurePresenceSites':['utilities.ts:6609'],'feasibleRequirePresenceSites':0},
 {'member':'getGlobalTypingsCacheLocation','modulePresenceSites':[3056],'additionalClosurePresenceSites':[],'feasibleRequirePresenceSites':1},
 {'member':'useCaseSensitiveFileNames','modulePresenceSites':[],'additionalClosurePresenceSites':[],'feasibleRequirePresenceSites':0,'typeAndTruthinessTests':['moduleNameResolver.ts:512','moduleNameResolver.ts:3425','moduleNameResolver.ts:3426','utilities.ts:6482 (callback adapter, not this host)'],'feasibleValueOrTypeReads':['moduleNameResolver.ts:512 (type test plus selected value read)']},
]
for x in rows:
 x['modulePresenceReads']=len(x['modulePresenceSites']);x['additionalClosurePresenceReads']=len(x['additionalClosurePresenceSites'])
(d/'presence.json').write_text(json.dumps({'definition':'Syntactic presence tests (truthiness/undefined checks on optional callbacks) and optional-call tests; ordinary callee reads are separate. Boolean type/value observations are explicit in their row. Whole-module sites outside the call closure are included and marked unreachable below.','rows':rows,'functionIdentityReads':0,'forwarderLengthReads':0,'writesToNodeSystemMembersAfterConstruction':0},indent=2)+'\n')
# All possible forwarded-member accesses in the conservative closure have been
# inspected: each is a direct call/optional call or a presence check. Boolean
# dispatch is not function identity observation. The cache's existing.directoryExists
# is a cached boolean, not a forwarded member.
forwarded={'fileExists','readFile','directoryExists','realpath','getCurrentDirectory','getDirectories','trace','getGlobalTypingsCacheLocation'}
reads=[a for a in audit['accesses'] if a['member']in forwarded and a['receiver']in ['host','state.host','hostOrText']]
for a in reads:
 assert not a['write']
 assert a['invoked'] or (a['file'],a['line'],a['member'])in [('src/compiler/moduleNameResolver.ts',124,'trace'),('src/compiler/moduleNameResolver.ts',1975,'realpath'),('src/compiler/utilities.ts',7819,'directoryExists')],a
functions=audit['functions'];assert not any(':getAutomaticTypeDirectiveNames' in x or ':getEffectiveTypeRoots' in x or ':loadEntrypointsFromTargetExports' in x for x in functions)
(d/'observers.json').write_text(json.dumps({'conservativeClosureFunctions':len(functions),'conservativeClosureEdges':len(audit['edges']),'forwardedMemberAccesses':reads,'identityReads':0,'lengthReads':0,'writes':0,'classification':'All forwarded callback reads in the closure are direct/optional calls or the three explicit presence tests. Dynamic field access and host reflection checked separately in CHAIN.md.','excludedCacheBoolean':'moduleNameResolver.ts:2463 existing.directoryExists'},indent=2)+'\n')
