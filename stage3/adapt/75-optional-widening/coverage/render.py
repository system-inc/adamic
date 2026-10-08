#!/usr/bin/env python3
"""Render the measured partition and one property/source/target pattern table."""
from collections import Counter, defaultdict
import gzip
import json
from pathlib import Path

root = Path(__file__).resolve().parent
inventory = json.loads(gzip.decompress((root.parent / 'evidence/ruling/analysis.json.gz').read_bytes()))
counts = json.loads((root / 'counts.json').read_text())
measured = {r['id']: r for r in counts['rows']}
guards = {r['id']: r for r in json.loads((root / 'guards.json').read_text())}
annotations = {}

def add(ids, pattern, check):
    for key in ids.split():
        assert key not in annotations
        annotations[key] = {'pattern': pattern, 'check': check}

add('parser.ts:1854:40', 'Array metadata', 'parser.ts:1854 tests length, not JSDocArray subtype. utilitiesPublic.ts:1258 canHaveJSDoc and :1266 node.jsDoc ??= []; parser.ts:3171 canHaveJSDoc and cache presence precede writes. The array is enriched; no subtype tag.')
add('commandLineParser.ts:3767:87', 'Option-schema dispatch', 'commandLineParser.ts:3767 supplies the type-acquisition option map; :3783 looks up that map and :3784 tests opt before :3785 writes opt.name. No CompilerOptions subtype test; schema restricts keys, so the CompilerOptions union arm is not a runtime claim.')
add('checker.ts:5568:16', 'Factory or cached result', 'checker.ts:5568 passes TypeFlags.TypeParameter to createTypeWithSymbol; :5525 forwards flags to createType. The resulting TypeParameter is constructed before its later cache writes; :15149 dispatches existing TypeParameter views by flag.')
add('checker.ts:13602:16', 'Factory or cached result', 'checker.ts:13600 getDeclaredTypeOfTypeParameter returns either its symbol cache or createTypeParameter(:5567). The cached result contract is TypeParameter; :15149 is the flag dispatcher before constraint writes. Cache provenance needs symbol/alias proof.')
add('checker.ts:19938:30', 'Factory or cached result', 'checker.ts:19937 tests SymbolFlags.TypeParameter before getDeclaredTypeOfSymbol; its TypeParameter result comes from getDeclaredTypeOfTypeParameter(:13600), then createTypeParameter(:5567). Symbol-to-result provenance is outside the local census.')
add('checker.ts:20668:76', 'Factory or cached result', 'checker.ts:20668 immediately assigns createTypeParameter(tp.symbol) to restrictiveInstantiation before writing constraint on that same allocation. No separate subtype test; constructor uses TypeFlags.TypeParameter(:5568).')
add('checker.ts:16568:39', 'Factory or cached result', 'checker.ts:16664 getOrCreateTypeFromSignature creates ObjectFlags.Anonymous before returning its cached object; :16568 receives that object and :16569 writes mapper. No local tag test; signature-cache provenance establishes the shape.')
add('checker.ts:20967:24', 'Factory or cached result', 'checker.ts:20965 takes AnonymousType; :20967 preserves its object flags and adds Instantiated, then :20978 writes target. Caller dispatch at :15034 checks ObjectFlags.Anonymous; inherited flags/cache provenance is needed for this particular result.')
add('checker.ts:7122:128', 'Flag-selected operation', 'checker.ts:7122 tests getConstraintTypeFromMappedType(type).flags & TypeFlags.TypeParameter before passing the cached result to getConstraintOfTypeParameter. :14917-14919 stores and returns the same mapped-type constraint; call-result identity is not proved by the local census.')
add('checker.ts:23674:67', 'Flag-selected operation', 'checker.ts:23671 relation === comparableRelation && sourceFlags & TypeFlags.TypeParameter dominates the call. sourceFlags is the flags alias the ruling did not prove immutable. Constraint writes are in getConstraintFromTypeParameter(:16841).')
add('checker.ts:23680:71', 'Flag-selected operation', 'checker.ts:23676 someType(constraint, c => !!(c.flags & TypeFlags.TypeParameter)) dominates the loop call at :23680. For a union, this says some constituent, not the entire value; individual constraint/cache writes dispatch at :15149. Do not promote this to an all-members subtype proof.')
add('checker.ts:23898:60', 'Flag-selected operation', 'checker.ts:23895 sourceFlags & TypeFlags.TypeVariable admits TypeParameter or IndexedAccess; getConstraintOfType(:15149-15152) dispatches those separately before the :16845 constraint write. Composite TypeVariable alone does not prove TypeParameter.')
add('checker.ts:7164:93', 'Unproved assertion', 'checker.ts:7163 checks isHomomorphicMappedTypeWithNonHomomorphicInstantiation. Its :7107 test proves homomorphism of type.target, while :20866 checks the target helper variable for TypeParameter. No same-result TypeParameter tag check at the :7164 cast was proved; keep refused.')
add('checker.ts:13286:47 checker.ts:25200:31 checker.ts:25200:44 checker.ts:31797:31', 'Union alternative or shared cache', 'checker.ts:13280 checks ClassOrInterface or Reference and :13285 checks Tuple before getTupleBaseType; :25199 and :31796 establish class/base-result provenance. No TypeParameter subtype test is appropriate for these base types. getConstraintOfType(:15149) dispatches TypeParameter before its cache write. The census selects a structurally compatible target union alternative.')
add('checker.ts:14505:34 checker.ts:14567:34', 'Union alternative or shared cache', 'checker.ts:15024 and :15042-15046 dispatch Object, Union and Intersection separately. Union/intersection resolvers call setStructuredTypeMembers(:5648) to initialize their common member caches. No ObjectType subtype is established or required; the target includes a shared-cache union alternative.')
add('checker.ts:16961:69 checker.ts:16972:93', 'Error or wildcard sentinel', 'checker.ts:16960 failed pushTypeResolution and :16971 failed popTypeResolution select errorType placeholders, not TypeParameter instances. No TypeParameter subtype check at these assignments; later constraint writes use the :15149 flag dispatcher. Graph witnesses do not prove sentinels reach every writer.')
add('checker.ts:20893:98', 'Error or wildcard sentinel', 'checker.ts:20893 tests equality with wildcardType; :20896 excludes wildcardType and isErrorType(t) before object/tuple instantiation. No AnonymousType subtype is claimed for the wildcard. Later structured-cache writes use :15024/:15034 dispatch.')
add('checker.ts:20952:16 checker.ts:20952:43', 'Error or wildcard sentinel', 'checker.ts:20952 tests isErrorType(elementType), returning errorType instead of allocating an array. No ObjectType subtype is claimed for the sentinel. The structured resolver dispatches by flags at :15024-15046 before its common-cache writer(:5648).')
add('checker.ts:32168:37 checker.ts:32169:17 checker.ts:32170:17', 'Error or wildcard sentinel', 'checker.ts:32167 tests isImportCall and :32168-32170 selects stringType, import options or anyType by argument index. No ObjectType subtype test at these contextual-type results. Cache writers receive StructuredType(:5648); :15024-15046 dispatches actual structured kinds, while primitive apparent types need separate wrapper provenance.')
add('checker.ts:38334:16 checker.ts:38334:79', 'Error or wildcard sentinel', 'checker.ts:38334 selects the meta-object or errorType by node.name.escapedText === "meta". The error arm is a sentinel, not ObjectType. No local subtype check; actual structured-cache resolution dispatch is :15024-15046. Alias reachability of sentinels remains unproved.')
add('checker.ts:37997:31 checker.ts:38009:31', 'Cache view', 'checker.ts:37996 hasDefaultOnly && type && !isErrorType(type), or :38008 allowSyntheticDefaultImports && type && !isErrorType(type), precedes the cache cast and absence test at :37998/:38010. No runtime SyntheticDefaultModuleType class exists; this is a lazy metadata view of the same Type object.')
add('builder.ts:575:36', 'Diagnostic enrichment', 'builder.ts:605 returns a fresh object from convertToDiagnosticRelatedInformation, including a spread of the diagnostic; :575 binds it, and :576-581 immediately restores diagnostic-only metadata. No Diagnostic subtype tag; construction plus initialized writes, with spread/alias provenance requiring review.')
add('builder.ts:1505:48', 'Diagnostic enrichment', 'builder.ts:1522 returns a fresh object from toReusableDiagnosticRelatedInformation; :1505 binds it and :1506-1511 writes the reusable diagnostic metadata. No subtype test; allocation and immediate enrichment establish the intended shape, subject to spread/alias review.')
add('tsbuildPublic.ts:313:18', 'Host enrichment', 'tsbuildPublic.ts:313 receives createSolutionBuilderHostBase, then :314 initializes reportErrorSummary before :315 returns the host. No subtype tag; constructor/result ownership and pre-return enrichment are the required proof.')

rows = []
for r in inventory['rows']:
    key = f"{r['File']}:{r['Line']}:{r['Column']}"
    active = r['bucket'] == 'rest' or key in measured and measured[key]['bucket'] == 'rest'
    if not active:
        continue
    if r['bucket'] == 'b':
        annotation = {'pattern':'Reduced intersection component', 'check':'The reported source is a reduced impossible component, not the evaluated expression. Every trace sets IsNeverIntersection (checker reduction at cohere/TypeScript/tsc/internal/checker/checker.go:22196-22203). Existing source predicates and typed writer boundaries are recorded in guards.json; none proves a runtime value inhabits this component.'}
    else:
        annotation = annotations[key]
    rows.append({**r, 'id':key, **annotation})
assert len(rows) == 76 and len(annotations) == 34
assert len({r['id'] for r in rows}) == 76
patterns = dict(sorted(Counter(r['pattern'] for r in rows).items()))
groups = defaultdict(list)
for r in rows:
    for property in sorted({w['property'] for w in r['writes']}):
        groups[(property, r['Source'], r['Target'])].append(r)
report = {'partition':{'a':2,'b':5,'c':17,'rest':76},'patterns_by_relation_site':patterns,'grouped_rows':len(groups),'rows':[]}

def escape(text):
    return str(text).replace('|','&#124;').replace('\r','').replace('\n',' ')

text = ['Coverage: **42/47 expressions ran; 5 unobserved**. Rest now has **76 relation sites**.',
        'History: task branch starts at `c72c0c1`; required area merge is `c1938019`.',
        'Checks: **301/301 driver cases, 136,020 compiler/conformance checks, fresh self-project exit 0**.',
        'Cause: all **47 census traces report reduced impossible components**, with both checkers agreeing.',
        'Limits: observations count expression evaluation; they do not count alias-reaching writes or prove zero-hit code dead.', '',
        '# Measurement', '',
        'The scratch tree came from the complete `stage3/apply.sh` sequence at `c1938019`. Only counter imports, one shared map in core.ts and 47 comma-expression increments were added for measurement. Each original expression still evaluates once and returns its original value. The AST manifest binds by kind, text, named-function scope, parent-relative offset and four ancestor hashes. Saved coordinates are used only to establish the initial catalog; instrumentation uses the syntax fingerprints. The nested utilities.ts conditional and its true-arm property access are independently counted.', '',
        'Each Node process writes all 47 map entries synchronously on normal exit. Upstream forks workers rather than using worker threads. Aggregation requires 301 driver exit maps, five successful-suite maps (host and four workers), and one fresh self-check map, with exactly the same 47 nonnegative integer keys in every map. The tiny control is the driver\'s 301st case. A zero map from an earlier failed build attempt is excluded and retained separately in the failure evidence. A crash or missing exit map fails aggregation.', '',
        'Upstream `npm ci`, `npm run build` and `npm test -- --workers=4 --lint=false --no-colors --runners=compiler,conformance` passed. The oracle reports zero baseline differences. `tsc --project src/compiler/tsconfig.json --noEmit --pretty false --tsBuildInfoFile FRESH_PATH` exited 0; the fresh build-info path avoids upstream\'s composite cache. No checker option was weakened in these coverage runs. Raw maps, reports, logs and source hashes are in evidence/.', '',
        '| Census site | 300 cases + tiny | Compiler/conformance | Fresh self-check | Total | Disposition |',
        '|---|---:|---:|---:|---:|---|']
for r in counts['rows']:
    text.append(f"| {r['id']} | {r['driver']} | {r['upstream']} | {r['self']} | {r['total']} | {'rest' if r['total'] else 'b, unobserved'} |")
text += ['', '# Compiler handoff for @system_adamic_compiler', '',
         '**The presumed whole-expression checker disagreement is not present.** The old census stored `found.source` and `found.target` from a recursive relation walk. Those can be a union constituent, a nested name/parent/moduleReference field, an array element, or a contravariant callback parameter. Its output omitted that path. Reproducing the f1c9173 traversal gives a printed `never` for all 47, but every such source has raw `TypeFlags.Intersection`, not `TypeFlags.Never`, and the checker\'s `ObjectFlags.IsNeverIntersection` bit is set. The outer expressions usually remain inhabited. At checker.ts:52739 the whole expression really prints never under the pinned checker; stock reports an error/any-backed never display, and the site was unobserved.', '',
         'An example is utilities.ts:4416:52: the outer type prints LiteralExpression & StringLiteral. Its raw union includes four impossible intersections without an id property. The census selects the first impossible constituent, notices the missing optional id and reports that constituent\'s printed never. At esDecorators.ts:1412:99 the source expression is a function; the path runs through its contravariant node parameter, then name, then an impossible union constituent. All 47 paths and flag comparisons are saved in evidence/checker.json.gz.', '',
         'Both stock 6.0.3 and the pinned checker retain and print the impossible constituent as never. The shim type aliases name the real checker type; Type_flags and Type_objectFlags agree with its exported Flags and ObjectFlags methods on all 47 traces. Thus this evidence identifies a census relation-normalization and attribution error, not a checker incorrectly making the executing whole value unreachable. No TypeScript program with the requested whole-value disagreement was found; claiming one would contradict these observations.', '',
         'The smallest reproducer found for the actual census error is [reduced-union.a](reduced-union.a), three lines:', '', '```ts',
         (root / 'reduced-union.a').read_text().rstrip(), '```', '',
         'Both checkers print narrow as A and accept it without diagnostics. Their raw union also includes ({ kind: 1 } & A), which correctly reduces to never. The read-only traced census traversal reports that impossible component as missing id in { id?: number }. Removing the contradictory union arm eliminates the report. This does not reproduce an erroneous whole-value never type.', '',
         'Option experiments use read-only scratch queries, never adapted-build options. The miniature reproducer reports the same impossible component with exactOptionalPropertyTypes off, noUncheckedIndexedAccess off, both plus verbatimModuleSyntax/erasableSyntaxOnly off, and strict off. Across the real 47 sites, the first three comparisons retain all 47 reports; disabling strict retains 43. Strict can affect which recursive relation is reached, but it does not turn the executing outer value into the reported impossible member. The complete production option profile, including overriding upstream strictBindCallApply/useUnknownInCatchVariables exceptions, ES2024, ESNext, bundler resolution, force module detection and noEmit, also reproduces all 47. Both profiles are saved separately. Production and coverage options stay unchanged.', '',
         'Recommended compiler follow-up: normalize relation constituents using the checker\'s reduced type before testing absent optional members; impossible constituents cannot supply a hidden field. Keep the outer expression type, path and constituent flags separate in the census. Do not make native unreachable decisions from a constituent\'s printed type or from this bucket. No compiler code changed in this unit. The five zero-hit sites remain b solely as unobserved cases, not as proven dead code.', '',
         '# Pattern table', '',
         'Patterns count **relation sites once**; a site can appear in several table rows because several properties are written. The table groups by the exact property written, census source type and census target type. The source/target columns are relation components, not necessarily the whole expression. The 42 moved sites retain their original never component for traceability. Source predicates, writer predicates and typed function boundaries for all 81 original b/rest sites are in [guards.json](guards.json); locations below refer to the restored, uninstrumented full adapted tree. A static parameter contract or a constructor argument is labeled as provenance, not a runtime subtype check. A may-write graph witness does not establish that an actual value reaches every listed writer.', '']
text += [f"- {name}: **{n}**" for name,n in patterns.items()]
text += ['', '| Written property | Census source | Census target | Relation sites | Pattern | Check or provenance before write; writer locations |', '|---|---|---|---|---|---|']
for (prop, source, target), members in sorted(groups.items()):
    checks = list(dict.fromkeys(r['check'] for r in members))
    locations = sorted({w['where'] for r in members for w in r['writes'] if w['property'] == prop})
    if all(r['pattern'] == 'Reduced intersection component' for r in members):
        # List actual enclosing predicates, without confusing an absence check with a subtype proof.
        source_checks = sorted({c['where'] + ' (' + c['branch'] + ': ' + c['expression'] + ')' for r in members for c in guards[r['id']]['site']['conditions'] if ('is' in c['expression'] or '.kind' in c['expression'] or 'flags' in c['expression']) and len(c['expression']) < 220})
        checks = ['Impossible relation constituent; no executed value has this subtype. Whole-value predicates at ' + ('; '.join(source_checks) if source_checks else 'the typed source/writer boundaries in guards.json') + '.']
    record = {'property':prop,'source':source,'target':target,'sites':[r['id'] for r in members],'patterns':dict(Counter(r['pattern'] for r in members)),'checks':checks,'writers':locations}
    report['rows'].append(record)
    pattern = '; '.join(f'{name} ({n})' for name,n in sorted(record['patterns'].items()))
    text.append('| ' + ' | '.join(map(escape,[prop,source,target,', '.join(record['sites']),pattern,' '.join(checks)+' Writers: '+', '.join(locations)])) + ' |')
text += ['', '# Validation and limits', '',
         'Six mutants were caught: dropping a site fails the 47-site assertion; changing its AST context fails unique binding; removing an exit map fails the process-count assertion; removing one counter key fails the exact-key check; removing the emitted JavaScript increment changes the fresh self-check counter total from 552 to 0 while tsc output remains identical; removing the contradictory arm eliminates the reproducer\'s reported never relation. See evidence/mutants.json and individual logs. These are coverage/reporting checks; no adaptation or native compiler change is proposed.', '',
         'Setup completed in 41.935s on 5 processors (cgroup quota four): Go 0.042s, Node 0.034s, submodules 0.096s, clang 0.312s, Markdown dependencies 1.197s, build-cache warm 41.871s. The first setup overlapped the branch checkout and failed during cache warming; the settled-branch retry passed. The initial counter build collided with core.ts\'s local process declaration; the next import used a namespace receiver that Node\'s EventEmitter rejected. A default node:process import fixed both; successful suite results above use that final implementation. Failure logs are retained.', '',
         'Not covered: language-service or other unittest runners, native tsc, the complete Adamic gate, the five unobserved code sites on additional workloads, actual alias-reaching writes, or an independent proof of all source-to-writer ownership paths. The table explicitly preserves unproved casts, sentinel paths and union-alternative paths. No TypeScript checkout, generated compiler bundle or altered baseline is committed.', '',
         '# Reproduction', '', 'Source /workspace/adamic-tools/env.sh (or the setup-printed path), then run from the repository:', '', '```sh',
         'python3 stage3/adapt/75-optional-widening/coverage/run.py /tmp/NEW_COVERAGE_RUN > /tmp/coverage-run.log 2>&1',
         '```', '',
         'The runner creates a new scratch tree, preserves stage3/patch-set.md, builds upstream, validates all three workloads and queries both checkers. Additional static option comparisons and the tiny reproducer are run by probe.py; mutants.py consumes the resulting scratch paths. Rendering this document from the saved measurements uses python3 coverage/render.py. Check commands and byte hashes are recorded in evidence/provenance.json.', '']
(root / 'patterns.json').write_text(json.dumps(report,indent=2)+'\n')
(root / 'READ_WRITE.md').write_text('\n'.join(text))
print(json.dumps({'relation_sites':len(rows),'patterns':patterns,'grouped_rows':len(groups)}))
