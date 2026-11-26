// #0d56xpd question 3: every ?. link in the adapted tree's src/compiler whose receiver tsc types as never
// undefined or null, and what tsc types the link's result as. Usage: node siblings.cjs <typescript> <adapted tree>
'use strict';
const ts = require(process.argv[2]);
const tree = process.argv[3].replace(/\/$/, '');
const configPath = tree + '/src/compiler/tsconfig.json';
const host = { ...ts.sys, onUnRecoverableConfigFileDiagnostic: (diagnostic) => { throw new Error(ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n')); } };
const config = ts.getParsedCommandLineOfConfigFile(configPath, {}, host);
console.log('SIB config files', config.fileNames.length, 'strictNullChecks', config.options.strictNullChecks, 'strict', config.options.strict);
const program = ts.createProgram({ rootNames: config.fileNames, options: config.options, projectReferences: config.projectReferences });
const checker = program.getTypeChecker();
const nullish = ts.TypeFlags.Any | ts.TypeFlags.Unknown | ts.TypeFlags.Undefined | ts.TypeFlags.Null | ts.TypeFlags.Void;
function nullable(type, depth = 0) {
	if (depth > 8) return true;
	if (type.flags & nullish) return true;
	if (type.isUnion()) return type.types.some((member) => nullable(member, depth + 1));
	if (type.flags & (ts.TypeFlags.TypeParameter | ts.TypeFlags.IndexedAccess | ts.TypeFlags.Conditional)) {
		const constraint = checker.getBaseConstraintOfType(type);
		return !constraint || constraint === type || nullable(constraint, depth + 1);
	}
	return false;
}
function every(type, test) {
	return type.isUnion() ? type.types.every(test) : test(type);
}
function kind(type) {
	if (nullable(type)) return 'maybe';
	if (every(type, (member) => (member.flags & ts.TypeFlags.NumberLike) !== 0)) return 'number';
	if (every(type, (member) => (member.flags & ts.TypeFlags.BooleanLike) !== 0)) return 'boolean';
	if (every(type, (member) => (member.flags & ts.TypeFlags.StringLike) !== 0)) return 'string';
	return 'other';
}
const counts = {};
for (const file of program.getSourceFiles()) {
	if (file.isDeclarationFile || !file.fileName.startsWith(tree + '/src/compiler/')) continue;
	const visit = (node) => {
		const link = (ts.isPropertyAccessExpression(node) || ts.isElementAccessExpression(node) || ts.isCallExpression(node)) && node.questionDotToken;
		if (link) {
			const receiver = checker.getTypeAtLocation(node.expression);
			if (!nullable(receiver)) {
				const result = checker.getTypeAtLocation(node);
				const syntax = ts.isPropertyAccessExpression(node) ? 'property' : ts.isElementAccessExpression(node) ? 'element' : 'call';
				const resultKind = kind(result);
				const key = syntax + ' ' + resultKind;
				counts[key] = (counts[key] || 0) + 1;
				const position = file.getLineAndCharacterOfPosition(node.getStart(file));
				const where = file.fileName.slice(tree.length + 1) + ':' + (position.line + 1) + ':' + (position.character + 1);
				let unit = node.parent;
				while (unit && !ts.isFunctionLike(unit)) unit = unit.parent;
				const unitName = unit && unit.name ? unit.name.getText(file) : '(anonymous or top level)';
				console.log(['SIB', syntax, resultKind, where, unitName, checker.typeToString(receiver).slice(0, 60), checker.typeToString(result).slice(0, 60), node.getText(file).replace(/\s+/g, ' ').slice(0, 100)].join('\t'));
			}
		}
		ts.forEachChild(node, visit);
	};
	visit(file);
}
for (const key of Object.keys(counts).sort()) console.log('SIBCOUNT', key, counts[key]);
