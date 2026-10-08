#!/usr/bin/env python3
"""Full-manifest vocabulary profiling; all generated changes stay in scratch."""
import argparse
from bisect import bisect_right
from collections import Counter
import ctypes
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import boundary
import scout

HERE=Path(__file__).resolve().parent

def native_source(source, entries=False):
    function=next(f for f in boundary.functions(source) if f['plain']=='abbreviation')
    old=function['name']; inner=old+'_body'
    body=function['text'].replace(old+'(',inner+'(',1)
    if entries:
        one='adamic_array * adamic_local_3063_entry = '
        two='adamic_array * adamic_local_3070_entry = '
        if body.count(one)!=1 or body.count(two)!=1: raise ValueError('changed vocabulary loops')
        body=body.replace(one,'++scout_entries[0];\n\t\t\t\t'+one).replace(two,'++scout_entries[1+adamic_temporary_3250];\n\t\t\t\t'+two)
    wrapper=f'''\nstatic adamic_object * {old}(adamic_string *text) {{
 bool timed=(scout_calls++ % 64)==0;uint64_t begin=timed?scout_nanoseconds():0;
 scout_active=1;adamic_object *value={inner}(text);scout_active=0;
 if(timed){{scout_elapsed+=scout_nanoseconds()-begin;scout_timed++;}}return value;
}}\n'''
    source=source[:function['start']]+body+wrapper+source[function['end']:]
    helper=(HERE/'testdata/vocabulary_sample.c').read_text()
    source='#define _GNU_SOURCE\n'+source.replace('#include <stdbool.h>\n','#include <stdbool.h>\n'+helper,1)
    marker='int main(int argc, char **argv) {'
    if source.count(marker)!=1:raise ValueError('changed main')
    return source.replace(marker,marker+'\n scout_start();',1)

def go_source(source, entries=False):
    marker='func (vocabulary *abbreviationVocabulary) find(name string) (abbreviationFinding, bool) {'
    if source.count(marker)!=1:raise ValueError('changed Go find')
    source=source.replace('import (','import (\n "io"\n "time"',1).replace(marker,marker.replace(' find(', ' scoutOriginalFind('),1)
    if entries:
        for i,phase in enumerate(('earlyPrefixes','suffixes','latePrefixes','segments')):
            marker='for _, entry := range vocabulary.'+phase+' {'
            if source.count(marker)!=1:raise ValueError('changed Go phase')
            source=source.replace(marker,marker+f'\n scoutEntries[{i}]++',1)
    return source+'''
var scoutCalls, scoutElapsed, scoutTimed uint64
var scoutEntries [4]uint64
func (vocabulary *abbreviationVocabulary) find(name string) (abbreviationFinding, bool) {
 timed:=scoutCalls%64==0;scoutCalls++
 if !timed { return vocabulary.scoutOriginalFind(name) }
 start:=time.Now();finding,found:=vocabulary.scoutOriginalFind(name)
 scoutElapsed+=uint64(time.Since(start));scoutTimed++
 return finding,found
}
func ScoutVocabularyReport(out io.Writer) {
 fmt.Fprintf(out,"SCOUT_WALL {\\"calls\\":%d,\\"sampled_ns\\":%d,\\"timed_calls\\":%d,\\"entries\\":[%d,%d,%d,%d]}\\n",scoutCalls,scoutElapsed,scoutTimed,scoutEntries[0],scoutEntries[1],scoutEntries[2],scoutEntries[3])
}
'''

def prepare(directory,baseline,original,archive):
    if directory.is_relative_to(scout.REPO):raise ValueError('scratch only')
    directory.mkdir(parents=True,exist_ok=True)
    shutil.copytree(original/'runtime',directory/'runtime',dirs_exist_ok=True)
    source=(original/'before.c').read_text()
    for entries in (False,True):
        name='native-entries' if entries else 'native-wall'
        flags=boundary.build(directory,archive,name,native_source(source,entries))
    replacements=json.loads((baseline/'oracle-overlay.json').read_text())['Replace']
    main=(baseline/'lint-oracle.go').read_text().replace('import (','import (\n scoutnexus "github.com/system-inc/cohere/internal/lint/rules/nexus"',1).replace('func main() {','func main() {\n defer scoutnexus.ScoutVocabularyReport(os.Stderr)',1)
    (directory/'oracle.go').write_text(main)
    for entries in (False,True):
        name='Go-entries' if entries else 'Go-wall'
        path=directory/(name+'.go');path.write_text(go_source((scout.REPO/'cohere/internal/lint/rules/nexus/abbreviation_vocabulary.go').read_text(),entries))
        overlay={**replacements,str(scout.REPO/'cohere/adamic_scout_lint.go'):str(directory/'oracle.go'),str(scout.REPO/'cohere/internal/lint/rules/nexus/abbreviation_vocabulary.go'):str(path)}
        op=directory/(name+'-overlay.json');op.write_text(json.dumps({'Replace':overlay}))
        with (directory/(name+'-build.txt')).open('wb') as log:
            subprocess.run(list(map(str,['go','build','-overlay='+str(op),'-o',directory/name,*replacements])),stdout=log,stderr=subprocess.STDOUT,cwd=scout.REPO/'cohere',check=True)
    (directory/'build.json').write_text(json.dumps({'flags':flags,'native_sample_interval_ns':1000000,'scope_clock_every':64,'manifest':str(baseline/'fused/manifest.txt')},indent=2)+'\n')

def category(name):
    if name=='adamic_retain':return 'retain'
    if name.startswith('adamic_release') or name=='release_last':return 'release'
    if 'internal/runtime/maps.' in name or any(s in name.lower() for s in ('hash','mapaccess','mapassign','aeshash','memhash')):return 'hash/map'
    if any(s in name for s in ('malloc','calloc','realloc','free','allocate','object_new','array_new','newobject','mallocgc','mcache','mspan')):return 'allocation/free'
    if name=='memeqbody':return 'string comparison'
    if any(s in name for s in ('runtime.gc','runtime.scan','runtime.grey','runtime.findObject','runtime.typePointers','runtime.tryDeferToSpanScan','runtime.wb','gcWriteBarrier')):return 'Go GC/barrier'
    if any(s in name for s in ('string_equal','memequal','memcmp','strcmp')):return 'string comparison'
    if name.startswith('unresolved:'):return name
    if name in ('units_next','indexbytebody','indexbody'):return 'other strings/regex'
    if any(s in name for s in ('string','strings.','bytealg','regexp','regex','utf8','memcpy','memmove','copy')):return 'other strings/regex'
    return 'other'

def libc_ifuncs():
    # Dynamic IFUNC entries are outside exported ELF symbol sizes. Resolve the same
    # loaded ELF, then require an exact .eh_frame FDE range; never guess by proximity.
    maps=[line.split() for line in Path('/proc/self/maps').read_text().splitlines() if '/libc.so.6' in line]
    if not maps:return None,[]
    path=Path(maps[0][5]).resolve();bias=int(maps[0][0].split('-')[0],16)-int(maps[0][2],16)
    frames=subprocess.check_output(['readelf','--debug-dump=frames',str(path)],text=True)
    bounds=[(int(a,16),int(b,16)) for a,b in re.findall(r'FDE .* pc=([0-9a-f]+)\.\.([0-9a-f]+)',frames)]
    libc=ctypes.CDLL(str(path));result=[]
    for name in ('memcmp','memcpy','memmove','memchr','strlen','strcmp','strchr','memset'):
        offset=ctypes.cast(getattr(libc,name),ctypes.c_void_p).value-bias
        exact=[(a,b) for a,b in bounds if a==offset]
        if len(exact)==1:result.append({'function':'libc IFUNC '+name,'start':exact[0][0],'end':exact[0][1]})
    return path,result

def pc_summary(binary,prefix):
    mappings=[];tables={};total=Counter();symbols=Counter();libc_path,ifuncs=libc_ifuncs()
    for line in Path(str(prefix)+'.maps').read_text().splitlines():
        fields=line.split()
        if len(fields)<6 or 'x' not in fields[1] or not fields[5].startswith('/'):continue
        start,end=[int(x,16) for x in fields[0].split('-')];offset=int(fields[2],16);path=fields[5]
        mappings.append((start,end,start-offset,path))
        argv=['nm','-n','-S','--defined-only',path] if Path(path).resolve()==binary.resolve() else ['nm','-D','-n','-S','--defined-only',path]
        result=subprocess.run(argv,capture_output=True,text=True)
        rows=[]
        for row in result.stdout.splitlines():
            bits=row.split()
            if len(bits)>=4 and bits[2] in ('t','T','i','I','w','W'):
                try:rows.append((int(bits[0],16),bits[3],int(bits[1],16)))
                except ValueError:pass
        rows.sort();tables[path]=([r[0] for r in rows],rows)
    for line in Path(str(prefix)+'.pcs').read_text().splitlines():
        address,n=line.split();address=int(address,16);n=int(n);name='unresolved'
        for start,end,bias,path in mappings:
            if start<=address<end:
                keys,rows=tables[path];i=bisect_right(keys,address-bias)-1
                name='unresolved:'+Path(path).name
                if i>=0 and address-bias<rows[i][0]+rows[i][2]:name=rows[i][1]
                if name.startswith('unresolved:') and Path(path).resolve()==libc_path:
                    matches=[f for f in ifuncs if f['start']<=address-bias<f['end']]
                    if matches:name=matches[0]['function']
                break
        total[category(name)]+=n;symbols[name]+=n
    return {'samples':sum(total.values()),'categories':dict(total),'symbols':dict(symbols),'libc_ifunc_ranges':ifuncs,'libc_path':str(libc_path)}

def scope(path):
    rows=[line[11:] for line in path.read_text().splitlines() if line.startswith('SCOUT_WALL ')]
    if len(rows)!=1:raise ValueError('missing scope result')
    value=json.loads(rows[0]);value['estimated_ns']=value['sampled_ns']/value['timed_calls']*value['calls'];return value

def measure(directory,baseline,original,rounds):
    scout.USAGE_HELPER=baseline/'usage';manifest=baseline/'fused/manifest.txt';want=(original/'compiler77/Go').read_bytes()
    report={'machine_before':scout.machine(),'rounds':rounds,'release':{},'scope':{}}
    for name,binary in (('native',original/'before'),('Go',baseline/'lint/oracle')):report['release'][name]=[]
    for i in range(rounds):
        for name in (('native','Go') if i%2==0 else ('Go','native')):
            binary=original/'before' if name=='native' else baseline/'lint/oracle';out=directory/f'{name}-{i}.answer'
            sample=scout.run([binary,'--manifest',manifest],out);scout.checked(out.read_bytes(),want);report['release'][name].append(sample)
    for name in ('native-wall','native-entries','Go-wall','Go-entries'):
        env=dict(os.environ)
        if name=='native-wall':env['SCOUT_PC']=str(directory/'native-wall')
        if name=='Go-wall':env['SCOUT_PPROF']=str(directory/'Go.pprof')
        out=directory/(name+'.answer');sample=scout.run([directory/name,'--manifest',manifest],out,env=env)
        scout.checked(out.read_bytes(),want);report['scope'][name]=scope(Path(str(out)+'.stderr'));report['scope'][name]['process']=sample
    report['native_pc']=pc_summary(directory/'native-wall',directory/'native-wall')
    report['machine_after']=scout.machine();report['bytes']=len(want)
    (directory/'wall.json').write_text(json.dumps(report,indent=2)+'\n')

def gate(directory,baseline,original):
    marker='func isAbbreviationCandidate(name string) bool {'
    path=scout.REPO/'cohere/internal/lint/rules/nexus/abbreviation_gate.go'
    source=path.read_text()
    if source.count(marker)!=1:raise ValueError('changed candidate gate')
    source=source.replace('package nexus','package nexus\nimport ("fmt"; "io"; "time")',1).replace(marker,marker.replace('isAbbreviationCandidate','scoutOriginalCandidate'),1)
    source += r'''
var scoutGateCalls, scoutGateMatched, scoutGateTimed, scoutGateElapsed uint64
func isAbbreviationCandidate(name string) bool {
 timed:=scoutGateCalls%64==0;scoutGateCalls++
 var start time.Time;if timed { start=time.Now() }
 matched:=scoutOriginalCandidate(name)
 if timed { scoutGateElapsed+=uint64(time.Since(start));scoutGateTimed++ }
 if matched { scoutGateMatched++ };return matched
}
func ScoutGateReport(out io.Writer) {
 fmt.Fprintf(out,"SCOUT_GATE {\"calls\":%d,\"matched\":%d,\"timed_calls\":%d,\"sampled_ns\":%d}\n",scoutGateCalls,scoutGateMatched,scoutGateTimed,scoutGateElapsed)
}
'''
    replacement=directory/'Go-gate.go';replacement.write_text(source)
    main=directory/'Go-gate-main.go';main.write_text((directory/'oracle.go').read_text().replace('func main() {','func main() {\n defer scoutnexus.ScoutGateReport(os.Stderr)',1))
    overlay=json.loads((directory/'Go-entries-overlay.json').read_text())['Replace'];overlay[str(path)]=str(replacement);overlay[str(scout.REPO/'cohere/adamic_scout_lint.go')]=str(main)
    op=directory/'Go-gate-overlay.json';op.write_text(json.dumps({'Replace':overlay}))
    targets=json.loads((baseline/'oracle-overlay.json').read_text())['Replace']
    with (directory/'Go-gate-build.txt').open('wb') as log:
        subprocess.run(list(map(str,['go','build','-overlay='+str(op),'-o',directory/'Go-gate',*targets])),stdout=log,stderr=subprocess.STDOUT,cwd=scout.REPO/'cohere',check=True)
    scout.USAGE_HELPER=baseline/'usage';out=directory/'Go-gate.answer';sample=scout.run([directory/'Go-gate','--manifest',baseline/'fused/manifest.txt'],out)
    scout.checked(out.read_bytes(),(original/'compiler77/Go').read_bytes())
    rows=[line[11:] for line in Path(str(out)+'.stderr').read_text().splitlines() if line.startswith('SCOUT_GATE ')]
    if len(rows)!=1:raise ValueError('missing gate')
    report=json.loads(rows[0]);report['estimated_ns']=report['sampled_ns']/report['timed_calls']*report['calls'];report['process']=sample
    (directory/'gate.json').write_text(json.dumps(report,indent=2)+'\n')

def checkpoint_prepare(directory,original,archive,valgrind):
    source=(original/'before.c').read_text()
    function=next(f for f in boundary.functions(source) if f['plain']=='abbreviation')
    caller=next(f for f in boundary.functions(source) if f['plain']=='noAbbreviatedIdentifier')
    old=function['name'];text=caller['text']
    if text.count(old+'(')!=1:raise ValueError('changed vocabulary caller')
    text=text.replace(old+'(','scout_profile_abbreviation(',1)
    source=source[:caller['start']]+text+source[caller['end']:]
    wrapper='''
static size_t scout_completed_lookups;
static adamic_object *scout_profile_abbreviation(adamic_string *text) {
 adamic_object *value=adamic_function_164_abbreviation(text);
 if ((++scout_completed_lookups % 65536)==0) { CALLGRIND_DUMP_STATS_AT("completed vocabulary calls"); }
 return value;
}
'''
    source=source[:function['end']]+wrapper+source[function['end']:]
    header=valgrind.parent.parent/'include/valgrind/callgrind.h'
    if not header.exists():raise ValueError('matching Valgrind headers required')
    source='#include "'+str(header)+'"\n'+source
    flags=boundary.build(directory,archive,'native-checkpoints',source)
    (directory/'checkpoint-build.json').write_text(json.dumps({'flags':flags,'interval_calls':65536,'ownership_changes':0},indent=2)+'\n')

def profile(directory,baseline,original,valgrind):
    from vocabulary_accounting import summarize,summarize_parts
    scout.USAGE_HELPER=baseline/'usage';manifest=baseline/'fused/manifest.txt'
    report={'machine_before':scout.machine(),'commands':{},'profile_environment':{'GOMAXPROCS':'1','GODEBUG':'asyncpreemptoff=1','VALGRIND_LIB':os.environ.get('VALGRIND_LIB','')}}
    env={**os.environ,**report['profile_environment']};want=(original/'compiler77/Go').read_bytes()
    for side,binary,function in (('native',directory/'native-checkpoints' if (directory/'native-checkpoints').exists() else original/'before','adamic_function_164_abbreviation'),('Go',baseline/'lint/oracle','github.com/system-inc/cohere/internal/lint/rules/nexus.(*abbreviationVocabulary).find')):
        profile_path=directory/(('native-checkpoints' if side=='native' and binary.name=='native-checkpoints' else side)+'.callgrind.out')
        argv=[valgrind,'--tool=callgrind','--cache-sim=yes','--collect-atstart=no','--toggle-collect='+function,'--callgrind-out-file='+str(profile_path),binary,'--manifest',manifest]
        report['commands'][side]=list(map(str,argv));out=directory/(side+'-callgrind.answer')
        scout.run(argv,out,env=env);scout.checked(out.read_bytes(),want)
        paths=[profile_path,*sorted(directory.glob(profile_path.name+'.*'))]
        (directory/(side+'-accounting.json')).write_text(json.dumps(summarize_parts(paths),indent=2)+'\n')
    report['machine_after']=scout.machine();(directory/'profile-environment.json').write_text(json.dumps(report,indent=2)+'\n')

def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('mode',choices=('prepare','measure','gate','profile','checkpoint-prepare'));p.add_argument('directory',type=Path);p.add_argument('--baseline',type=Path,required=True);p.add_argument('--original',type=Path,required=True);p.add_argument('--archive',type=Path);p.add_argument('--valgrind',type=Path);p.add_argument('--rounds',type=int,default=3);a=p.parse_args()
    if a.mode=='prepare':
        if not a.archive:p.error('--archive required')
        prepare(a.directory.resolve(),a.baseline.resolve(),a.original.resolve(),a.archive.resolve())
    elif a.mode=='checkpoint-prepare':
        if not a.archive or not a.valgrind:p.error('--archive and --valgrind required')
        checkpoint_prepare(a.directory.resolve(),a.original.resolve(),a.archive.resolve(),a.valgrind.resolve())
    elif a.mode=='profile':
        if not a.valgrind:p.error('--valgrind required')
        profile(a.directory.resolve(),a.baseline.resolve(),a.original.resolve(),a.valgrind.resolve())
    elif a.mode=='gate':gate(a.directory.resolve(),a.baseline.resolve(),a.original.resolve())
    else:measure(a.directory.resolve(),a.baseline.resolve(),a.original.resolve(),a.rounds)
if __name__=='__main__':main()
