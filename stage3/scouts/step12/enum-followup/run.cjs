// Stock TypeScript transpilation on Node is the oracle, never Adamic JS.
const fs = require('node:fs');
const path = require('node:path');
const {spawnSync} = require('node:child_process');
const ts = require(process.env.SCOUT_TYPESCRIPT);
if (ts.version !== '6.0.3') throw Error('requires TypeScript 6.0.3');
const out = path.resolve(process.argv[2]);
fs.mkdirSync(out);
const compilers = JSON.parse(fs.readFileSync(process.argv[3], 'utf8'));
const golden = '128 128\n83 83\n80 80\n80 80\n80 80\n';
const rows = [
 ['identifier-return', golden, '"abstract", "break", "name"', '"break", "break", "name"', 'BinaryExpression'],
 ['identifier-full-enum', golden, '"abstract", "break", "name"', '"break", "break", "name"', 'BinaryExpression'],
 ['identifier-split', golden, 'token = keyword;', 'token = SyntaxKind.Identifier;', ''],
 ['unproven-member', '99\n', 'Math.trunc(99)', 'Math.trunc(100)', 'unproven value assigned'],
];
function run(cmd,args,stem) {
 const r=spawnSync(cmd,args); if(r.error) throw r.error;
 fs.writeFileSync(path.join(out,stem+'.stdout'),r.stdout);
 fs.writeFileSync(path.join(out,stem+'.stderr'),r.stderr);
 return r;
}
function node(source,stem) {
 const file=path.join(out,stem+'.cjs');
 fs.writeFileSync(file,ts.transpileModule(source,{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.CommonJS}}).outputText);
 return run(process.execPath,[file],stem);
}
const results=[];
for(const [name,expected,before,after,diagnostic] of rows) {
 const file=path.join(__dirname,name+'.a'); const source=fs.readFileSync(file,'utf8');
 if(!source.startsWith('// a-check:') || source.split(before).length!==2) throw Error('invalid fixture '+name);
 const control=node(source,name+'.node');
 if(control.status!==0 || control.stderr.length || control.stdout.toString()!==expected) throw Error('Node golden '+name);
 const mutant=node(source.replace(before,after),name+'.mutant');
 if(mutant.status===control.status && mutant.stdout.equals(control.stdout) && mutant.stderr.equals(control.stderr)) throw Error('surviving mutant '+name);
 for(const [label,compiler] of Object.entries(compilers)) {
  const stem=name+'.'+label; const binary=path.join(out,stem+'.native');
  const build=run(compiler,['build',file,'-o',binary],stem+'.build');
  if(build.status!==(diagnostic?1:0) || !build.stderr.toString().includes(diagnostic)) throw Error('unexpected build '+stem+': '+build.stderr);
  let equal=null;
  if(!diagnostic) {
   const native=run(binary,[],stem+'.native');
   equal=native.status===0 && native.stdout.equals(control.stdout) && !native.stderr.length;
   if(!equal) throw Error('native mismatch '+stem);
   const mutationFile=path.join(out,name+'.mutant.a'); fs.writeFileSync(mutationFile,source.replace(before,after));
   const mb=run(compiler,['build',mutationFile,'-o',binary+'.mutant'],stem+'.mutant.build');
   if(mb.status!==0) throw Error('mutant must build '+stem);
   const mn=run(binary+'.mutant',[],stem+'.mutant.native');
   if(mn.status!==0 || !mn.stdout.equals(mutant.stdout) || mn.stdout.equals(control.stdout)) throw Error('native mutant catcher '+stem);
  }
  let provenControl = null;
  if(name==='unproven-member') {
   const provenFile=path.join(out,'proven-member.a');
   fs.writeFileSync(provenFile,source.replace('Math.trunc(99)','SyntaxKind.Identifier'));
   const pb=run(compiler,['build',provenFile,'-o',binary+'.proven'],stem+'.proven.build');
   if(pb.status!==0) throw Error('proven constant refused '+stem);
   const pn=run(binary+'.proven',[],stem+'.proven.native');
   if(pn.status!==0 || pn.stdout.toString()!=='80\n' || pn.stderr.length) throw Error('proven member wrong '+stem);
   provenControl=true;
  }
  results.push({proven_control:provenControl,name,compiler:label,node_stdout:expected,mutant_stdout:mutant.stdout.toString(),mutant_caught:true,build_exit:build.status,diagnostic:build.stderr.toString(),native_equal:equal});
 }
}
fs.writeFileSync(path.join(out,'results.json'),JSON.stringify(results,null,2)+'\n');
console.log('4 Node goldens, 4 source mutants, 15 build expectations, 6 native controls and 3 native mutants passed');
