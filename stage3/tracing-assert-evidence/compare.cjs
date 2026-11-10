'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
const [beforeTree,afterTree,out]=process.argv.slice(2);
const beforeDir=path.join(beforeTree,'built/local'),afterDir=path.join(afterTree,'built/local');
const names=fs.readdirSync(beforeDir).filter(x=>x.endsWith('.js')).sort();assert.deepEqual(fs.readdirSync(afterDir).filter(x=>x.endsWith('.js')).sort(),names);
function tokens(text){const scanner=ts.createScanner(ts.ScriptTarget.Latest,true,ts.LanguageVariant.Standard,text);const result=[];for(let k=scanner.scan();k!==ts.SyntaxKind.EndOfFileToken;k=scanner.scan()){let value=scanner.getTokenText();if(k===ts.SyntaxKind.Identifier)value=value.replace(/\d+$/,'');result.push([k,value]);}return result;}
const report=[];
for(const name of names){const before=fs.readFileSync(path.join(beforeDir,name),'utf8'),after=fs.readFileSync(path.join(afterDir,name),'utf8');
 if(before===after){report.push({name,classification:'byte-identical'});continue;}
 const re=/(Debug\d*\.assert\()(tracing\d*)(, "Tracing is not in progress"\))/g;let n=0;const expected=before.replace(re,(_,a,b,c)=>{n++;return a+'!!'+b+c;});assert(n<=1,name);
 assert.deepEqual(tokens(expected),tokens(after),'unreviewed emitted difference in '+name);
 report.push({name,assertionEdits:n,classification:n?'stopTracing argument plus any normalized esbuild identifiers':'pure esbuild identifier rename',normalization:'strip numeric suffix from identifier tokens only'});
}
const b=require(path.resolve(beforeDir,'typescript.js')),a=require(path.resolve(afterDir,'typescript.js'));const bk=Object.keys(b).sort(),ak=Object.keys(a).sort();assert.deepEqual(ak,bk);
fs.writeFileSync(out,JSON.stringify({files:report,exportKeys:bk,exportKeyCount:bk.length,exportKeysIdentical:true},null,2)+'\n');console.log(JSON.stringify(report));
