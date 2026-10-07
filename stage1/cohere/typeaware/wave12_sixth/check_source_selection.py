"""Source-selection subset only: no checker-backed binding or context analysis."""
import json
from pathlib import Path
import subprocess
import sys

owned = Path(__file__).resolve().parent
repository = owned.parents[3]
scratch = Path(sys.argv[1]); scratch.mkdir(parents=True, exist_ok=True)
commands = []

def run(name, args, directory=repository, expected=0):
    with (scratch / (name + '.stdout')).open('wb') as out, (scratch / (name + '.stderr')).open('wb') as err:
        result = subprocess.run(list(map(str, args)), cwd=directory, stdout=out, stderr=err)
    commands.append(dict(name=name, args=list(map(str, args)), code=result.returncode))
    stdout = (scratch / (name + '.stdout')).read_bytes()
    stderr = (scratch / (name + '.stderr')).read_bytes()
    assert result.returncode == expected, (name, result.returncode, stderr.decode())
    return stdout, stderr

undef = ['<Missing />', '<Missing.Inner />', '<Missing.Inner.Deep />', '<lower.Inner />',
         '<div />', '<custom-element />', '<ns:tag />', '<this.Component />', '<this />',
         '<Ω />', '<_Unknown />', '<$Unknown />', '<a-b />', '< /*before*/ Missing />',
         '<Missing></Missing>', '<Missing<T> />']
fragments = ['<React.Fragment />', '<React.Fragment></React.Fragment>', '<React.Fragment<T> />',
             '<React.Fragment key="k" />', '<React.Fragment {...props} />',
             '<A.React.Fragment />', '<React.Other />', '<></>',
             '<React.Fragment><React.Fragment /></React.Fragment>',
             '< /*before*/ React.Fragment />']
paths=[]; modes=[]
for mode, texts in [(1, undef), (2, fragments)]:
    for index, text in enumerate(texts):
        path=scratch/f'{mode}-{index}.tsx'
        prefix='export {};\n' if mode == 1 else 'export {}; declare const React: any; declare const A: any; declare const props: any;\n'
        # Multibyte source before a finding proves byte offsets are built from UTF-16 parser spans.
        path.write_text(prefix+'/* 😀 Ω */ const x = '+text+';\n')
        paths.append(path); modes.append(mode)
config=scratch/'tsconfig.json'
config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','jsx':'preserve','noEmit':True},'files':list(map(str,paths))}))
manifest=scratch/'manifest'; manifest.write_text('\n'.join(map(str,paths))+'\n')
virtual=repository/'cohere/adamic_wave12_selection_oracle.go'
overlay=scratch/'overlay.json'; overlay.write_text(json.dumps({'Replace':{str(virtual):str(owned/'oracle.go')}}))
oracle=scratch/'oracle'; run('oracle-build',['go','build','-overlay',overlay,'-o',oracle,virtual],repository/'cohere')
production,_=run('production-go',[oracle,config,manifest])
expected=[]; current=-1; findings=0
for row in production.decode().splitlines():
    if row.startswith('file\t'):
        current+=1; expected.append(row)
    elif not row.startswith('findings '):
        rule=row.split('\t')[2]
        target='react/jsx-no-undef' if modes[current] == 1 else 'react/jsx-fragments'
        if rule == target:
            expected.append(row); findings+=1
expected.append(f'findings {findings}')
expected=('\n'.join(expected)+'\n').encode(); (scratch/'expected').write_bytes(expected)
assert findings >= 15, findings
# The production stream is consumed only as expected output. It supplies no native spans or verdicts.
reference=scratch/'tag_reference.a'; reference.write_text((owned/'jsx_no_undef/tag_reference.a').read_text().replace("'../../../../typescript/parser/",repr(str(repository/'stage1/typescript/parser')+'/')[:-1]))
selection=scratch/'tag_selection.a'; selection.write_text((owned/'jsx_fragments/tag_selection.a').read_text().replace("'../../../../typescript/parser/",repr(str(repository/'stage1/typescript/parser')+'/')[:-1]))
imports=f'''import {{ panic, programArguments, readTextFile, utf8Length }} from 'adamic';
import {{ Parser }} from '{repository}/stage1/typescript/parser/parser.ts';
import {{ ParseNode, written }} from '{repository}/stage1/typescript/parser/nodes.ts';
import {{ Scanner }} from '{repository}/stage1/typescript/scanner/scanner.ts';
import {{ Diagnostic }} from '{repository}/stage1/cohere/typeaware/diagnostic.ts';
import {{ SuppliedNode }} from '{owned}/node.a';
import {{ report as undef }} from '{owned}/jsx_no_undef/messages.a';
import {{ report as fragment }} from '{owned}/jsx_fragments/messages.a';
import {{ tagReference }} from './tag_reference.a';
import {{ namedSelection }} from './tag_selection.a';
'''
entry=scratch/'selection-main.a'
entry.write_text(imports+'''
const args = programArguments();
let total = 0;
for(let index = 0; index < args.length; index += 2) {
    const path = args[index] ?? panic('missing path');
    const mode = args[index + 1] ?? panic('missing selection mode');
    const input = readTextFile(path);
    if(input.kind === 'Error') { panic(input.message); }
    const source = input.text;
    const parser = new Parser(source, path);
    const root = parser.file();
    const scanner = new Scanner(source);
    const listeners = new Map<string, number>();
    if(mode === '1') {
        listeners.set('JsxOpeningElement', 1);
        listeners.set('JsxSelfClosingElement', 1);
    }
    else {
        listeners.set('JsxElement', 3);
        listeners.set('JsxSelfClosingElement', 2);
    }
    const diagnostics: string[] = [];
    const pending: number[] = [root];
    while(pending.length > 0) {
        const current = pending.pop() ?? panic('missing pending node');
        const node = parser.node(current); // Fetch once before kind-indexed dispatch.
        const listener = listeners.get(node.kind) ?? 0;
        let finding: ParseNode | undefined;
        if(listener === 1) {
            finding = tagReference(parser, node);
        }
        else if(listener > 1) {
            const opening = listener === 3
                ? parser.node(node.children[0] ?? panic('missing opening')) : node;
            const selected = namedSelection(parser, opening);
            if(selected === 2) { panic('jsx-fragments needs checker declaration resolution'); }
            if(selected === 1) { finding = node; }
        }
        if(finding !== undefined) {
            scanner.pos = finding.pos;
            scanner.scan();
            const supplied = new SuppliedNode(utf8Length(source.slice(0, scanner.start)),
                utf8Length(source.slice(0, finding.end)));
            const diagnostic = listener === 1 ? undef(supplied) : fragment(supplied, false);
            diagnostics.push(diagnostic.written());
        }
        for(let child = node.children.length - 1; child >= 0; child--) {
            pending.push(node.children[child] ?? panic('missing child'));
        }
    }
    diagnostics.sort((left, right) => left < right ? -1 : left > right ? 1 : 0);
    console.log('file\\t' + written(path));
    for(const diagnostic of diagnostics) { console.log(diagnostic); }
    total += diagnostics.length;
}
console.log(`findings ${total}`);
''')
stage0=scratch/'adamic'; run('stage0',['go','build','-o',stage0,'./cmd/adamic'])
arguments=[value for path,mode in zip(paths,modes) for value in [path,str(mode)]]
for sanitized in [False, True]:
    binary=scratch/('native-asan' if sanitized else 'native')
    run(binary.name+'-build',[stage0,'build',entry,'-o',binary]+(['--sanitize'] if sanitized else []))
    got,err=run(binary.name+'-run',[binary,*arguments]); assert not err and got == expected, binary.name
source,err=run('source-node',['node','--disable-warning=ExperimentalWarning',repository/'oracle/node.mjs',entry,*arguments]); assert not err and source == expected
emitted,_=run('emitted-build',[stage0,'js',entry]); emitted_path=scratch/'emitted.mjs'; emitted_path.write_bytes(emitted)
got,err=run('emitted-node',['node','--disable-warning=ExperimentalWarning',repository/'oracle/node.mjs',emitted_path,*arguments]); assert not err and got == expected
mutants=[]
for name,file,before,after in [('undef-ascii',reference,'first >= 97','first >= 65'),('fragment-member',selection,"name.text === 'Fragment'","name.text === 'FragmentX'")]:
    original=file.read_text(); assert original.count(before)==1
    file.write_text(original.replace(before,after))
    binary=scratch/name;run(name+'-build',[stage0,'build',entry,'-o',binary]); got,err=run(name+'-run',[binary,*arguments])
    assert not err and got != expected
    mutants.append(dict(name=name,exit=0,stderr_bytes=0,first_difference=next(i for i,(a,b) in enumerate(zip(got,expected)) if a!=b)))
    file.write_text(original)
alias=scratch/'alias.tsx'; alias.write_text('export {}; const F = React.Fragment; const x = <F />;\n')
got,err=run('unresolved-alias',[scratch/'native',alias,'2'],expected=70)
assert not got and err == b'adamic: panic: jsx-fragments needs checker declaration resolution\n'
original=selection.read_text(); assert original.count('return 2;')==1
selection.write_text(original.replace('return 2;', 'return 0;'))
binary=scratch/'unresolved-bypass';run('unresolved-bypass-build',[stage0,'build',entry,'-o',binary]);got,err=run('unresolved-bypass-run',[binary,alias,'2']);assert got and not err
selection.write_text(original)
summary=dict(source_selection_only=True,cases=len(paths),findings=findings,bytes=len(expected),backends=['native','native-asan','source-node','emitted-node'],mutants=mutants,unresolved_alias=dict(normal_exit=70,bypass_exit=0),commands=commands)
(scratch/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print('SOURCE SELECTION SUBSET PASS',len(paths),'cases',findings,'findings',len(expected),'bytes; two byte-only mutants; unresolved alias refused')
