#!/usr/bin/env python3
"""Separate inclusive vocabulary-walk scope from its zero-ownership membership call."""
import argparse
import json
from pathlib import Path
import re
import boundary
import scout


def tracked(source):
    function=next(f for f in boundary.functions(source) if f['plain']=='abbreviation')
    callee=function['name']
    inventory=[];edits=[]
    for f in boundary.functions(source):
        body=source[f['body']:f['end']]
        for match in re.finditer(re.escape(callee)+r'\(',body):
            end=boundary.closing(body,match.end()-1)+1
            args=body[match.end():end-1]
            site=dict(id=len(inventory),key=f['key']+':vocabulary',category='vocabulary',caller=f['name'],callee=callee,
                      c_line=source.count('\n',0,f['body']+match.start())+1,eligible=False)
            inventory.append(site)
            edits.append((f['body']+match.start(),f['body']+end,f'scout_vocabulary({site["id"]}, {callee}, {args})'))
    members=[s for s in boundary.sites(source) if s['category']=='membership']
    for site in members:
        site={**site,'id':len(inventory)};inventory.append(site)
        args=site['expression'][len(site['callee'])+1:-1]
        edits.append((site['start'],site['end'],f'scout_membership({site["id"]}, {args})'))
    for start,end,replacement in sorted(edits,reverse=True):source=source[:start]+replacement+source[end:]
    helper=r'''
static adamic_object *scout_vocabulary(size_t id, adamic_object *(*fn)(adamic_string *), adamic_string *text) {
 scout_site *s=&scout_sites[id]; bool sample=(s->calls++ % 64)==0;
 size_t r=adamic_counted.retains,d=adamic_counted.releases;uint64_t begin=sample?scout_clock():0;
 adamic_object *value=fn(text);
 if(sample){s->nanoseconds+=scout_clock()-begin;s->samples++;}
 s->retains+=adamic_counted.retains-r;s->releases+=adamic_counted.releases-d;return value;
}
'''
    anchor='#include <stdbool.h>\n'
    source='#define _POSIX_C_SOURCE 200809L\n'+source.replace(anchor,anchor+'#define SCOUT_SITE_COUNT '+str(len(inventory))+'\n'+boundary.HEADER+helper,1)
    return source,inventory


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('mode',choices=('prepare','measure'));p.add_argument('directory',type=Path);p.add_argument('--baseline',type=Path,required=True);p.add_argument('--archive',type=Path);a=p.parse_args();d=a.directory.resolve();base=a.baseline.resolve()
    if a.mode=='prepare':
        if not a.archive:p.error('--archive required')
        source,inventory=tracked((base/'lint/main.c').read_text())
        (d/'vocabulary-sites.json').write_text(json.dumps(inventory,indent=2)+'\n')
        boundary.build(d,a.archive.resolve(),'vocabulary-sites',source,counted=True)
    else:
        scout.USAGE_HELPER=base/'usage';report={}
        for name,manifest in (('scanner',d/'single.txt'),('public23',d/'public-all.txt'),('compiler77',base/'fused/manifest.txt')):
            out=d/name/'vocabulary-sites';scout.run([d/'vocabulary-sites','--manifest',manifest],out)
            scout.checked(out.read_bytes(),(d/name/'Go').read_bytes())
            report[name]=boundary.parse_counts(str(out)+'.stderr',json.loads((d/'vocabulary-sites.json').read_text()))
        (d/'vocabulary.json').write_text(json.dumps(report,indent=2)+'\n')
if __name__=='__main__':main()
