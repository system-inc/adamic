"""Pin Unicode.SimpleFold transitions and extract every consuming fixture rune."""
import gzip,hashlib,json,subprocess,tempfile
from pathlib import Path
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[5]
families={
 '@next/next/no-html-link-for-pages':['next/no_html_link_for_pages_test.go'],
 '@typescript-eslint/no-empty-object-type':['typescript/no_empty_object_type_test.go','typescript/no_empty_object_type_corpus_test.go','typescript/no_empty_object_type_corpus_data_test.go'],
 'no-restricted-exports':['core/no_restricted_exports_test.go'],
 'no-restricted-imports':['core/no_restricted_imports_test.go','core/no_restricted_imports_matcher_test.go'],
}
with tempfile.TemporaryDirectory(prefix='wave10-fold-generator-') as directory:
 scratch=Path(directory)
 def generate(template,name,args=[]):
  source=scratch/(name+'.go');source.write_text((HERE/template).read_text())
  binary=scratch/name
  subprocess.run(['go','build','-o',str(binary),str(source)],cwd=ROOT,check=True)
  output=scratch/(name+'.json')
  with output.open('wb') as out:subprocess.run([str(binary),*args],cwd=ROOT,stdout=out,check=True)
  return json.loads(output.read_text())
 table=generate('fold_table.go.txt','table')
 paths=[str(ROOT/'cohere/internal/lint/rules'/file) for files in families.values() for file in files]
 values=generate('literals.go.txt','literals',paths)
 cases=[];coverage=[]
 for rule,files in families.items():
  literals=[value for file in files for value in values[str(ROOT/'cohere/internal/lint/rules'/file)]]
  cases.extend(literals);coverage.append({'rule':rule,'files':files,'literals':len(literals),'runes':sum(map(len,literals))})
 encoded=json.dumps(cases,ensure_ascii=True,separators=(',',':')).encode()
 (HERE/'fold-cases.json.gz').write_bytes(gzip.compress(encoded,mtime=0))
 provenance={'go':table['go'],'unicode':table['unicode'],'transitions':len(table['from']),'sha256':hashlib.sha256(json.dumps(table,separators=(',',':')).encode()).hexdigest(),'consumers':coverage}
 (HERE/'fold-coverage.json').write_text(json.dumps(provenance,indent=2)+'\n')
 def array(name,values):
  lines=[','.join(map(str,values[i:i+24])) for i in range(0,len(values),24)]
  return 'const '+name+': readonly number[] = [\n    '+',\n    '.join(lines)+'\n];\n'
 source='// Unicode.SimpleFold transitions generated from '+table['go']+', Unicode '+table['unicode']+'.\n'
 source+=array('foldFrom',table['from'])+array('foldTo',table['to'])
 source+='''
function foldStep(r: number): number {
    let low = 0; let high = foldFrom.length;
    while(low < high) {
        const middle = Math.floor((low + high) / 2);
        const key = foldFrom[middle] ?? -1;
        if(key < r) { low = middle + 1; }
        else { high = middle; }
    }
    if((foldFrom[low] ?? -1) === r) { return foldTo[low] ?? r; }
    return r;
}
export function simpleFold(r: number): number {
    let least = r;
    for(let folded = foldStep(r); folded !== r; folded = foldStep(folded)) {
        if(folded < least) { least = folded; }
    }
    return least;
}
'''
 (HERE.parent/'regexp_simple_fold.a').write_text(source)
 print('Go',table['go'],'Unicode',table['unicode'],'transitions',len(table['from']),'fixture literals',len(cases),'fixture runes',sum(map(len,cases)),'total queries',1114113+6+sum(map(len,cases)))
