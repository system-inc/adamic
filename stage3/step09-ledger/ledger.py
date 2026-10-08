"""Reconcile compiler-API inventories through exact incremental source edits."""
import argparse
from collections import Counter, defaultdict
from difflib import SequenceMatcher
import json
from pathlib import Path


def units(text):
    raw=text.encode('utf-16-le')
    return ''.join(chr(int.from_bytes(raw[i:i+2],'little')) for i in range(0,len(raw),2))


def equal_ranges(before,after):
    # Match lines first, refining only changed blocks. Positions remain UTF-16.
    a,b=before.splitlines(keepends=True),after.splitlines(keepends=True)
    aa=[0];bb=[0]
    for line in a:aa.append(aa[-1]+len(line))
    for line in b:bb.append(bb[-1]+len(line))
    result=[]
    for tag,i,j,k,l in SequenceMatcher(None,[line.rstrip("\r\n") for line in a],[line.rstrip("\r\n") for line in b],autojunk=False).get_opcodes():
        if tag=='equal':
            for ai,bi in zip(range(i,j),range(k,l)):
                result.append((aa[ai],aa[ai+1] if a[ai]==b[bi] else aa[ai]+len(a[ai].rstrip('\r\n')),bb[bi]))
        elif tag=='replace':
            left,right=before[aa[i]:aa[j]],after[bb[k]:bb[l]]
            if max(len(left),len(right))>20000:raise AssertionError('large unmatched source block needs explicit correspondence')
            for match in SequenceMatcher(None,left,right,autojunk=False).get_matching_blocks():
                if match.size:result.append((aa[i]+match.a,aa[i]+match.a+match.size,bb[k]+match.b))
    compact=[]
    for start,end,to in result:
        if compact and compact[-1][1]==start and compact[-1][2]+start-compact[-1][0]==to:
            compact[-1]=(compact[-1][0],end,compact[-1][2])
        else:compact.append((start,end,to))
    return compact


def mapped(start,end,ranges):
    for a,b,to in ranges:
        if a<=start and end<=b:return start+to-a,end+to-a
    return None


def reconcile(stock,adapted,transitions,proofs, *, ignore_rewrites=False):
    adapted_index={(s['file'],s['kind'],s['start']):s for s in adapted['sites']}
    declarations={(s['file'],s['start']):s for s in adapted['declarations']}
    proof_count=len(proofs)
    proofs={(s['file'],s['byte_start'],s['byte_end']):s for s in proofs}
    assert len(proofs)==proof_count,'duplicate cast-proof span'
    caches={}; output=[]; covered=set()
    semantic={}
    for transition in transitions:
        if transition.get('semantic_before'):
            before=json.loads(Path(transition['semantic_before']).read_text())
            after=json.loads(Path(transition['semantic_after']).read_text())
            semantic[transition['adaptation']]=({(d['file'],d['start']):d for d in before['declarations']},{(d['file'],d['start']):d for d in after['declarations']})
    for original in stock['sites']:
        row=dict(original);start,end=row['start'],row['end'];removed=None;history=[]
        for transition in transitions:
            file=row['file'];key=(transition['adaptation'],file)
            if key not in caches:
                before=Path(transition['before'])/file.removeprefix('src/')
                after=Path(transition['after'])/file.removeprefix('src/')
                if not before.exists() or not after.exists():raise AssertionError(('missing transition source',key))
                a,b=units(before.read_bytes().decode()),units(after.read_bytes().decode())
                caches[key]=equal_ranges(a,b) if a!=b else [(0,len(a),0)]
            position=mapped(start,end,caches[key])
            if position is None:
                removed=transition['adaptation'];break
            history.append((transition['adaptation'],start,position[0]))
            start,end=position
        current=adapted_index.get((row['file'],row['kind'],start)) if removed is None else None
        if removed is not None:
            row.update(disposition='open' if ignore_rewrites else 'rewritten',adaptation=removed,reason='original AST anchor replaced or removed by this adaptation')
        elif row['kind']=='any_declaration' and current is None:
            declaration=declarations.get((row['file'],start))
            if declaration and not declaration['any']:
                evidence=None
                for adaptation,before_pos,after_pos in history:
                    if adaptation not in semantic:continue
                    old,new=semantic[adaptation]
                    old,new=old.get((row['file'],before_pos)),new.get((row['file'],after_pos))
                    if old and new and old['any'] and not new['any']:
                        evidence=(adaptation,old,new);break
                if evidence:
                    row.update(disposition='open' if ignore_rewrites else 'rewritten',adaptation=evidence[0],reason='checker type changes from any to '+evidence[2]['type'],rewrite_type_evidence={'before':evidence[1],'after':evidence[2]},adapted=declaration)
                elif len(transitions)==1:
                    row.update(disposition='open' if ignore_rewrites else 'rewritten',adaptation=transitions[0]['adaptation'],reason='single-transition fixture declaration changes from any to '+declaration['type'],adapted=declaration)
                else:raise AssertionError(('unattributed semantic rewrite',row['file'],row['line']))
            else:row.update(disposition='open',reason='no unique final AST correspondence',correspondence_unresolved=True)
        elif current is None:
            row.update(disposition='open',reason='no unique final AST correspondence',correspondence_unresolved=True)
        else:
            covered.add((current['file'],current['kind'],current['start']))
            row.update(disposition='open',adapted=current)
            if row['kind']=='as_cast':
                proof=proofs.get((current['file'],current['byte_node_start'],current['byte_node_end']))
                if proof is None:raise AssertionError(('missing Adamic cast proof',current))
                row['adamic']=proof
                row['adamic_refuses_unchecked']=None if proof['state'] in ('unresolved','other_error') else proof['state']=='refused' and "a cast the runtime can't check" in proof.get('message','')
                if proof['state']=='upcast' and (current['assignable'] or current['const_assertion']) and not current['source_any'] and not current['target_any']:
                    row.update(disposition='proven_safe_upcast',proof_kind='const_assertion' if current['const_assertion'] else 'checker_assignable_and_adamic_castProof')
        row['id']=f"{row['file']}:{row['line']}:{row['column']}:{row['kind']}"
        output.append(row)
    additions=[]
    for current in adapted['sites']:
        key=(current['file'],current['kind'],current['start'])
        if key in covered:continue
        # A rewritten declaration can retain an any member elsewhere; never silently omit it.
        extra=dict(current,disposition='open',origin='adapted_only_or_unmatched')
        if current['kind']=='as_cast':
            proof=proofs.get((current['file'],current['byte_node_start'],current['byte_node_end']))
            if proof is None:raise AssertionError(('missing added cast proof',current))
            extra['adamic']=proof
            extra['adamic_refuses_unchecked']=None if proof['state'] in ('unresolved','other_error') else proof['state']=='refused' and "a cast the runtime can't check" in proof.get('message','')
            if proof['state']=='upcast' and (current['assignable'] or current['const_assertion']) and not current['source_any'] and not current['target_any']:
                extra.update(disposition='proven_safe_upcast',proof_kind='const_assertion' if current['const_assertion'] else 'checker_assignable_and_adamic_castProof')
        additions.append(extra)
    return output,additions


def main():
    p=argparse.ArgumentParser();p.add_argument('stock');p.add_argument('adapted');p.add_argument('transitions');p.add_argument('proofs');p.add_argument('output');p.add_argument('--mutant-ignore-rewrites',action='store_true');args=p.parse_args()
    read=lambda path:json.loads(Path(path).read_text())
    stock,adapted=read(args.stock),read(args.adapted)
    proofs=read(args.proofs)
    for proof in proofs:
        matches=[f['file'] for f in adapted['files'] if proof['file']==f['file'] or proof['file'].endswith('/'+f['file'])]
        if len(matches)!=1:raise AssertionError(('ambiguous proof path',proof['file'],matches))
        prefix=proof['file'][:-len(matches[0])] if proof['file']!=matches[0] else ''
        if prefix and proof.get('message'):proof['message']=proof['message'].replace(prefix,'')
        proof['file']=matches[0]
    rows,additions=reconcile(stock,adapted,read(args.transitions),proofs,ignore_rewrites=args.mutant_ignore_rewrites)
    files={f['file'] for f in stock['files']}|{f['file'] for f in adapted['files']}
    per_file={file:dict(Counter(f"{r['kind']}:{r['disposition']}" for r in rows if r['file']==file)) for file in sorted(files)}
    summary=dict(stock_files=len(stock['files']),adapted_files=len(adapted['files']),stock_kinds=dict(Counter(r['kind'] for r in rows)),dispositions=dict(Counter(r['disposition'] for r in rows)),by_kind={kind:dict(Counter(r['disposition'] for r in rows if r['kind']==kind)) for kind in sorted({r['kind'] for r in rows})},by_adaptation=dict(Counter(r['adaptation'] for r in rows if r['disposition']=='rewritten')),by_file=per_file,adapted_additions=dict(Counter(r['disposition'] for r in additions)))
    open_casts=[r for r in rows+additions if r['disposition']=='open' and r['kind']=='as_cast']
    summary['open_casts']=dict(total=len(open_casts),unchecked_refusal=sum(r.get('adamic_refuses_unchecked') is True for r in open_casts),other_observed=sum(r.get('adamic_refuses_unchecked') is False for r in open_casts),unresolved=sum(r.get('adamic_refuses_unchecked') is None for r in open_casts),states=dict(Counter(r['adamic']['state'] for r in open_casts)))
    summary['adapted_additions_by_file']={file:dict(Counter(f"{r['kind']}:{r['disposition']}" for r in additions if r['file']==file)) for file in sorted(files)}
    summary['outside_compiler']={file:summary['by_file'][file] for file in sorted(files) if not file.startswith('src/compiler/')}
    summary['source_files_added']=[file for file in sorted(files) if file not in {f['file'] for f in stock['files']}]
    out=Path(args.output);out.mkdir(exist_ok=True,parents=True)
    for name,value in [('ledger.json',rows),('adapted-additions.json',additions),('open.json',[r for r in rows+additions if r['disposition']=='open']),('SUMMARY.json',summary)]:
        (out/name).write_text(json.dumps(value,indent=2)+'\n')
    print(json.dumps({k:v for k,v in summary.items() if k not in ('by_file','adapted_additions_by_file')},indent=2))


if __name__=='__main__':main()
