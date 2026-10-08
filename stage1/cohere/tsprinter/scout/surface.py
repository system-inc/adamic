"""Reproduce the pinned formatter's physical-source inventory; no coverage credit inferred."""
import json
import re
import subprocess
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
GO = ROOT / 'cohere'
PORT = ROOT / 'stage1/cohere'
MAPPING = {
    'arena': ('tsprinter/doc.ts', 'indexed document representation, not Go slab API'),
    'doc': ('tsprinter/doc.ts', 'document algebra; settings: width, tabs, indentation'),
    'estree': ('estree', 'non-JSON conversion; tsprinter uses parser directly'),
    'javascript': ('tsprinter', 'partial expressions and statements; no complete TS/TSX contract'),
    'css': ('css', 'CSS/SCSS parser and printer; default options'),
    'css/mediaquery': ('mediaquery', 'media query parser'),
    'css/selector': ('selector', 'selector parser'),
    'css/values': ('values', 'value parser'),
    'graphql': ('graphql', 'parser plus graphql/printer'),
    'yaml': ('yaml', 'parser/printer, default options'),
    'markdown': ('markdownblocks', 'partial construction/layout; embedding off'),
    'markdown/mdast': ('markdownblocks', 'construction from Go resolved events, not source parser'),
    'markdown/micromark': ('markdownblocks', 'event primitives; grammar unfinished'),
    'formatfiles': ('formatfiles', 'enumeration slice'),
    'formatoptions': ('', 'no dedicated options resolver port'),
    'native': ('', 'no composed multi-language router port'),
    'printing': ('', 'no generic shared print/comment/embedding core port'),
    'prettier': ('', 'Go oracle and bundled fork, not production Adamic'),
    'comparison': ('', 'Go comparison machinery'),
    'differential': ('', 'Go differential machinery'),
    'oracletest': ('', 'Go oracle helpers'),
}
FIELDS = ['TabWidth','UseTabs','Semi','SingleQuote','PrintWidth','TrailingComma','BracketSpacing','BracketSameLine','ArrowParens','EndOfLine']

def lines(path):
    return path.read_text().splitlines()

def inventory():
    files = []
    for path in sorted((GO/'internal/format').rglob('*.go')):
        relative = path.relative_to(GO)
        if 'testdata' in relative.parts or any(p.startswith('.') for p in relative.parts):
            continue
        content = lines(path)
        package = str(path.parent.relative_to(GO/'internal/format'))
        parent = package
        while parent not in MAPPING and '/' in parent:
            parent = parent.rsplit('/',1)[0]
        destination, scope = MAPPING.get(parent, ('','unmapped tooling'))
        files.append(dict(path=str(relative), package=package, lines=len(content), test=path.name.endswith('_test.go'),
            port=('stage1/cohere/'+destination if destination else None), scope=scope,
            functions=[dict(line=i, declaration=line.strip()) for i,line in enumerate(content,1) if re.match(r'^func\s',line)],
            option_reads=[dict(line=i,field=field) for i,line in enumerate(content,1) for field in FIELDS if re.search(r'\.'+field+r'\b',line)]))
    option_structs=[]
    for record in files:
        if record['test']:continue
        content=lines(GO/record['path'])
        for i,line in enumerate(content):
            if re.match(r'^type \w*[Oo]ptions(?:\[.*\])? struct \{',line):
                end=i+1
                while end<len(content) and content[end]!='}':end+=1
                option_structs.append(dict(path=record['path'],line=i+1,declaration=line,fields=content[i+1:end]))
    packages = {}
    for record in files:
        p=packages.setdefault(record['package'], dict(files=0,lines=0,test_files=0,test_lines=0,port=record['port'],scope=record['scope']))
        prefix='test_' if record['test'] else ''
        p[prefix+'files']+=1
        p[prefix+'lines']+=record['lines']
    return dict(adamic=subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
        cohere=subprocess.check_output(['git','rev-parse','HEAD'],cwd=GO,text=True).strip(),
        convention='physical lines, including blanks/comments/generated/platform variants; excludes testdata; tests separate; lexical function starts and option member reads, not semantic coverage',
        packages=packages,files=files,option_structs=option_structs,
        port_sources=[dict(path=str(path.relative_to(ROOT)),lines=len(lines(path))) for path in sorted(PORT.rglob('*.ts')) if not any(part in {'testdata','gaps','scout'} for part in path.relative_to(PORT).parts)],
        totals=dict(Counter({'production_files':sum(not f['test'] for f in files),'production_lines':sum(f['lines'] for f in files if not f['test']), 'test_files':sum(f['test'] for f in files),'test_lines':sum(f['lines'] for f in files if f['test'])})))

if __name__=='__main__':
    result=inventory()
    (Path(__file__).parent/'surface.json').write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps({'totals':result['totals'],'packages':result['packages']},indent=2))
