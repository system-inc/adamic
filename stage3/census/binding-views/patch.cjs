const fs=require('node:fs'),cp=require('node:child_process'),assert=require('node:assert/strict'),ts=require('typescript');
const scout='/workspace/scratch/binding-views/scout/stage3/census/scout-ty5myjt/';
function patch(text,name){
 const tmp='/workspace/scratch/binding-views/tmp/patch-'+name;
 const crlf=text.includes('\r\n');
 fs.writeFileSync(tmp,text.replace(/\r\n/g,'\n'));
 cp.execFileSync('patch',['--silent',tmp,scout+name+'.patch']);
 const result=fs.readFileSync(tmp,'utf8');
 return crlf ? result.replace(/\n/g,'\r\n') : result;
}
module.exports={patch};
