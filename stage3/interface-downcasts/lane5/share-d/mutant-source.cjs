// Independently weaken the fixture and its self-recorded carrier together.
// Only the upstream mapped-declaration check can catch this mutation.
const fs=require('fs'),path=require('path'),cp=require('child_process');
const fixture=path.join(__dirname,'rank-2778/good.a'),ledger=path.join(__dirname,'witnesses.json');
const beforeFixture=fs.readFileSync(fixture),beforeLedger=fs.readFileSync(ledger);
try {
 const text=beforeFixture.toString();
 if(text.split('computeValue: (').length!==2) throw Error('unexpected fixture shape');
 fs.writeFileSync(fixture,text.replace('computeValue: (','computeValue?: ('));
 const evidence=JSON.parse(beforeLedger);
 evidence.members.find(m=>m.rank===2778).carriers=evidence.members.find(m=>m.rank===2778).carriers.replace('computeValue:','computeValue?:');
 fs.writeFileSync(ledger,JSON.stringify(evidence,null,2)+'\n');
 const result=cp.spawnSync(process.execPath,[path.join(__dirname,'verify.cjs'),...process.argv.slice(2)],{encoding:'utf8'});
 process.stdout.write(result.stdout||'');process.stderr.write(result.stderr||'');
 if(result.status===0 || !(result.stderr||'').includes('original mapped declaration changed 2778')) throw Error('source mutant escaped its intended catcher');
 console.log('Source mutant caught by original mapped declaration check; verifier exit '+result.status);
} finally {
 fs.writeFileSync(fixture,beforeFixture);fs.writeFileSync(ledger,beforeLedger);
}
