// Source Node goldens; real backends are checked by TestParserConstructionSource.
const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process');
const directory=__dirname,rows=[];
const expected={
 'factory-unset.a':'true\ntrue\ntrue\nREADY\n',
 'own-key-order.a':'kind\nfalse\n2,10,kind,beta,alpha\n2,10,kind,beta,alpha\ntrue\n{"2":2,"10":10,"kind":1,"beta":21,"alpha":3}\n',
 'node-array-keys.a':'0,1\nfalse\n0,1,end,pos,transformFlags,hasTrailingComma\ntrue\n',
 'node-array-unset.a':'true\ntrue\ntrue\ntrue\nfalse\nready\ntrue\nundefined\ntrue\n-1\n',
 'null-placeholder-pending.a':'true\ntrue\n',
 'node-array-json.a':'[1,2]\n','node-array-length.a':'2\n2\n2\n'
};
function run(text){const module=require('node:module');const code=module.stripTypeScriptTypes(text,{mode:'transform'});return cp.spawnSync(process.execPath,['--input-type=module','-e',code],{encoding:'utf8'});}
for(const name of Object.keys(expected)){const actual=run(fs.readFileSync(path.join(directory,name),'utf8'));if(actual.status!==0||actual.stdout!==expected[name]||actual.stderr)throw Error(JSON.stringify({name,...actual}));rows.push({name,stdout:actual.stdout,exit:actual.status,native:'TestParserConstructionSource and TestParserConstructionNullPlaceholderDelivered; placeholder pending cleared'});}
const use=run(fs.readFileSync(path.join(directory,'factory-use.a'),'utf8'));if(use.status!==1||use.stdout!=='true\n'||!use.stderr.includes('TypeError'))throw Error('unsafe declared-T source use changed');rows.push({name:'factory-use.a',stdout:use.stdout,exit:use.status,stderr:'TypeError observed; Adamic must check the declared-T use at exit 70',native:'TestParserConstructionUseCheck; declared-T consumption pinned at exit 70'});
// Source prediction mutants, deliberately not native runtime mutants.
const mutants=[
 ['synthesized-key','own-key-order.a',"return {kind:1} as FactoryNode;","return {kind:1,alpha:undefined,beta:undefined,'2':undefined,'10':undefined} as unknown as FactoryNode;"],
 ['declaration-order','own-key-order.a',"node.beta=20;","node.alpha=3;node.beta=20;"],
 ['extra-in-json','node-array-json.a','JSON.stringify(nodes)','JSON.stringify({...nodes})']
];
for(const [name,file,old,replacement] of mutants){const source=fs.readFileSync(path.join(directory,file),'utf8');if(source.split(old).length!==2)throw Error('mutation site '+name);const actual=run(source.replace(old,replacement));if(actual.status!==0||actual.stdout===expected[file])throw Error('mutant not caught '+name);rows.push({mutant:name,exit:actual.status,catcher:'original source Node stdout',native:'pending: source prediction mutant only'});}
fs.writeFileSync(path.join(directory,'observations.json'),JSON.stringify(rows,null,2)+'\n');console.log(JSON.stringify(rows,null,2));
