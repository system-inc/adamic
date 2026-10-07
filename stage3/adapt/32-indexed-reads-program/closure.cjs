"use strict";
const sites = [...require("./closure-sites.json"), ...require("./public-host-sites.json")];
const restored = sites.filter(s => s.kind === "restored-public-method");
const restoredKeys = ["CompilerHost.getDefaultLibLocation", "CompilerHost.createHash", "CompilerHost.readDirectory", "ModuleResolutionHost.trace", "ModuleResolutionHost.directoryExists", "ModuleResolutionHost.getDirectories", "ModuleResolutionHost.realpath", "ProgramHost.createHash", "ProgramHost.realpath", "ProgramHost.getEnvironmentVariable"];
if (JSON.stringify(restored.map(s => s.interface + "." + s.name).sort()) !== JSON.stringify(restoredKeys.sort())) throw new Error("ten restored public owner census drift");
function plan(ts, file, text, check) {
    const sf = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    if (sf.parseDiagnostics.length) throw new Error(`closure parse failure: ${file}`);
    const edits = [];
    for (const site of sites.filter(s => s.file === file)) {
        if (site.kind === "restored-public-method") {
            const found = [];
            function visit(n) {
                if ((ts.isMethodSignature(n) || ts.isPropertySignature(n)) && n.name.getText(sf) === site.name && ts.isInterfaceDeclaration(n.parent) && n.parent.name.text === site.interface) found.push(n);
                ts.forEachChild(n, visit);
            }
            visit(sf);
            if (found.length !== 1 || !found[0].questionToken) throw new Error("restored method owner drift");
            const n = found[0], actual = n.getText(sf);
            if (actual === site.after && ts.isPropertySignature(n)) {
                if (check) throw new Error("public method restoration missing");
                edits.push({at:n.getStart(sf), end:n.end, text:site.before});
            }
            else if (actual !== site.before || !ts.isMethodSignature(n)) throw new Error("restored method signature drift");
            continue;
        }
        if (site.kind === "optional-helper-call") {
            const found=[];
            function visit(n) {
                if(ts.isCallExpression(n)) {
                    const callee=n.expression;
                    const base=ts.isParenthesizedExpression(callee)&&ts.isAsExpression(callee.expression)?callee.expression.expression:callee;
                    let owner=n.parent;
                    while(owner&&!(ts.isFunctionDeclaration(owner)&&owner.name))owner=owner.parent;
                    if(base.getText(sf)===site.call&&owner?.name.text===site.function)found.push(n);
                }
                ts.forEachChild(n,visit);
            }
            visit(sf);
            if(found.length!==site.total)throw new Error("optional helper call count drift");
            const n=found[site.occurrence-1],before=site.call,after="("+site.call+" as "+site.type+")";
            if(JSON.stringify(n.arguments.map(a=>a.getText(sf)))!==JSON.stringify(site.arguments))throw new Error("optional helper call argument drift");
            if(n.expression.getText(sf)===before) {
                if(check)throw new Error("optional helper call view missing");
                edits.push({at:n.expression.getStart(sf),end:n.expression.end,text:after});
            }
            else if(n.expression.getText(sf)!==after)throw new Error("optional helper call view drift");
            continue;
        }
        if (site.kind === "internal-property") {
            const found = [];
            function visit(n) {
                if (ts.isPropertySignature(n) && n.name.getText(sf) === site.name && ts.isInterfaceDeclaration(n.parent) && n.parent.name.text === site.interface) {
                    const block = n.parent.parent;
                    if (ts.isModuleBlock(block) && ts.isModuleDeclaration(block.parent) && block.parent.name.getText(sf) === site.namespace && !n.parent.modifiers?.some(m => m.kind === ts.SyntaxKind.ExportKeyword)) found.push(n);
                }
                ts.forEachChild(n, visit);
            }
            visit(sf);
            if (found.length !== 1 || !found[0].questionToken || !found[0].type) throw new Error("private optional property owner drift");
            const n = found[0].type;
            if (n.getText(sf) === site.before) {
                if (check) throw new Error("private optional property undefined missing");
                edits.push({at:n.getStart(sf), end:n.end, text:site.after});
            }
            else if (n.getText(sf) !== site.after) throw new Error("private optional property type drift");
            continue;
        }
        if (["internal-method", "public-method", "call-receiver", "rest-type"].includes(site.kind)) {
            const found = [];
            function visit(n) {
                if (["internal-method", "public-method"].includes(site.kind) && (ts.isMethodSignature(n) || ts.isPropertySignature(n)) && n.name.getText(sf) === site.name && ts.isInterfaceDeclaration(n.parent) && n.parent.name.text === site.interface) found.push(n);
                if (site.kind === "call-receiver" && ts.isCallExpression(n) && ts.isPropertyAccessExpression(n.expression) && n.expression.name.text === "call") {
                    let owner = n.parent;
                    while (owner && !(ts.isFunctionDeclaration(owner) && owner.name)) owner = owner.parent;
                    let receiver = n.expression.expression;
                    const base = ts.isParenthesizedExpression(receiver) && ts.isAsExpression(receiver.expression) ? receiver.expression.expression : receiver;
                    if (owner?.name.text === site.function && base.getText(sf) === site.expression) found.push(receiver);
                }
                if (site.kind === "rest-type" && ts.isParameter(n) && ts.isArrowFunction(n.parent) && n.dotDotDotToken && n.name.getText(sf) === site.name) {
                    let owner = n.parent;
                    while (owner && !(ts.isFunctionDeclaration(owner) && owner.name)) owner = owner.parent;
                    if (owner?.name.text === site.function) found.push(n.type);
                }
                ts.forEachChild(n, visit);
            }
            visit(sf);
            if (found.length !== 1 || !found[0]) throw new Error("watch contract occurrence drift: " + site.kind);
            const n = found[0], actual = n.getText(sf);
            const before = site.kind === "call-receiver" ? site.expression : site.before;
            const after = site.kind === "call-receiver" ? "(" + site.expression + " as " + site.type + ")" : site.after;
            if (actual === before) {
                if (check) throw new Error("watch contract missing: " + site.kind);
                edits.push({at:n.getStart(sf), end:n.end, text:after});
            }
            else if (actual !== after) throw new Error("watch contract shape drift: " + site.kind);
            continue;
        }
        if (site.kind === "nonempty-initializer") {
            const found = [];
            function visit(n) {
                if (ts.isVariableDeclaration(n) && n.name.getText(sf) === site.name) {
                    let owner = n.parent;
                    while (owner && !(ts.isFunctionDeclaration(owner) && owner.name)) owner = owner.parent;
                    if (owner?.name.text === site.function) found.push(n);
                }
                ts.forEachChild(n, visit);
            }
            visit(sf);
            if (found.length !== 1 || !found[0].initializer) throw new Error("nonempty local owner drift");
            const n = found[0].initializer;
            if (n.getText(sf) === site.expression) {
                if (check) throw new Error("nonempty local contract missing");
                edits.push({at:n.end, text:" as " + site.after});
            }
            else if (!ts.isAsExpression(n) || n.expression.getText(sf) !== site.expression || n.type.getText(sf) !== site.after) throw new Error("nonempty local type drift");
            continue;
        }
        if (site.kind === "nonempty-parameter") {
            const found = [];
            function visit(n) {
                if (site.kind === "nonempty-parameter" && ts.isParameter(n) && n.name.getText(sf) === site.name && ts.isFunctionDeclaration(n.parent) && n.parent.name?.text === site.function) found.push(n);
                ts.forEachChild(n, visit);
            }
            visit(sf);
            if (found.length !== 1 || !found[0].type) throw new Error("nonempty declaration owner drift");
            const n = found[0].type, actual = n.getText(sf);
            if (actual === site.type) {
                if (check) throw new Error("nonempty tuple contract missing");
                edits.push({at:n.getStart(sf), end:n.end, text:site.after});
            }
            else if (actual !== site.after) throw new Error("nonempty declaration type drift");
            continue;
        }
        if (site.kind === "optional-tuple") {
            const owners = sf.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === site.function);
            if (owners.length !== 1 || !owners[0].type || !ts.isTypeOperatorNode(owners[0].type) || owners[0].type.operator !== ts.SyntaxKind.ReadonlyKeyword || !ts.isTupleTypeNode(owners[0].type.type)) throw new Error("closure tuple owner drift");
            const tuple = owners[0].type.type;
            if (tuple.elements.length !== 5) throw new Error("closure tuple length drift");
            const members = tuple.elements.filter(n => ts.isNamedTupleMember(n) && n.name.getText(sf) === site.name);
            if (members.length !== 1 || !members[0].questionToken) throw new Error("closure tuple member drift");
            const n = members[0].type, actual = n.getText(sf), wanted = site.type + " | undefined";
            if (actual === site.type) {
                if (check) throw new Error("closure tuple undefined missing");
                edits.push({at:n.end, text:" | undefined"});
            }
            else if (actual !== wanted) throw new Error("closure tuple type drift");
            continue;
        }
        const matches = [];
        function visit(n) {
            if (ts.isCallExpression(n) && n.expression.getText(sf) === site.call) {
                let owner = n.parent;
                while (owner && !(ts.isFunctionDeclaration(owner) && owner.name)) owner = owner.parent;
                if (owner && owner.name.text === site.function) matches.push(n);
            }
            ts.forEachChild(n, visit);
        }
        visit(sf);
        if (matches.length !== 1) throw new Error(`closure occurrence drift: ${file}:${site.function}:${site.call}`);
        const n = matches[0];
        if (JSON.stringify(n.arguments.map(a => a.getText(sf))) !== JSON.stringify(site.arguments)) throw new Error(`closure argument drift: ${file}:${site.function}`);
        if (n.typeArguments) {
            if (JSON.stringify(n.typeArguments.map(t => t.getText(sf))) !== JSON.stringify(site.typeArguments)) throw new Error(`closure type drift: ${file}:${site.function}`);
        }
        else {
            if (check) throw new Error(`closure types missing: ${file}:${site.function}`);
            edits.push({at: n.expression.end, text: `<${site.typeArguments.join(", ")}>`});
        }
    }
    for (const e of edits.sort((a,b) => b.at-a.at)) text = text.slice(0,e.at)+e.text+text.slice(e.end === undefined ? e.at : e.end);
    return {text, contracts: edits.length};
}
module.exports = {plan};
