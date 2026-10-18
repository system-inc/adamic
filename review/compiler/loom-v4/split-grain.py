from pathlib import Path
import re
for file, test, kind in [
 ('internal/oracle/checked_views_arrays_test.go','TestCheckedViewArrays','checkedViewArrayProbe'),
 ('internal/oracle/checked_views_arrays_mutants_test.go','TestCheckedViewArrayReadMutants','checkedViewArrayMutantProbe'),
 ('internal/oracle/checked_views_array_union_test.go','TestCheckedViewArrayUnionMembership','checkedViewArrayUnionProbe'),
 ('internal/native/view_callables_contract_test.go','TestViewCallableShapeNative','viewCallableShapeProbe'),
 ('internal/native/view_callables_signature_test.go','TestViewCallableProducerCertificateNative','viewCallableCertificateProbe'),
]:
 p=Path(file);s=p.read_text();start=s.index('func '+test+'(');table=s.index('range []struct',start);opening=s.index('{',table);sep=s.index('}{',opening);definition=s[opening:sep+1];end=s.index('\n\t} {',sep);rows=s[sep+2:end]; bodyStart=s.index('\n',s.index('t.Run(',end))+1;bodyEnd=s.index('\n\t\t})\n\t}\n}',bodyStart);after=bodyEnd+len('\n\t\t})\n\t}\n}');body=s[bodyStart:bodyEnd];body='\n'.join(line[2:] if line.startswith('\t\t') else line for line in body.splitlines());body=body.replace('\n\tt.Parallel()','')
 literals=re.findall(r'\{[^\n{}]*\}',rows);wrappers=[]
 for row in literals:
  name=re.search(r'"([^"]+)"',row).group(1);suffix=''.join(x.title() for x in name.split('-'));wrappers.append('func '+test+suffix+'(t *testing.T) {\n\tt.Parallel()\n\t'+kind+'Check(t, '+kind+row+')\n}')
 if not literals: raise ValueError(file)
 replacement='type '+kind+' struct '+definition+'\n\nfunc '+kind+'Check(t *testing.T, probe '+kind+') {\n\tt.Helper()\n'+body+'\n}\n\n'+'\n\n'.join(wrappers)
 p.write_text(s[:start]+replacement+s[after:])
