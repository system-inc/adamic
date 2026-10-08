"""Independent known-answer fixture and a real dropped-rewrite mutant."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[1]


def run(args,log,expected=0):
    with log.open('w') as stream:
        result=subprocess.run(args,cwd=ROOT,stdout=stream,stderr=subprocess.STDOUT)
    assert result.returncode==expected,(args,result.returncode,log.read_text())


def main():
    scratch=Path(tempfile.mkdtemp(prefix='step09-fixture-'))
    before=scratch/'before';after=scratch/'after'
    before.mkdir();after.mkdir()
    source=(HERE/'fixture/main.a').read_text()
    (before/'main.ts').write_text(source)
    (after/'main.ts').write_text(source.replace('value: any','value: number'))
    transitions=[dict(adaptation='fixture-40',before=str(before),after=str(after))]
    (scratch/'transitions.json').write_text(json.dumps(transitions))
    probe=os.environ['STEP09_CAST_PROBE']
    for name,tree in [('stock',before),('adapted',after)]:
        run(['node',str(HERE/'enumerate.cjs'),str(tree),str(scratch/(name+'.json')),'main.ts'],scratch/(name+'.txt'))
    stock=json.loads((scratch/'stock.json').read_text())
    assert len(stock['sites'])==3
    assert sorted(s['kind'] for s in stock['sites'])==['any_declaration','as_cast','explicit_any']
    run([probe,str(after/'main.ts')],scratch/'casts.json')
    casts=json.loads((scratch/'casts.json').read_text());assert len(casts)==1 and casts[0]['state']=='upcast',casts
    args=['python3',str(HERE/'ledger.py'),str(scratch/'stock.json'),str(scratch/'adapted.json'),str(scratch/'transitions.json'),str(scratch/'casts.json')]
    run(args+[str(scratch/'result')],scratch/'ledger.txt')
    result=json.loads((scratch/'result/SUMMARY.json').read_text())
    expected={'rewritten':2,'proven_safe_upcast':1}
    assert result['dispositions']==expected,result
    run(args+[str(scratch/'mutant'),'--mutant-ignore-rewrites'],scratch/'mutant.txt')
    mutated=json.loads((scratch/'mutant/SUMMARY.json').read_text())
    # Independent expected answer rejects the classification bug, not a parse/build failure.
    assert mutated['dispositions']!=expected and mutated['dispositions'].get('open')==2,mutated
    script=scratch/'node-fixture.cjs'
    script.write_text("const ts=require('typescript'),fs=require('node:fs');const source=fs.readFileSync(process.argv[2],'utf8');const js=ts.transpileModule(source,{compilerOptions:{module:ts.ModuleKind.CommonJS}}).outputText;const observed={};new Function('exports',js)(observed);if(observed.value!==1||observed.widened!==1)throw Error('fixture changed');console.log('value=1 widened=1');")
    run(['node',str(script),str(before/'main.ts')],scratch/'node.txt')
    (before/'nested.ts').write_text((HERE/'fixture/nested.a').read_text())
    run(['node',str(HERE/'enumerate.cjs'),str(before),str(scratch/'nested.json'),'nested.ts'],scratch/'nested-enumerate.txt')
    run([probe,str(before/'nested.ts')],scratch/'nested-casts.json')
    nested_casts=json.loads((scratch/'nested-casts.json').read_text())
    assert len(nested_casts)==2 and nested_casts[0]['byte_start']==nested_casts[1]['byte_start']
    assert nested_casts[0]['byte_end']!=nested_casts[1]['byte_end']
    assert sorted(c['state'] for c in nested_casts)==['refused','upcast'],nested_casts
    (scratch/'empty-transitions.json').write_text('[]')
    run(['python3',str(HERE/'ledger.py'),str(scratch/'nested.json'),str(scratch/'nested.json'),str(scratch/'empty-transitions.json'),str(scratch/'nested-casts.json'),str(scratch/'nested-result')],scratch/'nested-ledger.txt')
    nested_summary=json.loads((scratch/'nested-result/SUMMARY.json').read_text())
    assert nested_summary['dispositions']=={'open':1,'proven_safe_upcast':1},nested_summary
    evidence=dict(stock_sites=3,expected_dispositions=expected,observed=result['dispositions'],mutant=mutated['dispositions'],mutant_caught=True,nested_casts_distinct=True,nested_dispositions=nested_summary['dispositions'],node=(scratch/'node.txt').read_text(),logs=str(scratch))
    (scratch/'RESULT.json').write_text(json.dumps(evidence,indent=2)+'\n')
    print(json.dumps(evidence,indent=2))


if __name__=='__main__':main()
