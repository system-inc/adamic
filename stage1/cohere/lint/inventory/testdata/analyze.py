"""Derive package rankings and dispatch waves from the measured inventory."""
from collections import defaultdict, Counter
import json
from pathlib import Path
root = Path(__file__).resolve().parents[1]
data = json.loads((root/'inventory.json').read_text())
rules=data['rules']
packages=defaultdict(set)
for row in rules:
    for dep in row['dependencies']:
        if dep['kind']!='library': packages[dep['package']].add(row['name'])
ranked=sorted(packages.items(),key=lambda row:(-len(row[1]),row[0]))
ported=lambda row:any('ported with' in b['status'] or 'partial port:' in b['status'] for b in row['stage1'])
new=[row for row in rules if not ported(row)]
waves=Counter(row['wave'] for row in new)
lines=['# Porting pipeline','',f"Pinned inventory: {len(rules)} rules; {sum(ported(r) for r in rules)} have observed stage 1 selection sites; {len(new)} lack an observed implementation on available tips. Missing tips are unknown, so these are provisional dispatch counts.",'',
'## Observed branch status','']
for b in data['branches']: lines.append(f"- `{b['name']}`: {b['status']}, `{b.get('commit','unavailable')}`.")
lines += ['', 'The scanner tip documents 29 fully compared additions/baseline rules and one valid-source method-signature implementation with eight excluded malformed option/source combinations. Batch 2 adds twenty rules in disjoint files with rule_context.ts and batch2_registry.ts, for fifty observed implementations across those two tips. Its BATCH2.md reports 615 of 616 upstream cases compared and one explicitly refused malformed no-async-promise-executor case. Inherited REPORT.md describes the old five-rule baseline; BATCH2.md is the current batch report. Message tables are not implementations. Batch 3 appeared during this run and was fetched at fa9781c2c0911c52312002ba97ffd4b560d178ae. BATCH3.md adds the ten remaining frequency-selected rules, eight with full captured own-fixture coverage and two identifier rules blocked on nine JSX fixtures. Batch 4 remained unavailable at the final tip check; refresh that tip before dispatching presumed missing work.', '',
'## Land once before fan-out','',
'1. Freeze the parser/AST adapter, UTF-16 range convention, report/message records, fix edits, suggestion records and option/config error contract. The current parser is an indexed tree, not Go AST methods; adapter conformance must prove every mapped operation rather than hiding unsupported nodes. Let one integrator own dispatch/manifest and fix-conflict application. Preserve no-silent-clean behavior for unsupported input.',
'2. Lift shared syntax judgments in descending dependency count. Use the function ranking below and the complete inventory ranking, including transitive consumers. Common control-flow, scopes/references, token/comment collection, expression classification, static property values and JSX collectors each get one owner, one oracle suite and a versioned contract. Existing comments.ts, unicode.ts and Linter methods are starting evidence; batch 2 adds rule_context.ts and batch 3 adds rules/context.ts, rules/shared.ts, repair_edit.ts and repair_suggestion.ts. Reconcile those existing contracts before inventing another adapter; do not mark an entire Go package ported because one local method resembles it.',
'3. Land strict option decoding and message/settings adapters centrally. Each rule owns its schema/registration descriptor; the common decoder rejects invalid input rather than silently defaulting it. Required project options need fixtures with real settings; do not prioritize them using an unknown count interpreted as zero.',
'4. Build the symbol/type bridge in parallel with syntax helper work. Coarse NeedsTypeChecker is not a substitute for the exact per-rule reachable questions in inventory.md/JSON. Distinguish intrinsic error from any, constrained from unconstrained type, default-library symbol provenance, call/construct signatures, union/intersection parts, assignability, options and project resolution. A binding-only question is still a bridge prerequisite until symbol identity and scopes are proven.',
'5. Release a helper only after Go-versus-Node-versus-sanitized-native conformance and a compiling wrong-answer mutant. Then remove only its named blocker from dependent rules. Dispatch independent rule directories, with no worker writing shared lint.ts, global config or the dispatcher. Recompute waves at each integration tip.', '',
'## Helper packages ranked by distinct rule consumers','', '| Package | Consumers |', '|---|---:|']
for pkg,names in ranked: lines.append(f'| `{pkg}` | {len(names)} |')
lines += ['', '## Highest-use shared judgments to port first','', 'Counts include transitive consumers and deduplicate each rule. AST/reporting adapters above precede these algorithmic helpers. The complete function ranking is in inventory.md.', '', '| Function | Consumers |', '|---|---:|']
judgments={dep['symbol'] for r in rules for dep in r['dependencies'] if dep['kind'] in ('shared-judgment','scope','sibling','cohere')}
for helper in [h for h in data['helpers'] if h['symbol'] in judgments][:60]: lines.append(f"| `{helper['symbol']}` | {helper['rule_count']} |")
lines += ['', '## Dispatch waves','', 'Already observed ports are excluded from the following new-work queues. Ready means no unresolved shared algorithmic judgment detected after the common AST/reporting contract lands. It is not an assertion that every Go AST method already exists in Adamic. Interface dispatch and regex/standard-library mappings still require a source review before assignment. Helper blockers are conservative; exact API equivalence has not been proven by this inventory.', '']
for wave in ['syntax ready for AST/API adaptation','syntax waiting on helpers','type-aware / binding bridge']:
    rows=[r for r in new if r['wave']==wave]
    lines += [f'### {wave}: {len(rows)} rules','']
    for r in rows:
        if wave=='syntax waiting on helpers': reason='; '.join(r['waiting_on'])
        elif wave=='type-aware / binding bridge': reason='; '.join(r['checker_questions']) or 'NeedsTypeChecker declaration: inspect symbolic questions before dispatch'
        else: reason='no shared judgment blocker detected'
        lines.append(f"- `{r['name']}`: {reason}.")
    lines.append('')
lines += ['## Measurement boundaries','',f"Compiler: {len(data['corpora'].get('compiler',[]))} named files, {len(data['excluded'].get('compiler',[]))} parse exclusions. Repository: {len(data['corpora'].get('repository',[]))} tracked source files including deliberate negative fixtures, {len(data['excluded'].get('repository',[]))} parse exclusions.", '',
'Frequency is potential Go rule-API diagnostics under default/nil options and synthetic strict configuration, not the current repository lint gate. Failed rule runs and unavailable .a checker answers invalidate the whole per-rule count. Tailwind fixture/corpus tests failed for unavailable local theme repositories; their captured case counts are partial. This is useful prioritization evidence, not full fixture or native parity certification.', '',
'## Work ordering within a released wave','',
'First skip existing ports. Then favor high observed compiler/repository findings, many actual fixture configurations, and shared helpers that unlock many remaining rules. Keep source LOC and dependency count as scheduling estimates, not correctness metrics. Zero-firing rules still need positive controls; unknown frequencies cannot be sorted below zero. The filled prompts in PROMPTS.md cover a conditional fixer, configurable comments and a checker rule.']
(root/'PIPELINE.md').write_text('\n'.join(lines)+'\n')
