import pathlib,json
r=pathlib.Path('review/test-audit/internal-oracle-oct6_mutant');menu=json.loads((r/'menu.json').read_text());originals={m['file']:pathlib.Path(m['file']).read_text() for m in menu};pathlib.Path('/tmp/u066/originals.json').write_text(json.dumps(originals));changes=dict(originals)
f='internal/native/runtime/adamic.c';changes[f]=changes[f].replace('// write_all writes','static int u066_selected(const char *id) { const char *value = getenv("ADAMIC_MUTANT"); return value != NULL && strcmp(value, id) == 0; }\n\n// write_all writes',1)
for m in menu:
 id=m['id'];a=m['before'];s=changes[m['file']]
 if id=='M1':b='\t\t_exit(u066_selected("M1") ? 71 : 70);'
 elif id=='M2':b='if ((stream == adamic_stderr) != u066_selected("M2")) {'
 elif id=='M3':b='size_t whole = output_whole - (u066_selected("M3") ? 1 : 0);'
 elif id=='M4':b='if (len(files) != 1) != (os.Getenv("ADAMIC_MUTANT") == "M4") {'
 elif id=='W1':b=a+'\n\tif os.Getenv("ADAMIC_MUTANT") == "W1" { return "" }'
 elif id=='W2':b=a+'\n\tif os.Getenv("ADAMIC_MUTANT") == "W2" { return "" }'
 elif id=='W3':b='if err := native.Build(code, binary, native.Options{Sanitize: os.Getenv("ADAMIC_MUTANT") != "W3"}); err != nil {'
 elif id=='P1':b=a+'\n\tif os.Getenv("ADAMIC_MUTANT") == "P1" { return nil, nil }'
 else:b=a+'\n\tif (u066_selected("'+id+'")) { return; }'
 changes[m['file']]=s.replace(a,b,1)
for f,s in changes.items():
 if f in ['internal/lower/lower.go','internal/oracle/parameter_properties_test.go']:s=s.replace('import (','import (\n "os"',1)
 pathlib.Path(f).write_text(s)
