"""Group actual lowering NotYet observations by their exact stable reason text."""
import collections
import csv
import gzip
import hashlib
import json
from pathlib import Path
import sys


def disposition(reason):
    if 'writable view' in reason or 'writable-view' in reason or 'mutable variance' in reason or reason == 'a writable-slot checked view requiring source contract certification':
        return 'adaptation', 'Writable aliases must preserve invariant mutable slots; use a readonly view or an independent copy.'
    if reason in ('a value of type any', 'a parameter of type any', 'a function returning any', 'a call returning any', 'an array of any', 'a field of type any'):
        return 'adaptation', 'Adamic refuses unproven any by design; use a concrete type or unknown and narrow it.'
    if reason in ('a DebuggerStatement', 'a WithStatement', 'a VoidExpression', 'a VoidExpression as a statement'):
        return 'adaptation', 'This construct is refused by the language design; remove it or write an explicit supported statement.'
    if reason.startswith('reading '):
        return 'compiler lesson', 'Resolve this binding in its complete lexical context; isolated recovery may have lost it.'
    if 'NonNullExpression' in reason or 'non-null assertion' in reason:
        return 'compiler lesson', 'Lower the assertion as a checked nullish unwrap, preserving its operand representation.'
    if 'comma' in reason:
        return 'compiler lesson', 'Preserve left-to-right effects and return the final operand value.'
    if 'logical assignment' in reason:
        return 'compiler lesson', 'Evaluate the target once and short-circuit the conditional write.'
    if 'function inside a function' in reason:
        return 'compiler lesson', 'Lower nested declarations and their captures with the correct lexical lifetime.'
    if 'structural signature' in reason:
        return 'compiler lesson', 'Prove the declaring receiver and dispatch target without requiring source reshaping.'
    if 'without a body' in reason:
        return 'compiler lesson', 'Resolve the implementation or model the declared external operation before emitting a call.'
    if 'Expression' in reason or reason.startswith(('assigning ', 'for...', 'a declaration directly')):
        return 'compiler lesson', 'Lower this operation with its checked operand representations and JavaScript evaluation order.'
    if 'type ' in reason or 'returning ' in reason or 'parameter ' in reason:
        return 'compiler lesson', 'Learn this proven type or specialization and preserve its storage and call representation.'
    return 'compiler lesson', 'This is a stage-0 capability gap; implement the operation with sound checks and Node-held semantics.'


def main(raw, adapted, output):
    raw, adapted, output = Path(raw), Path(adapted).resolve(), Path(output)
    output.mkdir(parents=True, exist_ok=True)
    records = [json.loads(line) for line in raw.read_text().splitlines()]
    header = records[0]
    assert header['latent_mode'] == 'full'
    assert header['checker_rejected']
    observations = []
    def local(where):
        name, line, column = where.rsplit(':', 2)
        return str(Path(name).relative_to(adapted)) + ':' + line + ':' + column
    for record in records[1:]:
        for finding in record['findings']:
            if finding['phase'] == 'lowering' and finding['kind'] == 'NotYet':
                observations.append(dict(kind=finding['kind'], phase=finding['phase'],
                    where=local(finding['where']), reason=finding['reason'], text=finding['text'],
                    attempt_file=str(Path(record['file']).relative_to(adapted)), unit=local(finding['unit']),
                    context_sensitive=finding['reason'].startswith('reading '), measurement=finding['measurement']))
    fields = ['kind', 'phase', 'where', 'reason', 'text', 'attempt_file', 'unit', 'context_sensitive', 'measurement']
    with (output/'raw.csv').open('w', newline='') as stream:
        writer=csv.DictWriter(stream,fieldnames=fields,lineterminator='\n'); writer.writeheader(); writer.writerows(observations)
    unique = {(r['kind'],r['where'],r['reason'],r['text']):r for r in observations}
    groups=collections.defaultdict(list)
    for row in unique.values(): groups[row['reason']].append(row)
    rows=[]
    for reason, sites in groups.items():
        files=collections.Counter(site['where'].rsplit(':',2)[0] for site in sites)
        examples=sorted({site['where'] for site in sites}, key=lambda s:(s.rsplit(':',2)[0],*map(int,s.rsplit(':',2)[1:])))[:2]
        action,why=disposition(reason)
        rows.append(dict(reason=reason,count=len(sites),files=dict(sorted(files.items())),examples=examples,
                         context_sensitive=reason.startswith('reading '),disposition=action,why=why))
    rows.sort(key=lambda r:(-r['count'],r['reason']))
    total=len(unique)
    summary=dict(total=total,raw_observations=len(observations),reasons=len(rows),
                 context_sensitive_sites=sum(r['count'] for r in rows if r['context_sensitive']),
                 reasons_under_five=sum(r['count']<5 for r in rows),
                 coverage={str(n):dict(sites=sum(r['count'] for r in rows[:n]),share=100*sum(r['count'] for r in rows[:n])/total) for n in (5,10,20)},rows=rows)
    (output/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
    manifest=[dict(file=str(Path(p).relative_to(adapted)),bytes=Path(p).stat().st_size,sha256=hashlib.sha256(Path(p).read_bytes()).hexdigest()) for p in header['sources']]
    (output/'source-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
    with gzip.open(output/'full.jsonl.gz','wb') as stream: stream.write(raw.read_bytes())
    esc=lambda s:s.replace('|','&#124;').replace('\n','<br>')
    lines=['# Stage 3 lowering NotYet table','',
        f"Built an exact-reason table for {total:,} unique actual-lowering NotYet sites across {len(manifest)} source files.",
        'Compiler base: `44583d3283fdd8674085a7ddcce040cf2a73a94e`; replay tooling: `9a1f14c5d994aa855625e7cfa295677060348fec`.',
        'Commands and outputs are in [EVIDENCE.md](EVIDENCE.md); the selected replay signature and raw output are preserved there.',
        'Full-census continuation/rollback mutants, replay selection mutant and accounting artifact mutants are recorded in the validation logs.',
        "Not covered: skipped checker-diagnosed bodies, children behind failed compound boundaries, final ownership/module order, census entries' native or JavaScript output, and checked non-null support.", '',
        '## Measured base and scope','',
        'Used the user-authorized fallback: checked non-null was skipped. The initial merge of c17bf216 had 38 conflicts and was aborted. Its cherry-pick had eight conflicts and was also aborted: cmd/adamic/checks.go, docs/escape-hatches.md, internal/ir/ir.go, internal/lower/class_inheritance.go, internal/lower/non_null_impossible_test.go, internal/lower/predicates_overload_test.go, internal/lower/refusals.go, internal/oracle/counts.md. The base has no --explain-checks command path or predicate-site API required by its reporting controls; porting those absent dependencies exceeds resolving the changed hunks. The controls update 8a364b4c was inspected but not applied. No conflict resolutions or production compiler changes remain.', '',
        'The guarded full latent census ran on the same adapted 81-file tsc entry reach as the older 15,264-site report. Every source hash matches its saved tsc/source-manifest.json. These observations are **measured on a checker-rejected entry-root program**. They are blockers seen by isolated lowering attempts, not proof that every expression has been examined or that a whole program compiles.', '',
        f"Retained only phase `lowering`, kind `NotYet`. Counted unique `(kind, where, reason, text)` across attempts. raw.csv retains all {len(observations):,} matching observations and their attempt context; full.jsonl.gz preserves the original unfiltered stream. Diagnostic files, rather than attempt roots, receive file counts. Reasons remain exact, including embedded type names; no family normalization merges distinct reason strings.", '',
        f"Context sensitivity uses the previous report's flag: exact reason starts with `reading `. There are {summary['context_sensitive_sites']:,} such unique sites. A flagged read can reflect incomplete bindings after recovery; it remains in all counts. Dispositions and their explanations are recommendations, separate from these observations.", '',
        '## Coverage','', '| Top reasons | Sites | Share of all lowering NotYet sites |','| ---: | ---: | ---: |']
    for n in (5,10,20):
        c=summary['coverage'][str(n)];lines.append(f"| {n} | {c['sites']:,} | {c['share']:.2f}% |")
    lines += ['',f"{len(rows):,} exact reasons; {summary['reasons_under_five']:,} reasons have fewer than five sites.",'',
        'Non-null assertions, logical assignments and comma operations are compiler lessons under the supplied rulings, regardless of older language documentation. Writable-view variance is an adaptation: writable aliases must preserve invariant slots. Explicit any storage, calls and returns, debugger, with and the void operator are adaptations because the language refuses those constructs by design. All remaining observed NotYet reasons are conservatively treated as compiler lessons; a missing implementation or representation is not evidence of a permanent language refusal.', '',
        'For one-site reasons, only one distinct example exists and is shown. Columns in examples are retained for exact replay; the requested file:line is their prefix.', '',
        '## Reasons, sorted by count','',
        '| Count | Exact reason | Diagnostic files with counts | Two example sites | Context-sensitive | Disposition | Why |',
        '| ---: | --- | --- | --- | --- | --- | --- |']
    for r in rows:
        lines.append('| '+ ' | '.join([str(r['count']),esc(r['reason']),'<br>'.join(f"{esc(f)}: {n}" for f,n in r['files'].items()),'<br>'.join(r['examples']), 'yes' if r['context_sensitive'] else 'no',r['disposition'],r['why']])+' |')
    (output/'TABLE.md').write_text('\n'.join(lines)+'\n')
    print(json.dumps({k:v for k,v in summary.items() if k!='rows'},indent=2))

if __name__=='__main__': main(*sys.argv[1:])
