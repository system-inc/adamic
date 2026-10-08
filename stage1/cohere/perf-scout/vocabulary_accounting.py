"""Exclusive multi-event Callgrind accounting; never sum recursive inclusive costs."""
from collections import Counter
import json
from pathlib import Path
import re
from vocabulary_wall import category


def summarize(path):
    events=[];symbols={};function=None;pending=False;own={};calls=Counter();callee=None;summary=None;descriptions=[];files={};source_file='???';source_line=0;lines=Counter();sites=Counter();edge_count=0
    for line in Path(path).read_text().splitlines():
        if line.startswith('events:'):events=line.split()[1:]
        elif line.startswith('desc:'):descriptions.append(line[5:].strip())
        elif line.startswith('summary:'):summary=list(map(int,line.split()[1:]))
        elif line.startswith(('fl=','fi=','fe=','cfl=','cfi=')):
            name=line.split('=',1)[1];m=re.fullmatch(r'\((\d+)\)(?: (.*))?',name)
            if m:
                identifier,value=m.groups()
                if value is not None:files[identifier]=value
                name=files[identifier]
            if not line.startswith('c'):source_file=name
        elif line.startswith(('fn=','cfn=')):
            name=line.split('=',1)[1];m=re.fullmatch(r'\((\d+)\)(?: (.*))?',name)
            if m:
                identifier,value=m.groups()
                if value is not None:symbols[identifier]=value
                name=symbols[identifier]
            if line.startswith('fn='):function=name
            else:callee=name
        elif line.startswith('calls='):
            pending=True;edge_count=int(line[6:].split()[0])
            if function and callee:calls[(function,callee)]+=int(line[6:].split()[0])
        elif function and line and line[0] in '0123456789+-*':
            position=line.split()[0]
            if position!='*':source_line=source_line+int(position) if position[0] in '+-' else int(position)
            values=list(map(int,line.split()[1:]))
            values += [0]*(len(events)-len(values))
            if len(values)!=len(events):raise ValueError('event columns changed')
            if pending:
                sites[source_file,source_line,function,callee]+=edge_count
                pending=False;continue
            counters=own.setdefault(function,Counter())
            counters.update(dict(zip(events,values)))
            if 'Ir' in events:lines[source_file,source_line,function]+=values[events.index('Ir')]
    if not events or summary is None:raise ValueError('unfinished profile')
    summary += [0]*(len(events)-len(summary));total=dict(zip(events,summary))
    sums=Counter()
    for counters in own.values():sums.update(counters)
    if any(sums[e]!=total[e] for e in events):raise ValueError('self events do not sum to summary')
    categories={}
    for name,counters in own.items():categories.setdefault(category(name),Counter()).update(counters)
    return {'events':events,'total':total,'descriptions':descriptions,'categories':{k:dict(v) for k,v in categories.items()},'self':{k:dict(v) for k,v in own.items()},'line_self':[{'file':f,'line':n,'function':fn,'Ir':v} for (f,n,fn),v in lines.most_common()],'calls':[{'caller':a,'callee':b,'count':n} for (a,b),n in calls.most_common()],'sites':[{'file':f,'line':line,'caller':a,'callee':b,'count':n} for (f,line,a,b),n in sites.most_common()]}

def summarize_parts(paths):
    if not paths:raise ValueError('no checkpoint profiles')
    reports=[summarize(path) for path in paths]
    events=reports[0]['events'];total=Counter();own={};categories={};calls=Counter();lines=Counter();sites=Counter()
    for report in reports:
        if report['events']!=events:raise ValueError('checkpoint event set differs')
        total.update(report['total'])
        for name,value in report['self'].items():own.setdefault(name,Counter()).update(value)
        for name,value in report['categories'].items():categories.setdefault(name,Counter()).update(value)
        for edge in report['calls']:calls[edge['caller'],edge['callee']]+=edge['count']
        for site in report['sites']:sites[site['file'],site['line'],site['caller'],site['callee']]+=site['count']
        for line in report['line_self']:lines[line['file'],line['line'],line['function']]+=line['Ir']
    return {'events':events,'total':dict(total),'descriptions':reports[0]['descriptions'],'parts':[str(path) for path in paths],'categories':{k:dict(v) for k,v in categories.items()},'self':{k:dict(v) for k,v in own.items()},'line_self':[{'file':f,'line':n,'function':fn,'Ir':v} for (f,n,fn),v in lines.most_common()],'calls':[{'caller':a,'callee':b,'count':n} for (a,b),n in calls.most_common()],'sites':[{'file':f,'line':line,'caller':a,'callee':b,'count':n} for (f,line,a,b),n in sites.most_common()]}

if __name__=='__main__':
    import argparse
    p=argparse.ArgumentParser();p.add_argument('profile',type=Path);p.add_argument('output',type=Path);a=p.parse_args();a.output.write_text(json.dumps(summarize(a.profile),indent=2)+'\n')
