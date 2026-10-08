#!/usr/bin/env python3
"""Render unresolved public-owner choices without turning observations into contracts."""
import collections
import json
from pathlib import Path
H=Path(__file__).resolve().parent
ledger=json.loads((H/'evidence/contracts.json').read_text());obs=json.loads((H/'evidence/observations.json').read_text())
ledger['publicOwners']=[p for p in ledger['publicOwners'] if p['file'].startswith(('src/compiler/','src/services/'))]
by=obs['by_site']
public_map={
 'S01':['convertToObject','JsonConfigValue'], 'S02':['CompilerOptionsValue','parseJsonConfigFileContent','convertCompilerOptionsFromJson'],
 'S03':['parseJsonConfigFileContent','ParsedCommandLine'], 'S04':['parseJsonConfigFileContent','ParsedCommandLine'],
 'S05':['ParsedCommandLine','parseJsonConfigFileContent'], 'S06':['parseJsonConfigFileContent','ParsedCommandLine'],
 'S07':['parseJsonConfigFileContent','ParsedCommandLine'], 'S08':['convertCompilerOptionsFromJson'],
 'S09':['convertTypeAcquisitionFromJson'], 'S10':['convertCompilerOptionsFromJson','parseJsonConfigFileContent'],
 'S11':['convertTypeAcquisitionFromJson','parseJsonConfigFileContent'], 'S12':['parseJsonConfigFileContent','ParsedCommandLine'],
 'S13':['convertCompilerOptionsFromJson','convertTypeAcquisitionFromJson','parseJsonConfigFileContent'],
 'S14':['convertCompilerOptionsFromJson','convertTypeAcquisitionFromJson','parseJsonConfigFileContent'],
 'S15':['convertCompilerOptionsFromJson','convertTypeAcquisitionFromJson','parseJsonConfigFileContent'],
 'S16':['System'], 'S17':['System','WatchHost'], 'S18':['System','WatchHost'],
 'S19':['readConfigFile','parseConfigFileTextToJson','readJson','readJsonOrUndefined','JsonConfigValue'],
 'S20':['Node','Type','ObjectAllocator'],
}
json_check='An explicit checked raw-value boundary must accept invalid option values long enough to preserve tsc diagnostics, and validate only the requested field when consumed. A whole JSON assertion excludes supported JavaScript values, accessors and cycles. Cost: O(1) per scalar use, O(elements/fields) for a recursive structural check; repeated checks/copies can add traversal and allocation. A panic on unsupported input changes the API failure contract and needs a ruling.'
json_wrap='Carry a caller-specific Raw<C> separately from normalized CompilerOptions/TypeAcquisition/WatchOptions, with internal field conversions returning explicit validated unions. Use the existing recursive JsonConfigValue only for parser-produced values. Generic preservation can erase to zero added JavaScript; field verification and dynamic dictionary representation remain required. No unknown-to-final-type assertion or alias for any is allowed. Public any entrypoints remain refused unless a checked entry shim or a public sanction is approved.'
refuse='Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.'
for row in ledger['rows']:
 id=row['id'];row['public_declarations']=[{k:p[k] for k in ['name','file','line','signature']} for p in ledger['publicOwners'] if p['name'] in public_map[id]]
 row['decision']='undecided'
 runtime_name=('timer.setTimeout' if id in ['S17','S18'] else 'timer.clearTimeout' if id=='S16' else 'allocator.getNodeConstructor' if id=='S20' else row['owner_name'])
 event=by.get(runtime_name)
 row['acceptance_observation']=dict(site=runtime_name,projects=len(event['projects']) if event else 0,phases=event['phases'] if event else {},domains=event['domains'] if event else [],scope='301 noEmit projects only; zero means not exercised, never a universal absence claim')
 if id in ['S17','S18']:
  check='Keep an opaque, checked host-handle representation paired with the exact registration/cancellation provider; verify callback arity/domain at dispatch. Node Timeout and numeric/custom hosts must coexist. Cost: tag/owner comparisons O(1) per start/stop plus a retained provider/handle pair; Node-specific inspection excludes other hosts. Public check/failure semantics need approval.'
  wrap='Thread TimerHost<H,A> through System/WatchHost, queue state, builder state and cancellation; internal storage is H | undefined (dynamic polling queue additionally stores false). Zero runtime cost if specialization and lifetime ownership are proven. Cross-host cancellation is rejected statically. Public erased methods still require a checked ABI shim or sanction; timers also require callback lifetime/ownership support.'
 elif id=='S16':
  check='Do not add a check to paper over this observation: source already says NodeJS.Timeout | undefined. Stock checker agrees. First diagnose the latent checker external declaration resolution. Cost: investigation only; no unsupported source any is claimed.'
  wrap='An internal concrete Node timer type is already present here. Wrapping it does not fix the unexplained latent any; preserve the anomaly and use System/WatchHost generic handles only at their actual public declarations.'
 elif id=='S20':
  check='A completed-Node check immediately after allocation would reject actual headers: parent starts undefined and text/symbol/id can be absent. A loud read-before-initialization check at the first required-field read could preserve successful paths. Cost: O(1) initialized/tag tests per unproven read plus state tracking; escaping aliases and parent lifetime must be modeled. Arbitrary descriptor/prototype checks are not an ownership proof.'
  wrap='Separate AllocatedNode/Header from completed Node, and TypeHeader from Type; factories prove field completion before exposing the completed interface. Public constructor-return any bridges remain refused until that staged owner relationship is approved. Zero runtime overhead only where initialization is proven; late or optional completion needs flags/checks. Do not annotate an incomplete constructor as already complete.'
 elif id in ['S01','S19']:
  check='Check/represent recursive JSON plus recovery undefined at the boundary, preserving number/string/boolean/null/array roots and existing readJson fallback on falsy values. Cost: O(1) root tag and O(nodes) if a whole result must be validated. Object-only checks reject observed primitive returns. Loose recovery/error diagnostics and getters at public raw-input APIs need separate policies.'
  wrap='Give internal parser/result owners JsonConfigValue and propagate it through readJsonOrUndefined/readJson and their consumers. A wrapper calling the proven internal parser can erase to zero JavaScript; calling an any-returning public API still requires a sound check. Public convertToObject/config results require a sanction or a separate checked ABI shim; casts to object are false.'
 else:check=json_check;wrap=json_wrap
 if id=='S02':check='Do not strengthen the existing shallow array predicate and claim byte equality: [{}] currently returns true. Validate elements at the conversion/use boundary, preserving invalid-input diagnostics, or explicitly approve a changed failure policy. Cost: O(elements) checks. '+check
 row['options']={'loud_boundary_check':check,'internal_only_wrapper':wrap,'leave_refused':refuse}
 row['supplemental_member_ledger']= ['setTimeout','clearTimeout'] if id in ['S16','S17','S18'] else ['getNodeConstructor','getTokenConstructor','getIdentifierConstructor','getPrivateIdentifierConstructor','getSourceFileConstructor','getTypeConstructor'] if id=='S20' else []
 if id=='S18':row['supplemental_member_ledger']+=['timerToBuildInvalidatedProject']
 if id=='S17':row['supplemental_member_ledger']+=['pollScheduled']
(H/'contracts.json').write_text(json.dumps(ledger,indent=2)+'\n')
text='''# Step 09: public any contracts, decisions open

Measured on TypeScript 6.0.3, pin 050880ce59e30b356b686bd3144efe24f875ebc8, with the adapted pipeline from origin/codex/stage3-explicit-any-2 at 1e920a31b601422c75a26ead083e9c1423c34b66. That branch supplies RESIDUE.md; current main does not contain it. This scout changes no adaptation, compiler or public declaration. Main/compiler base is 45487a809f89885a3fc651cd590e7dabf31362dc.

The residue has **19 source contracts**: 16 JSON/config/predicate observations, two host timer storage observations and one staged allocator observation. A twentieth observation is a concrete Node timer annotation that the latent checker reports any; it is retained as an anomaly, not counted as an unedited any. Three option-conversion overload rows share one implementation and are not three independent execution sites.

## What Node actually did

The instrumented scratch tsc CLI matched unchanged stdout, stderr and exit goldens on **301/301** acceptance projects in 121.311 seconds. The existing driver runs --noEmit: it neither watches nor builds projects. All 301 project captures and their exact shape counts are in evidence/project-captures.json.gz; evidence/observations.json is their aggregate. Four public-only functions are tree-shaken out of the CLI (see instrumentation.json). No zero-hit function is declared safe or impossible.

The raw-config worker's json argument was undefined in all 301 runs: this corpus enters through a JsonSourceFile, not the arbitrary JavaScript public json path. It returns ParsedCommandLine objects with raw data and normalized options. The list predicate ran 2,608 times: values were 301 empty arrays, 312 string-element arrays, 765 booleans, 929 strings and 301 object-shaped raw option values. Every result was boolean; these observations do not prove the declared predicate for invalid external arrays.

The CLI called getNodeConstructor 7,280 times across all 301 projects. Bare Node headers completed 18,204,760 allocations with parent undefined; Token/Identifier headers also had parent undefined. Type headers completed 87,197 allocations with only flags and no symbol. This is the allocation stage, not the state after factory/checker initialization. The counts include standard-library parsing repeated by separate CLI processes.

No setTimeout/clearTimeout registration or cancellation was observed in these 301 noEmit projects. Focused API calls, explicitly separate from acceptance coverage, returned a Node Timeout object and cancelled the same handle. The timer fixture also uses a numeric custom-host handle. The acceptance corpus cannot establish all custom-host domains or callback ownership.

Focused actual API calls returned primitive number, boolean and string through readJson/readJsonOrUndefined and convertToObject; convertToObject also returned null/arrays. readJson's existing || {} changes null and false to {}. Malformed input recovers differently through strict/loose readers. The full API uses NodeObject as its allocator provider; its parent is undefined too. evidence/counterexamples.json distinguishes those API calls from bare CLI constructor observations.

## Rules and interpretation

The accepted escape-hatch document says: "As written: any, as unknown as, expando additions, Object.defineProperty and Function are refused" and requires proven predicate bodies. A loud boundary check below is a proposed checked representation/ABI decision, not permission to admit any or cast unknown to an interface. A check that cannot prove initialization, ownership or callable argument relationships must leave the value refused. Costs below are inferred operation/allocation costs, not benchmark results. Nothing is decided in this scout.

## Caller coverage

contracts.json records the exact adapted file:line:column, declaration/signature, every resolved internal source reference and direct call, and linked public owner declarations. It scans all src, including services, server, tools, harness and tests; test callers are recorded alongside production callers. Shorthand captured methods are resolved by their value symbol. Structural timer/allocator member names have a separate complete member-reference ledger with context and test/production labels; name matches are candidate uses, not a claim that structurally distinct hosts share one symbol. External application callers cannot be enumerated by an internal census.

Every reference and all declaration text are retained in contracts.json, including function values passed onward. The Markdown below lists all source-symbol references and links to supplemental member lists. Public-owned raw fields still permit arbitrary JavaScript, including invalid option values. A recursive JSON type is sound for parser-produced JSON, but cannot silently replace that public input contract.

'''
for r in ledger['rows']:
 text+=f'## {r["id"]}: {r["file"]}:{r["line"]}:{r["column"]}\n\n{r["residue"]}\n\nSource owner: {r["owner_name"]}, {r["owner"]["file"]}:{r["owner"]["line"]}.\n\n```typescript\n{r["signature"]}\n```\n\nPublic/dependent declaration owners:\n\n'
 for p in r['public_declarations']:
  sig=p['signature']
  if p['name'] in ['System','WatchHost','Node','Type','ParsedCommandLine'] and sig.startswith(('export interface','interface')):
   relevant=[line for line in sig.splitlines() if any(w in line for w in ['interface ','setTimeout','clearTimeout','parent:','symbol:','raw?:'])]
   sig='\n'.join(relevant)+'\n// Full declaration retained in contracts.json.'
  text+=f'- {p["name"]}: {p["file"]}:{p["line"]}.\n\n```typescript\n{sig}\n```\n\n'
 text+='Internal callers/references:\n\n'
 if not r['references']:text+='No resolved source-symbol reference. For allocator callback/member invocation use the supplemental structural member ledger below; an uncalled public-only API is still an external contract.\n\n'
 for ref in r['references']:text+=f'- {ref["file"]}:{ref["line"]}:{ref["column"]}: {ref["use"]}, `{ref["text"]}`.\n'
 if r['references']:text+='\n'
 for member in r['supplemental_member_ledger']:
  text+=f'Member ledger `{member}` ({len(ledger["memberReferences"][member])} mentions, including declarations and tests):\n\n'
  for ref in ledger['memberReferences'][member]:text+=f'- {ref["file"]}:{ref["line"]}:{ref["column"]}, {ref["domain"]}.\n'
  text+='\n'
 ob=r['acceptance_observation'];text+=f'Acceptance observation: {ob["projects"]}/301 projects; phases {json.dumps(ob["phases"],sort_keys=True)}. '
 text+=('No execution observed; see the focused fixtures/counterexamples for any additional facts.' if not ob['projects'] else 'All observed domains are retained under this row in contracts.json; no universal input-domain inference is made.')+'\n\n'
 for label,value in r['options'].items():text+=f'- **{label.replace("_"," ")}**: {value}\n'
 text+='\nDecision: **undecided**.\n\n'
text+='''## Fixtures, mutants and limits

Three extracted .a functions are run with source Node, with independent exact expected stdout, stderr and exit. convertToObject preserves primitive/array/object roots; isCompilerOptionsValue demonstrates the shallow invalid-array acceptance; scheduleBuildInvalidatedProject forwards callback arguments and correlates numeric/Node handles. Each has a successfully executing source mutant rejected only by stdout: returnValue false, array predicate false, and cancellation with undefined. fixtures/status.json records current-main compile observations and exact diagnostics. No native success is claimed. These are scoped scout probes, not additions to internal/oracle or shared fixtures_test.go; local counts.md records the three fixtures and three mutants.

The fixture bodies are cut from original owners without statement rewriting. Minimal dependency interfaces and drivers are scaffolding, not approved contracts. The public-any-ts loader resolves only the Node fixture dependency to the freshly built 6.0.3 API. Native checker failure at that module is recorded honestly. Current Adamic ambient timer declarations are missing too. No native memory/ownership proof is made.

Runtime shape inspection uses own property descriptors (without invoking getters), arrays and depth-two summaries, with counts aggregated per process. It returns every value unchanged, and the 301 goldens verify the tested outputs. Proxy traps, getter behavior, object identity visible through stack inspection, unsupported hosts, watch/build mode, dynamic external callers and deeper alias/ownership contracts remain unmeasured. The instrumented timing is not a boundary-check performance estimate. No full gate or upstream full test suite was run.

## Questions for @system_adamic

1. May public convertToObject/config results acquire a sanctioned recursive JSON-plus-recovery type, or must a separate checked public/native entry shim preserve the exact upstream declarations?
2. Should arbitrary JavaScript raw-config input remain supported natively (including invalid values, accessors and cycles), and must checks feed existing diagnostics rather than panic? Where is the observable failure boundary?
3. Should internal raw-to-normalized conversion carry caller-specific generics, while the shallow predicate remains exactly as written? Who owns the predicate's current false promise for list elements?
4. Is TimerHost<H,A> allowed as a sanctioned public specialization, or should native use a checked opaque handle paired with its original provider? What are the cancellation and callback lifetime guarantees for custom hosts?
5. Should allocated AST/type headers have a separate internal type until completion, or use checked read-before-initialization fields? When may a partially initialized header escape to an API/client?
6. Who diagnoses the concrete NodeJS.Timeout annotation reported any by the latent checker? This scout does not treat it as another public any adaptation.

All choices remain undecided. This report sends no message to @system_adamic.
'''
(H/'REPORT.md').write_text('\n'.join(line.rstrip() for line in text.splitlines())+'\n')
print('rendered 20 observations, including 19 source contracts; decisions remain open')
