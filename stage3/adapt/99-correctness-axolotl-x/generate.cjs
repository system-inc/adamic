'use strict';
// Generate reviewable owner pins from a fully adapted main tree, never during apply.
const e = require('./engine.cjs');
const {fs, path, assert} = e;
const tree = path.resolve(process.argv[2]);
const ownership = require('./ownership.json');
const handoff = site => ownership.filter(row => row.file===site.file && (site.owner===row.owner || site.owner.startsWith(row.owner+'.'))).sort((a,b)=>b.owner.length-a.owner.length)[0];
const normalized = new Map(e.sources(tree).map(name => [name, e.normalize(fs.readFileSync(name,'utf8'))]));
const inventory = e.scan(tree,e.program(tree,normalized));
assert.equal(inventory.sites.length, 46);
assert.equal(inventory.refused.length, 0);
inventory.left = inventory.sites.filter(handoff).map(site=>({...site, ...{pinnedOwner:handoff(site).owner, adapter:handoff(site).adapter, reason:'owner pinned by '+handoff(site).adapter+'; folded there'}}));
inventory.sites = inventory.sites.filter(site=>!handoff(site));
assert.equal(inventory.left.length,18);
assert.equal(inventory.sites.length,28);
if (process.argv[3]) {
    const pristine = path.resolve(process.argv[3]);
    const probe = e.scan(pristine);
    assert.equal(probe.sites.length,46);
    inventory.origins = probe.sites.map(({file,owner,line,column,before,after})=>({file,owner,line,column,before,after}));
}
const owners = [], imports = [], files = [...new Set(inventory.sites.map(site => site.file))];
for (const file of files) {
    const before = e.normalize(fs.readFileSync(path.join(tree,'src/compiler',file),'utf8'));
    const sites = inventory.sites.filter(site => site.file === file);
    let after = e.edit(before, sites);
    const removal = e.removableImport(file, after);
    if (removal) { after = e.edit(after,[removal]); imports.push({file, ...removal}); }
    const beforeOwners = e.ownerMap(e.parse(file,before));
    const afterOwners = e.ownerMap(e.parse(file,after));
    for (const owner of new Set(sites.map(site => site.owner))) {
        const b = beforeOwners.get(owner).text, a = afterOwners.get(owner).text;
        owners.push({file, owner, before: b, after: a, ownerBeforeSha256: e.digest(b), ownerAfterSha256: e.digest(a)});
    }
}
fs.writeFileSync(path.join(__dirname,'owners.json'), JSON.stringify(owners,null,2)+'\n');
fs.writeFileSync(path.join(__dirname,'inventory.json'), JSON.stringify({...inventory,imports},null,2)+'\n');
console.log(JSON.stringify({owners:owners.length, sites:inventory.sites.length, refused:inventory.refused.length, left:inventory.left.length, imports:imports.map(x=>x.file)},null,2));
