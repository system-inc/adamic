'use strict';
const assert = require('node:assert/strict'), fs = require('node:fs'), path = require('node:path');
const ts = require('typescript');
const {compiler} = require('../../47-host-errors/adapt.cjs');
const {validateBindings,plannedProgram} = require('../adapt.cjs');
function inspect(tree,done) { const p=compiler(tree,ts);return validateBindings(p,p.getSourceFile(path.join(tree,'src/compiler/sys.ts')),done).evidence; }
function mutants(tree) {
    const program=compiler(tree,ts),file=program.getSourceFile(path.join(tree,'src/compiler/sys.ts'));
    const cases=[
        ['module collision',file.text+'\nconst _crypto = {};\n'],
        ['enclosing parameter shadow',file.text.replace('function createSHA256Hash(data: string)','function createSHA256Hash(data: string, _crypto: typeof import("node:crypto"))')],
        ['bare namespace value',file.text+'\nvoid _crypto;\n'],
    ];
    return cases.map(([name,text])=>{
        assert.notEqual(text,file.text);const p=plannedProgram(program,file,text);let message='';
        try{validateBindings(p,p.getSourceFile(file.fileName),true);}catch(e){message=e.message;}
        assert.match(message,/collision or bare namespace use:|captured or shadowed use:/,name+' survived');
        console.log(`CAUGHT ${name}: ${message}`);return{name,message,caught:true};
    });
}
const result={before:inspect(path.resolve(process.argv[2]),false),after:inspect(path.resolve(process.argv[3]),true),mutants:mutants(path.resolve(process.argv[3]))};
fs.writeFileSync(process.argv[4],JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result,null,2));
