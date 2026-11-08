const fs=require('fs'),path=require('path'),crypto=require('crypto'),cp=require('child_process');
const [stock,adapted]=process.argv.slice(2),repo=path.resolve(__dirname,'../../..');
const git=(cwd,...args)=>cp.execFileSync('git',args,{cwd,encoding:'utf8'}).trim();
const digest=data=>crypto.createHash('sha256').update(data).digest('hex');
function files(root,dir){let result=[];for(const name of fs.readdirSync(path.join(root,dir)).sort()){const relative=path.join(dir,name),stat=fs.statSync(path.join(root,relative));if(stat.isDirectory())result.push(...files(root,relative));else result.push({file:relative.replaceAll('\\','/'),sha256:digest(fs.readFileSync(path.join(root,relative)))});}return result;}
const inputs={};for(const [name,root]of [['stock',stock],['adapted',adapted]]){const list=files(root,'src/compiler');inputs[name]={root,compilerFiles:list,aggregateSha256:digest(JSON.stringify(list)),compilerConfig:JSON.parse(fs.readFileSync(path.join(root,'src/compiler/tsconfig.json'))),baseConfig:JSON.parse(fs.readFileSync(path.join(root,'src/tsconfig-base.json'))),trackedChanges:git(root,'status','--short','--untracked-files=no')};}
const api=path.join(repo,'stage3/api/node_modules/typescript/lib/typescript.js');
const result={recorded:new Date().toISOString(),main:git(repo,'rev-parse','origin/main'),stock:git(stock,'rev-parse','HEAD'),adaptedProductKey:cp.execFileSync('bash',['stage3/apply.sh','--key'],{cwd:repo,encoding:'utf8'}).trim(),node:process.version,platform:process.platform,architecture:process.arch,typescriptApi:{version:require(api).version,sha256:digest(fs.readFileSync(api))},stockNodeTypes:require(path.join(stock,'node_modules/@types/node/package.json')).version,stockEsbuild:require(path.join(stock,'node_modules/esbuild/package.json')).version,inputs};
if(result.stock!=='050880ce59e30b356b686bd3144efe24f875ebc8')throw Error('Wrong stock pin');
fs.writeFileSync(path.join(__dirname,'provenance.json'),JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify({main:result.main,stock:result.stock,adaptedProductKey:result.adaptedProductKey}));
