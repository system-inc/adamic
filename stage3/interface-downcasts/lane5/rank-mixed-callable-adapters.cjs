// Candidate syntax inventory only. Producer signatures and view reachability are not known.
const fs = require('fs');
const path = require('path');
const ts = require(process.argv[2]);
const directory = __dirname;
function mixedSlots(signature) {
  const source = ts.createSourceFile('contract.ts', `type Contract = ${signature};`, ts.ScriptTarget.Latest, true);
  const slots = [];
  function inspect(type, position) {
    while (type && ts.isParenthesizedTypeNode(type)) type = type.type;
    if (!type || !ts.isUnionTypeNode(type)) return;
    let scalar = false, boxed = false;
    for (const member of type.types) {
      const text = member.getText(source);
      scalar ||= /^(number|boolean|true|false|[0-9]+)$/.test(text);
      // String has a heap representation in native. Named aliases are deliberately unresolved.
      boxed ||= text === 'string' || ts.isTypeLiteralNode(member) || ts.isArrayTypeNode(member)
        || ts.isFunctionTypeNode(member) || (ts.isLiteralTypeNode(member) && ts.isStringLiteral(member.literal));
    }
    if (scalar && boxed) slots.push({position, union: type.getText(source)});
  }
  function visit(node) {
    if (ts.isFunctionTypeNode(node) || ts.isCallSignatureDeclaration(node) || ts.isMethodSignature(node)) {
      inspect(node.type, 'result');
      for (const parameter of node.parameters) inspect(parameter.type, `parameter ${parameter.name.getText(source)}`);
    }
    ts.forEachChild(node, visit);
  }
  visit(source);
  return slots;
}
const inventories = [
  ['static', 'callable-pairs-ranked.json', 'read_count'],
  ['unknown_fallback', 'unknown-callable-pairs-ranked.json', 'reads'],
];
const result = {method: 'Explicit scalar/boxed unions in callable slots, including nested callback slots; aliases and producer-side variance unresolved; candidate counts only', inventories: {}};
for (const [name, file, count] of inventories) {
  const rows = JSON.parse(fs.readFileSync(path.join(directory, file), 'utf8')).flatMap(pair => {
    const slots = [...new Set((pair.declared_types || [pair.declared_type]).flatMap(mixedSlots).map(JSON.stringify))].map(JSON.parse);
    return slots.length ? [{rank: pair.rank, type: pair.type, field: pair.field, reads: pair[count], slots, witness: pair.witness || pair.sites[0]}] : [];
  });
  result.inventories[name] = {candidate_pairs: rows.length, candidate_reads: rows.reduce((n, row) => n + row.reads, 0), rows};
}
fs.writeFileSync(path.join(directory, 'mixed-callable-adapters-ranked.json'), JSON.stringify(result, null, 2) + '\n');
for (const [name, inventory] of Object.entries(result.inventories)) {
  console.log(`${name}: ${inventory.candidate_pairs} candidate pairs / ${inventory.candidate_reads} candidate reads`);
  for (const row of inventory.rows) console.log(`${row.rank}: ${row.reads} ${row.type}.${row.field}`);
}
