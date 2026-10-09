// Mutate the real emitted C. Every mutant must compile and fail semantically.
const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process'),assert=require('node:assert/strict');
const [scratchArg,runtimeArg]=process.argv.slice(2),scratch=path.resolve(scratchArg),runtime=path.resolve(runtimeArg);
const rows=JSON.parse(fs.readFileSync(path.join(__dirname,'fixtures.json'))),results=[];
const flags=['-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all'];
function run(command,label,env={}){const r=cp.spawnSync(command[0],command.slice(1),{encoding:'utf8',env:{...process.env,...env}});assert.equal(r.error,undefined);assert.equal(r.signal,null);fs.writeFileSync(path.join(scratch,label+'.stdout'),r.stdout);fs.writeFileSync(path.join(scratch,label+'.stderr'),r.stderr);return {command,exit:r.status,stdout:r.stdout,stderr:r.stderr};}
for(const row of rows){
 const source=fs.readFileSync(path.join(scratch,row.file+'.c.stdout'),'utf8');let mutant,name;
 if(row.proven){
  assert.ok(!source.includes('non-null assertion failed'));
  const anchor='double adamic_local_2_affected = (0x0p+00);';assert.equal(source.split(anchor).length,2);
  mutant=source.replace(anchor,anchor+'\n\t{ static const char message[] = "spurious non-null check"; adamic_panic(message, sizeof message - 1); }');name='insert spurious panic at erased site';
 }else if(!row.fails){
  const guard='if (!(adamic_temporary_3.present))';assert.equal(source.split(guard).length,2);
  mutant=source.replace(guard,'if (adamic_temporary_3.present)');name='invert checked presence test';
 }else{
  const call='adamic_panic((&adamic_string_0)->bytes, (&adamic_string_0)->length);';assert.equal(source.split(call).length,2);
  mutant=source.replace(call,'(void)0;');name='drop checked failure panic';
 }
 const c=path.join(scratch,row.file+'.mutant.c'),binary=path.join(scratch,row.file+'.mutant.native');fs.writeFileSync(c,mutant);
 const build=run(['clang',...flags,'-I',runtime,c,'-Xlinker','--whole-archive',path.join(runtime,'runtime.a'),'-Xlinker','--no-whole-archive','-lm','-o',binary],row.file+'.mutant-build');assert.equal(build.exit,0,build.stderr);
 const actual=run([binary],row.file+'.mutant',{ASAN_OPTIONS:row.fails?'detect_leaks=1':'detect_leaks=0',UBSAN_OPTIONS:'halt_on_error=1'});
 assert.ok(!actual.stderr.includes('Sanitizer'),actual.stderr);
 if(row.fails)assert.deepEqual([actual.exit,actual.stdout,actual.stderr],[0,row.node_stdout,'']);
 else{assert.equal(actual.exit,70);assert.ok(actual.stderr.startsWith('adamic: panic:'));assert.notEqual(actual.stdout,row.node_stdout);}
 results.push({fixture:row.file,mutant:name,caught_by:row.fails?'exact expression/location panic, exit 70, and before-only stdout':'Node stdout/exit golden',build_exit:build.exit,actual});
}
fs.writeFileSync(path.join(__dirname,'mutants.json'),JSON.stringify(results,null,2)+'\n');console.log('3 emitted-C mutants compile; all 3 caught semantically, without sanitizer or C compiler failure');
