"""Independent Go strconv.Quote comparison over printable Unicode range boundaries."""
import json
from pathlib import Path

def validate_quote(scratch, run, stage0, owned, repository):
    raw,_=run('quote-go-oracle',['go','run',owned/'quote_oracle.go'])
    reference=json.loads(raw);expected=reference['Expected'].encode()
    entry=scratch/'quote-check.a'
    entry.write_text("import { quote } from "+repr(str(owned/'jsx_no_constructed_context_values/quote.a'))+";\nimport { written } from "+repr(str(repository/'stage1/typescript/parser/nodes.ts'))+";\nconst names: readonly string[] = "+json.dumps(reference['Names'])+";\nfor(const name of names) { console.log(written(quote(name))); }\n")
    for sanitize in [False,True]:
        binary=scratch/('quote-native-asan' if sanitize else 'quote-native')
        run(binary.name+'-build',[stage0,'build',entry,'-o',binary]+(['--sanitize'] if sanitize else []))
        got,err=run(binary.name+'-run',[binary]);assert not err and got==expected,binary.name
    got,err=run('quote-source-node',['node','--disable-warning=ExperimentalWarning',repository/'oracle/node.mjs',entry]);assert not err and got==expected
    emitted,err=run('quote-emitted-build',[stage0,'js',entry]);assert not err
    js=scratch/'quote-emitted.mjs';js.write_bytes(emitted)
    got,err=run('quote-emitted-node',['node','--disable-warning=ExperimentalWarning',repository/'oracle/node.mjs',js]);assert not err and got==expected
    original=(owned/'jsx_no_constructed_context_values/quote.a').read_text()
    observations=[]
    for name,before,after in [
        ('newline', 'code === 10', 'code === -1'),
        ('ascii-control', "padStart(2, '0')", "padStart(1, '0')"),
        ('unicode-control', "padStart(4, '0')", "padStart(3, '0')"),
        ('supplementary-control', "padStart(8, '0')", "padStart(7, '0')"),
    ]:
        assert original.count(before)==1
        module=scratch/('quote-'+name+'-mutant.a');module.write_text(original.replace(before,after,1))
        fixture=scratch/('quote-'+name+'-check.a');fixture.write_text(entry.read_text().replace(repr(str(owned/'jsx_no_constructed_context_values/quote.a')),repr(str(module))))
        binary=scratch/('quote-'+name+'-mutant');run(binary.name+'-build',[stage0,'build',fixture,'-o',binary])
        got,err=run(binary.name+'-run',[binary]);assert not err and got!=expected,name
        observations.append(dict(name=name,first_difference=next(i for i,(a,b) in enumerate(zip(got,expected)) if a!=b)))
    guard=scratch/'quote-surrogate-guard.a';guard.write_text("import { quote } from "+repr(str(owned/'jsx_no_constructed_context_values/quote.a'))+";\nconsole.log(quote('\\ud800'));\n")
    binary=scratch/'quote-surrogate-guard';run(binary.name+'-build',[stage0,'build',guard,'-o',binary])
    got,err=run(binary.name+'-run',[binary],expected=70);assert not got and err==b'adamic: panic: quoted name contains an unpaired surrogate\n'
    module=scratch/'quote-surrogate-mutant.a';module.write_text(original.replace('if(code >= 55296', 'if(false && code >= 55296',1))
    fixture=scratch/'quote-surrogate-mutant-check.a';fixture.write_text(guard.read_text().replace(repr(str(owned/'jsx_no_constructed_context_values/quote.a')),repr(str(module))))
    binary=scratch/'quote-surrogate-mutant';run(binary.name+'-build',[stage0,'build',fixture,'-o',binary]);got,err=run(binary.name+'-run',[binary]);assert got and not err
    observations.append(dict(name='surrogate-guard',normal_exit=70,mutant_exit=0))
    return dict(unicode_version=reference['Version'],cases=len(reference['Names']),bytes=len(expected),mutants=observations)
