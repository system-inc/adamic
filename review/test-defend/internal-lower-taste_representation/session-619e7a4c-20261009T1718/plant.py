import pathlib,json,difflib,subprocess
root=pathlib.Path.cwd();e=pathlib.Path(__file__).resolve().parent;s=pathlib.Path('/tmp/defend-taste')
plan=[('D1','internal/lower/object.go','return nil, l.notYet(node, "a narrowed scalar in a boxed union field")','return nil, l.notYet(node, "a boxed union field")','change constant','TestTasteRepresentationLimitsStayExplicit'),('D2','internal/lower/typed_arrays.go','made.Arguments != nil && len(made.Arguments.Nodes) > 1','made.Arguments != nil && len(made.Arguments.Nodes) > 2','off by one a bound','TestTypedArrayGaps'),('D3','internal/lower/object.go','\t\tproperty.ViewContract = l.result.ViewContractTypes[property.ViewTypeID]\n','','drop statement','TestViewObjectContractsAreAvailableToEraser')]
result=[]
for mid,file,old,new,menu,row in plan:
 text=(root/file).read_text();line=text[:text.index(old)].count('\n')+1;changed=text.replace(old,new,1);out=s/(mid+'.go');out.write_text(changed)
 diff=''.join(difflib.unified_diff(text.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file))
 (e/(mid+'.diff')).write_text(diff);(s/(mid+'-overlay.json')).write_text(json.dumps({'Replace':{str(root/file):str(out)}}))
 result.append(dict(mutant=mid,file_line=file+':'+str(line),change=old.strip()+' -> '+(new.strip() or '(drop)'),menu=menu,target=row))
(e/'plan.json').write_text(json.dumps(result,indent=2)+'\n')
