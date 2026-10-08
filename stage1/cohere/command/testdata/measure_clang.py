#!/usr/bin/env python3
"""Preserve stage 0's flags; archive its C objects and measure the final link."""
import json, pathlib, subprocess, sys, tempfile, time
clang='/workspace/adamic-tools/llvm/bin/clang'
args=sys.argv[1:]
if '--version' in args or '-c' in args or '-o' not in args:
    raise SystemExit(subprocess.call([clang,*args]))
output=args[args.index('-o')+1];flags=[];sources=[];libraries=[];i=0
while i<len(args):
    arg=args[i]
    if arg=='-o':i+=2;continue
    if arg.endswith('.c'):sources.append(arg)
    elif arg.endswith('.a') or arg in ['-lm','-lpthread','-ldl']:libraries.append(arg)
    else:flags.append(arg)
    i+=1
start=time.perf_counter()
with tempfile.TemporaryDirectory(prefix='scout42-link-') as directory:
    objects=[]
    for i,source in enumerate(sources):
        obj=str(pathlib.Path(directory)/f'{i}.o')
        subprocess.run([clang,*flags,'-c',source,'-o',obj],check=True);objects.append(obj)
    compiled=time.perf_counter()-start
    archive=output+'.a';subprocess.run(['ar','rcs',archive,*objects],check=True)
    start=time.perf_counter()
    subprocess.run([clang,*flags,'-o',output,'-Xlinker','--whole-archive',archive,'-Xlinker','--no-whole-archive',*libraries],check=True)
    link=time.perf_counter()-start
pathlib.Path(output+'.measure.json').write_text(json.dumps(dict(c_objects=len(objects),c_compile_seconds=compiled,link_seconds=link,binary_bytes=pathlib.Path(output).stat().st_size,program_runtime_archive_bytes=pathlib.Path(archive).stat().st_size),indent=2)+'\n')
