#!/usr/bin/env python3
"""Hand-measure the step-41 ruling in scratch C; never modify shared sources."""
import argparse
from collections import Counter
from concurrent.futures import ThreadPoolExecutor
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import statistics
import subprocess
import sys
import time
import scout

HERE = Path(__file__).resolve().parent
NODE = re.compile(r'(?:adamic_function_\d+|scout_lent)_(?:Parser_node|RuleContext_node|Linter_node)')
DEF = re.compile(r'^static [^\n;]*\b(adamic_function_\d+_\w+)\([^\n]*\) \{\n', re.M)


def closing(text, start):
    pairs = {'(': ')', '{': '}', '[': ']'}
    stack = [pairs[text[start]]]
    quote = ''
    escaped = False
    for index in range(start + 1, len(text)):
        ch = text[index]
        if quote:
            if escaped: escaped = False
            elif ch == '\\': escaped = True
            elif ch == quote: quote = ''
            continue
        if ch in ('"', "'"): quote = ch
        elif ch in pairs: stack.append(pairs[ch])
        elif ch == stack[-1]:
            stack.pop()
            if not stack: return index
    raise ValueError('unclosed C expression')


def functions(source):
    groups = Counter()
    result = []
    for match in DEF.finditer(source):
        name = match[1]
        plain = re.sub(r'^adamic_function_\d+_', '', name)
        ordinal = groups[plain]; groups[plain] += 1
        start = match.end() - 2
        end = closing(source, start) + 1
        result.append(dict(name=name, plain=plain, key=f'{plain}#{ordinal}', start=match.start(), body=start,
                           end=end, text=source[match.start():end]))
    return result


def aliases_and_reads(body, variable):
    aliases = {variable}
    changed = True
    while changed:
        changed = False
        for lhs, rhs in re.findall(r'adamic_object \*\s+(\w+) = (\w+);', body):
            if rhs in aliases and lhs not in aliases: aliases.add(lhs); changed = True
    releases = []
    reasons = []
    for line in body.splitlines():
        for var in aliases:
            if not re.search(r'\b' + var + r'\b', line): continue
            if re.search(r'adamic_object \*\s+' + var + r' = ', line): continue
            if re.search(r'adamic_object \*\s+\w+ = ' + var + ';', line): continue
            if re.fullmatch(r'\s*adamic_release\(' + var + r'\);', line):
                releases.append(var); continue
            # Only ordinary value loads. Slot pointers and calls receiving the object are excluded.
            if 'adamic_value *' not in line and (var + '->' in line or 'adamic_object_data_field(' + var + ',' in line):
                scrub = re.sub(r'\b' + var + r'->', 'FIELD->', line)
                scrub = scrub.replace('adamic_object_data_field(' + var + ',', 'FIELD(')
                if not re.search(r'\b' + var + r'\b', scrub): continue
            reasons.append(line.strip())
    return sorted(aliases), releases, sorted(set(reasons))


def sites(source):
    result = []
    for function in functions(source):
        # Forwarders return an owned reference; never change their public contract.
        if function['plain'] in ('Parser_node', 'RuleContext_node', 'Linter_node'): continue
        body = source[function['body']:function['end']]
        occurrences = Counter()
        targets = re.compile(NODE.pattern + r'\(|adamic_function_\d+_written\(|adamic_array_search_from\(')
        for match in targets.finditer(body):
            callee = match.group()[:-1]
            category = 'node' if NODE.fullmatch(callee) else 'encoder' if callee.endswith('_written') else 'membership'
            if category == 'encoder' and function['plain'] != 'run': continue
            if category == 'membership' and function['plain'] != 'abbreviation': continue
            end = closing(body, match.end() - 1) + 1
            before = body[:match.start()].split('\n')[-1]
            varmatch = re.search(r'adamic_object \*\s+(\w+) = $', before)
            aliases, releases, reasons = aliases_and_reads(body, varmatch[1]) if varmatch else ([], [], ['no unique result binding'])
            parser_phase = function['plain'].startswith(('Parser_', 'Expressions_', 'Statements_', 'JSX_'))
            eligible = category == 'node' and bool(releases) and not reasons and not parser_phase
            ordinal = occurrences[category]; occurrences[category] += 1
            key = function['key'] + ':' + category + ':' + str(ordinal)
            absolute = function['body'] + match.start()
            result.append(dict(id=len(result), key=key, category=category, caller=function['name'], caller_key=function['key'],
                               callee=callee, start=absolute, end=function['body'] + end,
                               c_line=source.count('\n', 0, absolute) + 1,
                               expression=body[match.start():end], aliases=aliases, release_vars=releases,
                               eligible=eligible, excluded_reasons=reasons + (['parser construction phase'] if parser_phase else [])))
    return result


def lent_functions(source):
    selected = [f for f in functions(source) if f['plain'] in ('Parser_node', 'RuleContext_node', 'Linter_node')]
    if len(selected) != 3: raise ValueError('expected three unique node forwarders')
    prototypes = []
    bodies = []
    names = {f['name']: 'scout_lent_' + f['plain'] for f in selected}
    for f in selected:
        clone = f['text']
        for before, after in names.items(): clone = clone.replace(before, after)
        if f['plain'] == 'Parser_node':
            clone, count = re.subn(r'adamic_retain\((adamic_temporary_\d+->reference)\)', r'\1', clone)
            if count != 1: raise ValueError('Parser.node retain anchor changed')
        elif 'adamic_retain(' in clone or 'adamic_release(' in clone:
            raise ValueError('forwarder ownership changed')
        prototypes.append(clone.split(' {\n', 1)[0] + ';')
        bodies.append(clone)
    return names, '\n'.join(prototypes), '\n\n'.join(bodies)


def borrow(source, selected_keys=None, unpaired=False):
    original = sites(source)
    selected = [s for s in original if s['eligible'] and (selected_keys is None or s['key'] in selected_keys)]
    if not selected: raise ValueError('no proven read-only node site')
    names, prototypes, bodies = lent_functions(source)
    edits = [(s['start'], s['start'] + len(s['callee']), names[s['callee']]) for s in selected]
    for start, end, replacement in sorted(edits, reverse=True): source = source[:start] + replacement + source[end:]
    if not unpaired:
        by_caller = {}
        for site in selected: by_caller.setdefault(site['caller'], set()).update(site['release_vars'])
        for f in reversed(functions(source)):
            if f['name'] not in by_caller: continue
            body = f['text']
            for variable in by_caller[f['name']]:
                body, count = re.subn(r'\badamic_release\(' + variable + r'\);', '/* scout: ruled borrowed read, no release */', body)
                if not count: raise ValueError('paired caller release missing')
            source = source[:f['start']] + body + source[f['end']:]
    anchor = '#include <stdbool.h>\n'
    if source.count(anchor) != 1: raise ValueError('include anchor changed')
    source = source.replace(anchor, anchor + prototypes + '\n', 1) + '\n' + bodies + '\n'
    return source, selected


HEADER = r'''
#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>
#include <time.h>
#include "count.h"
typedef struct { uint64_t calls, retains, releases, caller_releases, samples, nanoseconds; } scout_site;
static scout_site scout_sites[SCOUT_SITE_COUNT];
static uint64_t scout_clock(void) { struct timespec now; if(clock_gettime(CLOCK_MONOTONIC, &now)!=0) abort(); return (uint64_t)now.tv_sec*1000000000u+(uint64_t)now.tv_nsec; }
static uint64_t scout_calibration;
static adamic_object *scout_node(size_t id, adamic_object *(*fn)(adamic_object *, double), adamic_object *owner, double index) {
 scout_site *s=&scout_sites[id]; bool sample=(s->calls++ % 1024)==0;
 size_t r=adamic_counted.retains, d=adamic_counted.releases; uint64_t begin=sample?scout_clock():0;
 adamic_object *value=fn(owner,index);
 if(sample){s->nanoseconds+=scout_clock()-begin;s->samples++;}
 s->retains+=adamic_counted.retains-r;s->releases+=adamic_counted.releases-d;return value;
}
static adamic_string *scout_encoder(size_t id, adamic_string *(*fn)(adamic_string *), adamic_string *text) {
 scout_site *s=&scout_sites[id];bool sample=(s->calls++ % 1024)==0;
 size_t r=adamic_counted.retains,d=adamic_counted.releases;uint64_t begin=sample?scout_clock():0;
 adamic_string *value=fn(text);
 if(sample){s->nanoseconds+=scout_clock()-begin;s->samples++;}
 s->retains+=adamic_counted.retains-r;s->releases+=adamic_counted.releases-d;return value;
}
static double scout_membership(size_t id,const adamic_array *a,adamic_value value,enum adamic_equality eq,bool svz,double from,bool has,bool last) {
 scout_site *s=&scout_sites[id];bool sample=(s->calls++ % 1024)==0;
 size_t r=adamic_counted.retains,d=adamic_counted.releases;uint64_t begin=sample?scout_clock():0;
 double found=adamic_array_search_from(a,value,eq,svz,from,has,last);
 if(sample){s->nanoseconds+=scout_clock()-begin;s->samples++;}
 s->retains+=adamic_counted.retains-r;s->releases+=adamic_counted.releases-d;return found;
}
static void scout_release(size_t id,void *value) { scout_sites[id].caller_releases++;adamic_release(value); }
__attribute__((constructor)) static void scout_calibrate(void) {uint64_t total=0;for(size_t i=0;i<10000;i++){uint64_t a=scout_clock();total+=scout_clock()-a;}scout_calibration=total/10000;}
__attribute__((destructor)) static void scout_report(void) {
 fprintf(stderr,"scout_calibration %llu\n",(unsigned long long)scout_calibration);
 for(size_t i=0;i<SCOUT_SITE_COUNT;i++){scout_site *s=&scout_sites[i];if(s->calls)fprintf(stderr,"scout_site %zu %llu %llu %llu %llu %llu %llu\n",i,(unsigned long long)s->calls,(unsigned long long)s->retains,(unsigned long long)s->releases,(unsigned long long)s->caller_releases,(unsigned long long)s->samples,(unsigned long long)s->nanoseconds);}
}
'''


def instrument(source):
    inventory = sites(source)
    edits = []
    for s in inventory:
        args = s['expression'][len(s['callee']) + 1:-1]
        wrapper = {'node': 'scout_node', 'encoder': 'scout_encoder', 'membership': 'scout_membership'}[s['category']]
        extra = s['callee'] + ', ' if s['category'] != 'membership' else ''
        edits.append((s['start'], s['end'], f'{wrapper}({s["id"]}, {extra}{args})'))
    for start, end, replacement in sorted(edits, reverse=True): source = source[:start] + replacement + source[end:]
    # Only proven nonescaping bindings have an unambiguous paired caller release.
    for f in reversed(functions(source)):
        body = f['text']
        for s in inventory:
            if not s['eligible'] or s['caller'] != f['name']: continue
            for variable in set(s['release_vars']):
                body = body.replace('adamic_release(' + variable + ');', f'scout_release({s["id"]}, {variable});')
        source = source[:f['start']] + body + source[f['end']:]
    header = '#define _POSIX_C_SOURCE 200809L\n'
    source = header + source
    anchor = '#include <stdbool.h>\n'
    source = source.replace(anchor, anchor + '#define SCOUT_SITE_COUNT ' + str(len(inventory)) + '\n' + HEADER, 1)
    return source, inventory


def command(argv, log, env=None):
    with Path(log).open('wb') as output:
        result = subprocess.run(list(map(str, argv)), stdout=output, stderr=subprocess.STDOUT, env=env)
    if result.returncode: raise RuntimeError('command failed: ' + str(argv[:3]) + '; see ' + str(log))


def build(directory, archive, name, source, counted=False, sanitized=False):
    c = directory/(name + '.c');c.write_text(source)
    flags = ['-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-DADAMIC_TSGO','-g']
    flags += ['-O1','-fsanitize=address,undefined','-fno-sanitize-recover=all'] if sanitized else ['-O2']
    if counted: flags += ['-DADAMIC_COUNT']
    key = 'sanitized' if sanitized else 'counted' if counted else 'release'
    runtime = directory/'runtime';objects = directory/('objects-' + key);objects.mkdir(exist_ok=True)
    sources = sorted(runtime.glob('*.c'))
    def one(path):
        obj=objects/(path.stem+'.o')
        if not obj.exists():command(['clang',*flags,'-I'+str(runtime),'-c',path,'-o',obj],objects/(path.stem+'.log'))
        return obj
    with ThreadPoolExecutor(max_workers=4) as pool: units=list(pool.map(one,sources))
    command(['clang',*flags,'-I'+str(runtime),'-c',c,'-o',directory/(name+'.o')],directory/(name+'-compile.log'))
    command(['clang',*flags,directory/(name+'.o'),*units,archive,'-lm','-lpthread','-ldl','-o',directory/name],directory/(name+'-link.log'))
    return flags


def prepare(directory, baseline, archive):
    if directory.is_relative_to(scout.REPO): raise ValueError('generated experiments must live outside shared repository')
    directory.mkdir(parents=True,exist_ok=True)
    shutil.copytree(HERE,directory/'package-copy',dirs_exist_ok=True)
    runtime=directory/'runtime';runtime.mkdir(exist_ok=True)
    for path in (baseline/'lint').iterdir():
        if path.suffix in ('.c','.h') and path.name!='main.c':shutil.copyfile(path,runtime/path.name)
    before=(baseline/'lint/main.c').read_text()
    after=(baseline/'encoding/main.c').read_text()
    borrowed,selected=borrow(before)
    (directory/'selected.json').write_text(json.dumps(selected,indent=2)+'\n')
    top=None
    variants=[('before',before,False),('borrow',borrowed,False),('encoder',after,False)]
    for name,source,_ in variants:
        flags=build(directory,archive,name,source)
        tracked,inventory=instrument(source)
        (directory/(name+'-sites.json')).write_text(json.dumps(inventory,indent=2)+'\n')
        build(directory,archive,name+'-sites',tracked,counted=True)
    # Lifetime validation is separate from release timing and counter instrumentation.
    build(directory,archive,'borrow-sanitized',borrowed,sanitized=True)
    first=next(s for s in selected if 'Linter_ancestry' in s['caller'])
    mutant,_=borrow(before,{first['key']},unpaired=True)
    build(directory,archive,'unpaired-mutant',mutant,sanitized=True)
    (directory/'build.json').write_text(json.dumps(dict(flags=flags,base=scout.REPO.as_posix(),original_sha256=hashlib.sha256(before.encode()).hexdigest(),selected_sites=len(selected)),indent=2)+'\n')


def parse_counts(path, inventory):
    data=Path(path).read_text();calibration=int(re.search(r'^scout_calibration (\d+)$',data,re.M)[1]);rows=[]
    for match in re.finditer(r'^scout_site (\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+)$',data,re.M):
        index,calls,retains,releases,paired,samples,nanos=map(int,match.groups());row={**inventory[index], 'calls':calls,'retains':retains,'releases':releases,'paired_releases':paired,'samples':samples,'sampled_ns':nanos}
        row['calibrated_ns_per_call']=max(0,nanos/samples-calibration)
        row['estimated_total_ns']=row['calibrated_ns_per_call']*calls
        rows.append(row)
    return dict(calibration_ns=calibration,sites=sorted(rows,key=lambda r:r['estimated_total_ns'],reverse=True),global_counts=re.search(r'^adamic: counts:.*$',data,re.M)[0])


def measure(directory, baseline, rounds):
    scout.USAGE_HELPER=baseline/'usage'
    report={'machine_before':scout.machine(),'workloads':{}}
    # Same real sources, full registered rules; one representative file retains the 77-root program.
    manifest=directory/'single.txt'
    old=(baseline/'fused/manifest.txt').read_text().splitlines()
    row=next(x for x in old if x.split('\t')[0].endswith('/scanner.ts'))
    manifest.write_text(old[0]+'\n'+row+'\n')
    public=directory/'public-all.txt'
    public_files=[r.split('\t')[0] for r in scout.rows(baseline/'public.txt')]
    public_config=directory/'public-tsconfig.json';public_config.write_text(json.dumps({'compilerOptions':{'strict':True},'files':public_files}))
    public_rows=[]
    for raw in scout.rows(baseline/'public.txt'):
        fields=raw.split('\t');fields[1]='all';public_rows.append('\t'.join(fields))
    public.write_text('program '+str(public_config)+'\n'+'\n'.join(public_rows)+'\n')
    for workload,inputs in (('scanner',manifest),('public23',public),('compiler77',baseline/'fused/manifest.txt')):
        dest=directory/workload;dest.mkdir(exist_ok=True)
        scout.run([baseline/'lint/oracle','--manifest',inputs],dest/'Go')
        want=(dest/'Go').read_bytes();scout.no_skips(want)
        scout.run([directory/'before','--manifest',inputs,'--record',dest/'transcript'],dest/'record');scout.checked((dest/'record').read_bytes(),want)
        scout.run(['node','--disable-warning=ExperimentalWarning',scout.REPO/'oracle/node.mjs',baseline/'lint/main.ts','--manifest',inputs,'--replay',dest/'transcript'],dest/'Node');scout.checked((dest/'Node').read_bytes(),want)
        result={'bytes':len(want),'sha256':hashlib.sha256(want).hexdigest(),'samples':{s:[] for s in ('before','borrow','encoder')},'sites':{}}
        for r in range(rounds):
            names=list(result['samples']);names=names[r%3:]+names[:r%3]
            for side in names:
                output=dest/f'{side}-{r}';result['samples'][side].append(scout.run([directory/side,'--manifest',inputs],output));scout.checked(output.read_bytes(),want)
        for side in ('before','borrow','encoder'):
            output=dest/(side+'-sites');scout.run([directory/(side+'-sites'),'--manifest',inputs],output);scout.checked(output.read_bytes(),want)
            result['sites'][side]=parse_counts(str(output)+'.stderr',json.loads((directory/(side+'-sites.json')).read_text()))
        report['workloads'][workload]=result
        (directory/'measurements.json').write_text(json.dumps(report,indent=2)+'\n')
    # Broad fixture parity without using fixture runs as timing samples.
    for name in ('owned','upstream','repository'):
        dest=directory/name;dest.mkdir(exist_ok=True);want=(baseline/'measure'/name/'Go.answer').read_bytes()
        output=dest/'borrow';scout.run([directory/'borrow','--manifest',baseline/(name+'.txt')],output);scout.checked(output.read_bytes(),want)
    tiny=directory/'tiny.txt';tiny.write_text(scout.rows(baseline/'owned.txt')[0]+'\n')
    output=directory/'sanitized-answer';scout.run([directory/'borrow-sanitized','--manifest',manifest],output);scout.checked(output.read_bytes(),(directory/'scanner/Go').read_bytes())
    with (directory/'mutant.stdout').open('wb') as out,(directory/'mutant.stderr').open('wb') as err:
        mutant=subprocess.run([directory/'unpaired-mutant','--manifest',tiny],stdout=out,stderr=err)
    if mutant.returncode==0 or b'AddressSanitizer' not in (directory/'mutant.stderr').read_bytes():raise ValueError('unpaired lifetime mutant did not fail under ASan')
    report['mutant_exit']=mutant.returncode;report['machine_after']=scout.machine()
    (directory/'measurements.json').write_text(json.dumps(report,indent=2)+'\n')


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('mode',choices=('prepare','measure'));p.add_argument('directory',type=Path);p.add_argument('--baseline',type=Path,required=True);p.add_argument('--archive',type=Path);p.add_argument('--rounds',type=int,default=3);a=p.parse_args()
    if a.mode=='prepare':
        if not a.archive:p.error('--archive required')
        prepare(a.directory.resolve(),a.baseline.resolve(),a.archive.resolve())
    else:
        if a.rounds<3:p.error('at least three rounds required')
        measure(a.directory.resolve(),a.baseline.resolve(),a.rounds)
if __name__=='__main__':main()
