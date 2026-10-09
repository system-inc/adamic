import pathlib,subprocess,time,json,difflib,os
root=pathlib.Path.cwd(); out=root/'review/test-audit/internal-native-normalize'; source=root/'internal/native/runtime/normalize.c'; original=source.read_text(); meta=[]
flags=['-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all']
def run(label,command,env=None):
    start=time.monotonic()
    with (out/(label+'.log')).open('w') as log:
        result=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,env=env)
    item={'label':label,'command':command,'seconds':round(time.monotonic()-start,3),'exit':result.returncode}
    if env: item['cache']=env.get('ADAMIC_BUILD_CACHE_DIR')
    meta.append(item); (out/'commands.json').write_text(json.dumps(meta,indent=2)+'\n'); print(json.dumps(item),flush=True)
    return result.returncode
try:
    for n in [2,3]:
        assert run('timing-family-'+str(n),['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run','^TestNormalizeMatchesNode'])==0,'clean family red'
    for n in [1,2,3]:
        assert run('timing-random-'+str(n),['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run','^TestNormalizeRandomMatchesNode$'])==0,'clean random red'
    variants=[('M1','hangul_t_count = 28','hangul_t_count = 27'),('M2','mapping == NULL || (mapping->compatibility && !compatibility)','mapping == NULL || (mapping->compatibility && compatibility)'),('M3','composite(list->items[starter], point)','composite(point, list->items[starter])'),('P1','adamic_string *adamic_string_normalize(const adamic_string *string, const adamic_string *form) {','adamic_string *adamic_string_normalize(const adamic_string *string, const adamic_string *form) {\n\treturn adamic_string_allocate(0);')]
    for name,old,new in variants:
        assert old in original
        changed=original.replace(old,new,1)
        source.write_text(changed)
        (out/(name+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/internal/native/runtime/normalize.c',tofile='b/internal/native/runtime/normalize.c')))
        assert run(name+'-clang',['clang',*flags,'-I','internal/native/runtime','-fsyntax-only',str(source)])==0
        env=os.environ.copy(); env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u050/cache/'+name
        assert run(name+'-build',['timeout','90','go','run','./review/test-audit/internal-native-normalize/buildhelper',str(out/'harness.c'),'/tmp/u050-'+name+'-harness'],env)==0
        run(name,['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run','^TestNormalize(MatchesNode|RandomMatchesNode$)'],env)
finally:
    source.write_text(original)
