'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const ts = require('typescript');
const [filename, scoutFile] = process.argv.slice(2);
const source = ts.createSourceFile(filename, fs.readFileSync(filename, 'utf8'), ts.ScriptTarget.Latest, true);
const sites = JSON.parse(fs.readFileSync(scoutFile, 'utf8'));
const matches = new Map(sites.map(site => [site.id, []]));
const normalize = text => text.replace(/\s+/g, ' ');

function walk(node, owningFunction, statement) {
    if (ts.isFunctionLike(node)) owningFunction = node;
    if (ts.isStatement(node)) statement = node;
    if (ts.isExpressionNode(node) && owningFunction && statement) {
        const functionName = owningFunction.name?.getText(source) ?? '<arrow/anonymous>';
        const expressionText = normalize(node.getText(source));
        for (const site of sites) {
            if (site.function.name !== functionName || normalize(site.expression) !== expressionText) continue;
            const {line, character} = source.getLineAndCharacterOfPosition(node.getStart(source));
            matches.get(site.id).push({
                id: site.id, group: site.group, scout: site.where,
                current: `src/compiler/checker.ts:${line + 1}:${character + 1}`,
                relation: site.relation_position, function: functionName, expression: expressionText,
                statementMatches: normalize(statement.getText(source)) === normalize(site.statement.text),
            });
        }
    }
    ts.forEachChild(node, child => { walk(child, owningFunction, statement); });
}
walk(source, undefined, undefined);
const mapped = [];
for (const site of sites) {
    const candidates = matches.get(site.id);
    const exact = candidates.filter(candidate => candidate.statementMatches);
    const column = site.where.split(":").at(-1);
    const matchingColumn = exact.filter(candidate => candidate.current.split(":").at(-1) === column);
    const best = matchingColumn.length ? matchingColumn : exact.length ? exact : candidates;
    assert.equal(best.length, 1, `${site.id} mappings: ${JSON.stringify(best)}`);
    mapped.push(best[0]);
}
console.log(JSON.stringify(mapped, null, 2));
