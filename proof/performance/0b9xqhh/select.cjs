"use strict";
// Usage: node select.cjs <tsTree> <sites.tsv> <outDir>
// Finds each site's expression in <tsTree>'s src/compiler program, resolves the member names the census listed on
// the site's target type to their declarations, and writes sites.csv and members.json into <outDir>.
const fs = require("node:fs");
const path = require("node:path");
const common = require("./common.cjs");

const [tree, sitesPath, outDir] = process.argv.slice(2).map(a => a && path.resolve(a));
if (!tree || !sitesPath || !outDir) {
    console.error("usage: node select.cjs <tsTree> <sites.tsv> <outDir>");
    process.exit(2);
}

// A row whose census list is empty, with the members the ruling's author named for it. nodeFactory.ts:7489's target
// is (SourceMapRange | undefined)[], so its view is SourceMapRange's, read through the array's element.
const emptyListMembers = {
    "src/compiler/factory/nodeFactory.ts:7489": { names: ["source"], elementOfArray: true },
};

const ts = common.loadTypeScript();
console.log(`typescript API ${ts.version}`);
const { program, checker } = common.loadProgram(ts, tree);
const tsv = common.readTsv(sitesPath);
console.log(`sites: ${tsv.rows.length} rows, columns ${JSON.stringify(tsv.header)}`);

const siteRows = [];
const savedSites = [];
const declarations = {};
const memberOrder = [];
let skipped = 0;
let unmatched = 0;
let unresolvedNames = 0;

for (const row of tsv.rows) {
    const [file, lineText, columnText, relation, censusSource, censusTarget, fieldsText = "", cleared = ""] = row;
    const site = `${file}:${lineText}:${columnText}`;
    // A non-empty last column names the wave that already cleared the site, so it no longer crosses as a view.
    if (cleared.trim()) {
        skipped++;
        continue;
    }
    const record = { site, relation, census_source: censusSource, census_target: censusTarget };
    siteRows.push(record);
    let names = fieldsText.split(",").map(s => s.trim()).filter(Boolean);
    let elementOfArray = false;
    if (!names.length) {
        const fallback = emptyListMembers[`${file}:${lineText}`];
        if (!fallback) {
            record.status = "no member list";
            unmatched++;
            continue;
        }
        names = fallback.names;
        elementOfArray = fallback.elementOfArray;
        record.note = `empty list: ${names.join(",")} on the array's element`;
    }
    const sourceFile = program.getSourceFile(path.join(tree, file));
    if (!sourceFile) {
        record.status = "unmatched: file not in the program";
        unmatched++;
        continue;
    }
    let expression;
    let chosen;
    for (const candidate of common.candidatePositions(sourceFile, Number(lineText), Number(columnText))) {
        expression = common.relationExpression(ts, sourceFile, candidate.position);
        if (expression) {
            chosen = candidate;
            break;
        }
    }
    if (!expression) {
        record.status = "unmatched: no expression starts there";
        unmatched++;
        continue;
    }
    record.unit = chosen.unit;
    record.expression = `${ts.SyntaxKind[expression.kind]} ${expression.getText(sourceFile).replace(/\s+/g, " ").slice(0, 60)}`;
    record.source = checker.typeToString(checker.getTypeAtLocation(expression), undefined, ts.TypeFormatFlags.NoTruncation);
    const target = common.siteTarget(ts, checker, expression, elementOfArray);
    if (!target) {
        record.status = "unmatched: no contextual type";
        unmatched++;
        continue;
    }
    record.target = checker.typeToString(target, undefined, ts.TypeFormatFlags.NoTruncation);
    const memberTexts = [];
    const missing = [];
    for (const name of names) {
        const resolved = common.resolveName(ts, checker, target, name);
        if (!resolved.length) {
            missing.push(name);
            unresolvedNames++;
            continue;
        }
        const ids = [];
        for (const r of resolved) {
            for (const { id, declaration } of common.memberIds(ts, checker, tree, r, name)) {
                ids.push(id);
                if (declarations[id]) {
                    declarations[id].sites.push(site);
                    continue;
                }
                const entry = { name, optional: (r.symbol.flags & ts.SymbolFlags.Optional) !== 0, sites: [site] };
                if (declaration) {
                    const declarationFile = declaration.getSourceFile();
                    const where = declarationFile.getLineAndCharacterOfPosition(declaration.getStart(declarationFile));
                    Object.assign(entry, { file: path.relative(tree, declarationFile.fileName), line: where.line + 1, column: where.character + 1 });
                }
                else {
                    entry.owner = checker.typeToString(r.owner, undefined, ts.TypeFormatFlags.NoTruncation);
                }
                declarations[id] = entry;
                memberOrder.push(id);
            }
        }
        memberTexts.push(`${name}: ${[...new Set(ids)].join(" | ")}`);
    }
    record.members = memberTexts.join("; ");
    record.unresolved = missing.join(",");
    record.status = missing.length ? "matched, some names unresolved" : "matched";
    savedSites.push({ site, file, position: expression.getStart(sourceFile), names, elementOfArray });
}

fs.mkdirSync(outDir, { recursive: true });
fs.writeFileSync(path.join(outDir, "sites.csv"), common.csv(siteRows, ["site", "relation", "status", "unit", "expression", "source", "target", "census_source", "census_target", "members", "unresolved", "note"]));
fs.writeFileSync(path.join(outDir, "members.json"), JSON.stringify({ tree, members: memberOrder, declarations, sites: savedSites }, null, 1) + "\n");
console.log(`sites: ${tsv.rows.length} rows, ${skipped} already cleared, ${siteRows.length} taken, ${unmatched} unmatched, ${unresolvedNames} names unresolved`);
console.log(`members: ${memberOrder.length} declarations`);
for (const id of memberOrder) console.log(`  ${id}${declarations[id].optional ? "" : " (required)"} from ${declarations[id].sites.length} sites`);
for (const r of siteRows) if (r.status !== "matched") console.log(`  site ${r.site}: ${r.status}${r.unresolved ? ` (${r.unresolved})` : ""}`);
if (!memberOrder.length) {
    console.error("no member resolved");
    process.exit(1);
}
