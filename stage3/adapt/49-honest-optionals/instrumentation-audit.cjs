const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const root=path.resolve(process.argv[2]),tokens=['__adamic_nonnull_probe','__adamic_slots','__49read','ADAMIC_NONNULL_LOG','ADAMIC_SLOT_LOG','ADAMIC_49_LOG'],files=[];
function visit(dir){for(const e of fs.readdirSync(dir,{withFileTypes:true})){const file=path.join(dir,e.name);if(e.isDirectory())visit(file);else if(e.isFile())files.push(file);}}
visit(path.join(root,'src'));const built=path.join(root,'built/local');for(const n of fs.readdirSync(built).filter(n=>n.endsWith('.js')))files.push(path.join(built,n));
for(const file of files){const text=fs.readFileSync(file,'utf8');for(const token of tokens)assert.ok(!text.includes(token),'instrumentation leaked into '+file+': '+token);}
console.log(JSON.stringify({files_checked:files.length,instrumentation_matches:0,tokens},null,2));
