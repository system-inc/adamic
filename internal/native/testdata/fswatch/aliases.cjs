const fs=require('node:fs'),os=require('node:os'),path=require('node:path');
const directory=fs.mkdtempSync(path.join(os.tmpdir(),'fs-watch-shared-alias-'));
const a=path.join(directory,'a'),b=path.join(directory,'b');fs.writeFileSync(a,'');fs.linkSync(a,b);
let seen=0;function changed(event,name){console.log(event+' '+name);if(++seen===2){first.close();second.close();fs.rmSync(directory,{recursive:true,force:true});}}
const first=fs.watch(a,changed),second=fs.watch(b,changed);fs.appendFileSync(b,'x');
