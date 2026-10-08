"""Render full latent evidence and independently audit stock-AST coverage and counts."""
import collections
import copy
import csv
import gzip
import hashlib
import json
from pathlib import Path
import re
import sys

sys.path.insert(0, str(Path(__file__).resolve().parents[2] / 'meter'))
from report import lowering_summary, MEASUREMENT, REJECTED_ENTRY_MEASUREMENT

KINDS = ('NotYet', 'Refused', 'SkippedDependency', 'error', 'panic', 'Boundary')
GIANTS = ('createTypeChecker', 'createNodeFactory', 'createPrinter', 'transformES2015',
          'createProgram', 'createBinder', 'transformClassFields', 'createScanner')


def read_rows(path):
    if path.exists():
        return [json.loads(line) for line in path.read_text().splitlines()]
    with gzip.open(str(path) + '.gz', 'rt') as stream:
        return [json.loads(line) for line in stream]


def intervals_size(spans):
    result, end = 0, -1
    for start, stop in sorted(spans):
        if stop > max(start, end):
            result += stop - max(start, end)
        end = max(end, stop)
    return result


def relative(name, original):
    return str(Path(name).relative_to(original))


def local_where(where, original):
    if not where:
        return ''
    name, line, column = where.rsplit(':', 2)
    return relative(name, original) + ':' + line + ':' + column


def unique(rows, phase=None):
    return {(f['kind'], f['where'], f['reason'], f['text'])
            for row in rows[1:] for f in row['findings']
            if phase is None or f['phase'] == phase}


def audit_units(rows, stock, original):
    assert rows[0].get('latent_mode') == 'full', 'full mode'
    assert {relative(row['file'], original) for row in rows[1:]} == set(stock), 'stock coverage'
    diagnostics = rows[0]['diagnostic_sites']
    for row in rows[1:]:
        name = relative(row['file'], original)
        observed = {local_where(unit['where'], original): unit for unit in row['units']
                    if unit['kind'] == 'KindFunctionDeclaration'}
        expected = {unit['where']: unit for unit in stock[name]}
        assert len(observed) == len([u for u in row['units'] if u['kind']=='KindFunctionDeclaration']), 'duplicate declaration'
        assert set(observed) == set(expected), 'stock declaration coverage'
        for where, unit in observed.items():
            external = expected[where]
            assert (unit.get('name',''), unit.get('depth',0), unit.get('body_start',0),unit.get('body_end',0)) == (
                external['name'], external['depth'],external['body_start'],external['body_end']), 'stock declaration span'
            if external['parent_function']:
                assert local_where(unit['parent'],original) == external['parent_function'], 'stock parent'
            if not external['body_end']:
                continue
            children = [child for child in stock[name] if child['parent_function'] == where]
            owned = [d['text'] for d in diagnostics if d['file'] == row['file']
                     and d['start'] < external['body_end']
                     and (d['end'] > external['body_start'] or d['start'] >= external['body_start'])
                     and not any(d['start'] >= child['start'] and d['end'] <= child['end'] for child in children)]
            assert unit['checker_diagnostics'] == owned, 'direct diagnostic ownership'
            assert (unit['status'] == 'split_checker_body') == bool(owned), 'direct eligibility'
            if owned:
                assert not any(f['unit'] == unit['where'] for f in row['findings']), 'diagnosed body entered'


def audit_counts(summary, rows, owners):
    sites = unique(rows)
    counts = collections.Counter(site[0] for site in sites)
    assert summary['totals'] == {kind: counts[kind] for kind in KINDS if kind != 'Boundary'}, 'unique totals'
    assert summary['boundaries'] == counts['Boundary'], 'boundary count'
    reasons = collections.Counter(site[0]+': '+site[2] for site in sites if site[0] in ('NotYet','Refused'))
    assert summary['per_reason'] == dict(reasons), 'reason counts'
    for row in summary['reason_rows']:
        reason=row['reason']
        matching=[prefix for prefix in owners if prefix!='seen as' and reason.startswith(prefix)]
        owner=owners.get('seen as','OWNER BLANK') if ' seen as ' in reason and 'seen as' in owners else (
            owners[max(matching,key=len)] if matching else 'OWNER BLANK')
        assert row['owner']==owner, 'owner assignment'
        assert row['NotYet']==reasons['NotYet: '+reason] and row['Refused']==reasons['Refused: '+reason], 'ranked reason counts'


def summarize_body(unit, all_sites, lower_sites, root):
    path, line, col = unit['where'].rsplit(':',2)
    data=Path(path).read_bytes()
    # Diagnostic columns are UTF-16. Convert their positions to byte offsets.
    lines=data.decode().splitlines(keepends=True)
    starts=[];total=0
    for text in lines: starts.append(total);total+=len(text.encode())
    def contained(site):
        if not site[1]:return False
        name,row,column=site[1].rsplit(':',2)
        if name != path:return False
        text=lines[int(row)-1]; wanted=int(column)-1; characters=0; consumed=0
        for character in text:
            if consumed>=wanted:break
            consumed+=2 if ord(character)>0xffff else 1;characters+=1
        position=starts[int(row)-1]+len(text[:characters].encode())
        return unit.get('declaration_start',unit['body_start'])<=position<unit['body_end']
    keys={site for site in all_sites if contained(site)}
    lowered={site for site in lower_sites if contained(site)}
    counts=collections.Counter(site[0] for site in keys)
    lower_counts=collections.Counter(site[0] for site in lowered)
    return {'name':unit['name'],'where':local_where(unit['where'],root),'bytes':unit['body_end']-unit['body_start'],
            'totals':{kind:counts[kind] for kind in KINDS},
            'actual_lowering':{kind:lower_counts[kind] for kind in KINDS},
            'per_reason':dict(collections.Counter(site[0]+': '+site[2] for site in keys if site[0] in ('NotYet','Refused')))}


def audit_same_checker(baseline, full):
    assert sorted(baseline[0]['diagnostics'])==sorted(full[0]['diagnostics']), 'same checker diagnostics'
    key=lambda site:(site['file'],site['start'],site['end'],site['text'])
    assert sorted(map(key,baseline[0]['diagnostic_sites']))==sorted(map(key,full[0]['diagnostic_sites'])), 'same checker spans'


def write_csv(path, fields, rows):
    with path.open('w',newline='') as output:
        writer=csv.DictWriter(output,fieldnames=fields,lineterminator='\n');writer.writeheader()
        writer.writerows({key:row.get(key,'') for key in fields} for row in rows)


def build(adapted, run):
    owners=json.loads(Path(__file__).resolve().parents[2].joinpath('meter/owners.json').read_text())
    compiler_full=read_rows(run/'compiler/full.jsonl')
    original=Path(compiler_full[1]['file'].split('/src/')[0])
    result={'count_definition':'unique (kind, where, reason, text), refusal scan and actual lowering combined; recovery boundaries separate',
            'original_adapted':str(original), 'scopes':{}}
    for scope in ('compiler','tsc'):
        rows=compiler_full if scope=='compiler' else read_rows(run/'tsc/full.jsonl')
        manifest=json.loads((run/scope/'source-manifest.json').read_text())
        for entry in manifest:
            assert hashlib.sha256((adapted/entry['file']).read_bytes()).hexdigest()==entry['sha256'], 'source hash'
        stock=json.loads((run/scope/'stock-units.json').read_text())
        audit_units(rows,stock,original)
        print(scope+': stock declaration and diagnostic ownership audit passed',flush=True)
        label=MEASUREMENT if scope=='compiler' else rows[0]['measurement']
        expected={str(original/entry['file']) for entry in manifest}
        full=lowering_summary(rows,expected,label);audit_counts(full,rows,owners)
        all_sites=unique(rows);lower_sites=unique(rows,'lowering')
        lower_counts=collections.Counter(site[0] for site in lower_sites)
        full['actual_lowering_totals']={kind:lower_counts[kind] for kind in KINDS}
        units=[dict(unit,file=row['file']) for row in rows[1:] for unit in row['units']]
        excluded=[unit for unit in units if unit['status']=='split_checker_body']
        nested=[unit for unit in excluded if unit.get('depth',0)>0]
        by_file=collections.defaultdict(list)
        for unit in nested:by_file[unit['file']].append((unit['body_start'],unit['body_end']))
        full['context_sensitive_reading_sites']=sum(row['NotYet'] for row in full['reason_rows'] if row['reason'].startswith('reading '))
        full['excluded_nested_count']=len(nested)
        full['excluded_nested_gross_bytes']=sum(u['body_end']-u['body_start'] for u in nested)
        full['excluded_nested_union_bytes']=sum(intervals_size(spans) for spans in by_file.values())
        full['unit_counts']=dict(collections.Counter(u['status'] for u in units))
        full['function_declarations']=sum(u['kind']=='KindFunctionDeclaration' for u in units)
        external_units={item['where']:item for items in stock.values() for item in items}
        for unit in units:
            if unit['kind']=='KindFunctionDeclaration':
                external=external_units[local_where(unit['where'],original)]
                unit['declaration_start']=external['start']
        full['giants']=[summarize_body(unit,all_sites,lower_sites,original)
                        for name in GIANTS for unit in units
                        if unit.get('name')==name and not unit.get('depth',0) and unit.get('body_end')]
        baseline=None
        if scope=='compiler':
            old=read_rows(run/'compiler/baseline.jsonl');audit_same_checker(old,rows);baseline=lowering_summary(old,expected,MEASUREMENT)
            changed=list(old);changed[0]=dict(old[0],diagnostics=old[0]['diagnostics']+['planted diagnostic'])
            try:audit_same_checker(changed,rows)
            except AssertionError as error:assert str(error)=='same checker diagnostics';print('compiler: checker comparison mutant caught')
            else:raise AssertionError('checker comparison mutant survived')
            audit_counts(baseline,old,owners)
            old_units=[u for row in old[1:] for u in row['units'] if u.get('body_end')]
            skipped=[u for u in old_units if u['status']=='skipped_checker_body']
            baseline['top_level_body_bytes']=sum(u['body_end']-u['body_start'] for u in old_units)
            baseline['skipped_bodies']=len(skipped)
            baseline['skipped_body_bytes']=sum(u['body_end']-u['body_start'] for u in skipped)
        result['scopes'][scope]={'full':full,'baseline':baseline,'checker_diagnostics':len(rows[0]['diagnostics'])}
        write_csv(run/scope/'reasons.csv',['reason','owner','unowned','context_sensitive_read','NotYet','Refused','count'],
                  [dict(row,unowned=row['owner']=='OWNER BLANK',context_sensitive_read=row['reason'].startswith('reading ')) for row in full['reason_rows']])
        giant_rows=[]
        owner_by_reason={row['reason']:row['owner'] for row in full['reason_rows']}
        for giant in full['giants']:
            groups={}
            for key,count in giant['per_reason'].items():
                kind,reason=key.split(': ',1)
                group=groups.setdefault(reason,{'function':giant['name'],'where':giant['where'],'reason':reason,'owner':owner_by_reason[reason],'unowned':owner_by_reason[reason]=='OWNER BLANK','NotYet':0,'Refused':0,'count':0})
                group[kind]+=count;group['count']+=count
            giant_rows.extend(sorted(groups.values(),key=lambda row:(-row['count'],row['reason'])))
        write_csv(run/scope/'giant-reasons.csv',['function','where','reason','owner','unowned','NotYet','Refused','count'],giant_rows)
        write_csv(run/scope/'excluded-nested.csv',['name','where','file','depth','body_start','body_end','bytes','checker_diagnostics'],
                  [dict(unit,where=local_where(unit['where'],original),file=relative(unit['file'],original),
                        bytes=unit['body_end']-unit['body_start'],checker_diagnostics='\n'.join(unit['checker_diagnostics'])) for unit in nested])
        write_csv(run/scope/'boundaries.csv',['unit','where','reason','text','start','end'],
                  [dict(f,unit=local_where(f['unit'],original),where=local_where(f['where'],original))
                   for row in rows[1:] for f in row['findings'] if f['kind']=='Boundary'])
        # Prove audits reject count and missing-unit mutations using real observations.
        mutant=copy.deepcopy(full);mutant['totals']['NotYet']+=1
        try:audit_counts(mutant,rows,owners)
        except AssertionError as error:assert str(error)=='unique totals';print(scope+': count mutant caught')
        else:raise AssertionError('count mutant survived')
        mutant=list(rows);index=next(i for i,row in enumerate(rows[1:],1) if any(u.get('depth',0)>0 for u in row['units']));mutant[index]=dict(rows[index],units=list(rows[index]['units']));target=mutant[index]
        target['units'].remove(next(u for u in target['units'] if u.get('depth',0)>0))
        try:audit_units(mutant,stock,original)
        except AssertionError as error:assert str(error)=='stock declaration coverage';print(scope+': missing nested unit mutant caught')
        else:raise AssertionError('nested unit mutant survived')
        if full['reason_rows']:
            mutant=copy.deepcopy(full);mutant['reason_rows'][0]['owner']='forged owner'
            try:audit_counts(mutant,rows,owners)
            except AssertionError as error:assert str(error)=='owner assignment';print(scope+': owner mutant caught')
            else:raise AssertionError('owner mutant survived')
    (run/'report.json').write_text(json.dumps(result,indent=2)+'\n')
    old=result['scopes']['compiler']['baseline'];full=result['scopes']['compiler']['full'];entry=result['scopes']['tsc']['full']
    lines=['Main 74fb6490 latent full evidence', '',
           'Measurement-only overlay. Production compiler files are unchanged. No backend was invoked.', '',
           '| Scope | Legacy NotYet | Full NotYet | Legacy Refused | Full Refused | Recovery boundaries |',
           '| --- | ---: | ---: | ---: | ---: | ---: |',
           f"| Compiler files ({full['source_files']}) | {old['totals']['NotYet']} | {full['totals']['NotYet']} | {old['totals']['Refused']} | {full['totals']['Refused']} | {full['boundaries']} |",
           f"| tsc entry reach ({entry['source_files']}) | checker blocked | {entry['totals']['NotYet']} | checker blocked | {entry['totals']['Refused']} | {entry['boundaries']} |", '',
           f"Legacy skips {old['skipped_bodies']} top-level bodies: {old['skipped_body_bytes']:,} / {old['top_level_body_bytes']:,} bytes ({100*old['skipped_body_bytes']/old['top_level_body_bytes']:.2f}%).", '',
           '| Giant body | Bytes | Full NotYet | Full Refused | Actual-lowering NotYet | Actual-lowering Refused | Boundaries |',
           '| --- | ---: | ---: | ---: | ---: | ---: | ---: |']
    for giant in full['giants']:
        lines.append(f"| {giant['name']} ({giant['where']}) | {giant['bytes']:,} | {giant['totals']['NotYet']} | {giant['totals']['Refused']} | {giant['actual_lowering']['NotYet']} | {giant['actual_lowering']['Refused']} | {giant['totals']['Boundary']} |")
    for scope in ('compiler','tsc'):
        summary=result['scopes'][scope]['full']
        lines += ['',f"{scope}: {summary['measurement']}; {result['scopes'][scope]['checker_diagnostics']} checker diagnostics.",
                  f"{summary['function_declarations']} function declarations; unit statuses: {summary['unit_counts']}.",
                  f"Nested checker exclusions: {summary['excluded_nested_count']}; {summary['excluded_nested_gross_bytes']:,} gross body bytes; {summary['excluded_nested_union_bytes']:,} union bytes.",
                  f"Actual-lowering-only totals: {summary['actual_lowering_totals']}.",
                  f"Context-sensitive reading-X sites: {summary['context_sensitive_reading_sites']}; retained in the totals, flagged in reasons.csv. These can reflect incomplete isolated bindings after an earlier failure and need individual interpretation.", '',
                  '| Reason | Owner | NotYet | Refused | Total |','| --- | --- | ---: | ---: | ---: | ---: |']
        for row in summary['reason_rows'][:10]:
            reason=row['reason'].replace('|','&#124;').replace('\n',' ')
            lines.append(f"| {reason} | {row['owner']} | {row['NotYet']} | {row['Refused']} | {row['count']} |")
        lines += ['',f"Complete reasons, with every unowned row flagged, are in {scope}/reasons.csv; excluded names and diagnostics are in {scope}/excluded-nested.csv; all recovery boundaries are in {scope}/boundaries.csv."]
    lines += ['', 'Limits: a failed compound statement can prevent examination of its children. Signature/prologue failures can prevent examination of the remaining body. Nested declarations are still independently attempted. Boundary spans identify affected constructs and may overlap; they are not an exact count of unexamined bytes. Gross excluded-body bytes may overlap; union bytes avoid double counting. Isolated ancestor bindings and sibling signatures are observational context, not proof of closure support. Giant counts include signature sites and nested bodies within the giant declaration. Findings combine refusal scanning and actual lowering unless explicitly marked otherwise. Final ownership, module order, and backends were not tested.']
    (run/'report.md').write_text('\n'.join(lines)+'\n')
    print('PASS: source hashes, stock AST declaration coverage/spans/parents, direct diagnostic ownership, exact unique totals/reasons/owners, named exclusions')
    return result


if __name__=='__main__':
    build(Path(sys.argv[1]).resolve(),Path(sys.argv[2]).resolve())
