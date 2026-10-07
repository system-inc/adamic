from pathlib import Path
import subprocess, os, json, runpy
root=Path('/workspace/adamic'); report=root/'cloud/reports/string-views'
share=root/'internal/native/runtime/string_share.c'; slice=root/'internal/native/runtime/string_slice_impl.h'
mutants=[
('offset',share,'shared->bytes = string->bytes + offset;','shared->bytes = string->bytes + offset + 1;','TestRuntimeStringViews','./internal/native'),
('missing_hold',share,'adamic_retain((adamic_string *)owner);','(adamic_string *)owner;','TestNativeAgreesWithNode/internal/oracle/testdata/shared_slices','./internal/oracle'),
('extra_hold',share,'adamic_retain((adamic_string *)owner);','adamic_retain(adamic_retain((adamic_string *)owner));','TestNativeAgreesWithNode/internal/oracle/testdata/shared_slices','./internal/oracle'),
('spare_capacity',share,'owner->capacity > owner->length ? owner->capacity : owner->length','owner->length','TestRuntimeStringViews','./internal/native'),
('ratio',share,'#define SHARE_FRACTION 8','#define SHARE_FRACTION 16','TestRuntimeStringViews','./internal/native'),
('borrowed',share,'bool borrowed = owner->heap.references == 0 && owner->index == NULL;','bool borrowed = false;','TestRuntimeStringViews','./internal/native'),
('ascii_value',slice,'bytes[value] = (char)value;','bytes[value] = (char)(value ^ 1);','TestRuntimeStringViews','./internal/native'),
('ascii_allocation',slice,'return ascii_character((unsigned char)string->bytes[(size_t)index]);','adamic_string *temporary = adamic_string_share(string, (size_t)index, 1); adamic_release(temporary); return ascii_character((unsigned char)string->bytes[(size_t)index]);','TestRuntimeStringViews','./internal/native'),
]
results=[]
for name,path,old,new,test,package in mutants:
    original=path.read_text(); assert original.count(old)==1,(name,original.count(old))
    logfile=Path('/tmp/string-views-mutant-'+name+'.log')
    try:
        path.write_text(original.replace(old,new))
        command=['go','test',package,'-run',test,'-count=1','-v','-timeout','30m']
        with logfile.open('wb') as out:
            result=subprocess.run(command,cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=out,stderr=subprocess.STDOUT)
        results.append(dict(name=name,exit=result.returncode,command=command,log=str(logfile)))
        print(name,result.returncode,flush=True)
        assert result.returncode!=0,name+' survived'
        assert '[-Werror' not in logfile.read_text(),name+' killed by warning'
    finally: path.write_text(original)
(report/'mutants.json').write_text(json.dumps(results,indent=2)+'\n')
