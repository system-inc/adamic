'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const [before,after]=process.argv.slice(2),sha=x=>crypto.createHash('sha256').update(x).digest('hex'),changes=[],same=[];
const names=fs.readdirSync(path.join(before,'built/local')).filter(n=>n.endsWith('.js')||n.endsWith('.d.ts'));
for(const name of names){const x=fs.readFileSync(path.join(before,'built/local',name),'utf8'),y=fs.readFileSync(path.join(after,'built/local',name),'utf8');if(x!==y){assert(name.endsWith('.js'),name+' declaration changed');const count=x.split('mappingCharCodes.length = 0;').length-1;assert(count>0,name);assert.equal(y,x.replaceAll('mappingCharCodes.length = 0;','mappingCharCodes.splice(0);'),name);changes.push({name,changedClearStatements:count,before:sha(x),after:sha(y)});}else same.push(name);}
assert(changes.length>0);console.log(JSON.stringify({compared:names.length,changes,identical:same},null,2));
