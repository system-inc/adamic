'use strict';
const fs=require('fs'),path=require('path'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const tree=process.argv[2],file=path.join(tree,'src/compiler/core.ts'),text=fs.readFileSync(file,'utf8'),s=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
const owners=s.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='levenshteinWithMax');if(owners.length!==1)throw Error('owner count');
const n=owners[0],before=n.getText(s),filled=before.replaceAll('new Array(s2.length + 1)','new Array<number>(s2.length + 1).fill(0)');
let after=filled; let reads=0; for(const expression of ['previous[j - 1]','previous[j]','current[j - 1]','previous[s2.length]']){reads+=after.split(expression).length-1;after=after.replaceAll(expression,expression+'!');} if(reads!==6)throw Error('want six checked reads');
if(before===after||after.split('.fill(0)').length!==3)throw Error('initializer count');
fs.writeFileSync('/workspace/scratch/levenshtein/measured-owners.json',JSON.stringify({before,after},null,2)+'\n');
const original=before.replace('function levenshteinWithMax(', 'function levenshteinOriginal('),candidate=after.replace('function levenshteinWithMax(', 'function levenshteinCandidate(');
const wrapper=`function levenshteinWithMax(s1: string, s2: string, max: number): number | undefined {
 const a = levenshteinOriginal(s1, s2, max), b = levenshteinCandidate(s1, s2, max);
 (globalThis as any).__levenshteinRecord(s1, s2, max, a, b);
 if (!Object.is(a, b)) throw new Error("levenshtein counterexample");
 return a;
}`;
fs.writeFileSync(file,text.slice(0,n.getStart(s))+original+'\n'+candidate+'\n'+wrapper+text.slice(n.end));
console.log(JSON.stringify({owner:before,replacements:2}));
