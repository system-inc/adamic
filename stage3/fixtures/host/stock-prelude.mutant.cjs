// Evidence harness: old|new real|spaces|removed|duplicate|process-added hash|full logs
const fs=require('node:fs'),vm=require('node:vm'),path=require('node:path');
const root=path.resolve(__dirname,'../../..'), filename=__dirname+'/stock-check.cjs';
const [version,mutation,mode]=process.argv.slice(2);
let text=fs.readFileSync(root+'/internal/load/prelude.d.ts','utf8');
if(mutation!=='real')text=require(root+'/stage3/fixtures/host/'+(mutation==='process-added'?'process-added':'console-'+mutation)+'.mutant.cjs')(text);
const fakeFS={...fs,readFileSync(file,...args){return file===root+'/internal/load/prelude.d.ts'?text:fs.readFileSync(file,...args);}};
const req=require('node:module').createRequire(filename);
let code=fs.readFileSync(version==='old'?process.env.STOCK_CHECK_BEFORE:filename,'utf8');
if(mode==='hash')code=code.slice(0,code.indexOf('const mutant='))+"\nconsole.log(require('node:crypto').createHash('sha256').update(prelude).digest('hex'));";
vm.runInNewContext(code,{require:name=>name==='fs'?fakeFS:req(name),process:{...process,argv:['node',filename,root,path.resolve(process.argv[5])]},__dirname:path.dirname(filename),__filename:filename,console,Buffer},{filename});
