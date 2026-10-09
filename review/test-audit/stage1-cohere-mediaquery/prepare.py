import pathlib,json,difflib
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-mediaquery';(out/'diffs').mkdir(exist_ok=True);plan=[]
def add(id,file,old,new,kind):
 s=(root/file).read_text();assert s.count(old)==1,(id,s.count(old));n=s.replace(old,new,1)
 if id=='P2':n=n.replace('\t"fmt"\n','').replace('\t"path/filepath"\n','')
 (out/'diffs'/f'{id}.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),n.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 plan.append(dict(id=id,file=file,line=s[:s.index(old)].count('\n')+1,old=old,new=new,kind=kind))
base='stage1/cohere/mediaquery/'
add('M1',base+'nodes.ts','return new MediaNode(type, after, before, value, sourceIndex, undefined);','return new MediaNode(type, before, after, value, sourceIndex, undefined);','production')
add('M2',base+'whitespace.ts','case 0x09:','case 0x08:','production')
add('M3',base+'parsers.ts',"!rest.startsWith('url')","!rest.startsWith('URL')",'production')
add('M4','internal/lower/generic.go','const maximumGenericDepth = 32','const maximumGenericDepth = 0','production')
s=(root/(base+'mediaquery_test.go')).read_text();a=s.index('func firstDifference(');b=s.index('\n// lowered',a)
add('W1',base+'mediaquery_test.go',s[a:b],'func firstDifference(got string, want string) string { return "" }\n','witness')
s=(root/(base+'main.ts')).read_text();a=s.index('const casesPath =')
add('P1',base+'main.ts',s[a:],'','probe')
s=(root/'internal/lower/lower.go').read_text();a=s.index('func Lower(');b=s.index('\ntype lowering struct',a)
add('P2','internal/lower/lower.go',s[a:b],'func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n\treturn nil, nil\n}\n','probe')
(out/'plan.json').write_text(json.dumps(plan,indent=2)+'\n')
