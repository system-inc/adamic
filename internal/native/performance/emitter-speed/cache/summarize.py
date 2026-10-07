import collections,json,re,sys
from pathlib import Path

def summarize(path):
    symbols={}; functions=collections.defaultdict(list); function=None; pending=False; events=[]; summary=[]; footer=[]
    for text in path.read_text().splitlines():
        if text.startswith('events:'): events=text.split()[1:]
        elif text.startswith('summary:'): summary=list(map(int,text.split()[1:]))
        elif text.startswith('totals:'): footer=list(map(int,text.split()[1:]))
        elif text.startswith(('fn=','cfn=')):
            key,value=text.split('=',1); match=re.match(r'\((\d+)\)(?: (.*))?$',value)
            if match:
                identifier,name=match.groups()
                if name is not None: symbols[identifier]=name
                value=symbols[identifier]
            if key=='fn': function=value
        elif text.startswith('calls='): pending=True
        elif function and text and text[0] in '0123456789+-*':
            if pending: pending=False;continue
            values=list(map(int,text.split()[1:]));values.extend([0]*(len(events)-len(values)))
            if len(values)!=len(events): raise ValueError('unexpected event vector')
            if not functions[function]: functions[function]=[0]*len(events)
            functions[function]=[a+b for a,b in zip(functions[function],values)]
    summed=[sum(v[i] for v in functions.values()) for i in range(len(events))]
    if summed!=footer: raise ValueError('self events differ from totals footer')
    difference=[a-b for a,b in zip(summary,footer)]
    if any(difference[i] for i,event in enumerate(events) if event!='Ir'): raise ValueError('non-Ir summary discrepancy')
    # Prove the event reconciliation checks every event, not just instructions.
    caught=[]
    for i,event in enumerate(events):
        mutant=footer.copy();mutant[i]+=1
        if summed==mutant: raise ValueError('accounting mutant escaped')
        caught.append(event)
    def row(name,values):
        x=dict(zip(events,values));return dict(function=name,**x,L1_data=x.get('D1mr',0)+x.get('D1mw',0),L1_instruction=x.get('I1mr',0),branch_mispredicts=x.get('Bcm',0)+x.get('Bim',0))
    rows=[row(name,v) for name,v in functions.items()]
    result=dict(events=events,summary=dict(zip(events,summary)),self_totals=row('whole',footer),summary_minus_self=dict(zip(events,difference)),accounting_mutants_caught=caught,by_Ir=sorted(rows,key=lambda x:-x['Ir']),by_L1_data=sorted(rows,key=lambda x:-x['L1_data']),by_L1_instruction=sorted(rows,key=lambda x:-x['L1_instruction']),by_branch=sorted(rows,key=lambda x:-x['branch_mispredicts']))
    path.with_suffix('.json').write_text(json.dumps(result,indent=2)+'\n')
    return result
if __name__=='__main__':
    for name in sys.argv[1:]:
        result=summarize(Path(name));print(name,json.dumps(result['self_totals']))
