// tsc's insertStatementsAfterPrologue (src/compiler/utilities.ts) writes to.splice(statementIndex, 0, ...from),
// where to is T[], from is readonly T[] and T is a Statement, an object (#p1qw4kf). The body is tsc's, with the
// prologue read narrowed, since Adamic checks with noUncheckedIndexedAccess.
interface Statement {
	readonly kind: number;
	readonly text: string;
}

function insertAfterPrologue<T extends { readonly kind: number }>(to: T[], from: readonly T[] | undefined, isPrologue: (node: T) => boolean): T[] {
	if (from === undefined || from.length === 0) {
		return to;
	}
	let statementIndex = 0;
	for (; statementIndex < to.length; ++statementIndex) {
		const node = to[statementIndex];
		if (node === undefined || !isPrologue(node)) {
			break;
		}
	}
	to.splice(statementIndex, 0, ...from);
	return to;
}

function statement(kind: number, text: string): Statement {
	return { kind, text };
}

function show(label: string, statements: readonly Statement[]): void {
	console.log(`${label}: ${statements.map((node) => `${node.kind}:${node.text}`).join(' ')}`);
}

const isDirective = (node: Statement): boolean => node.kind === 1;
const body = [statement(1, 'use strict'), statement(1, 'use asm'), statement(2, 'call'), statement(3, 'return')];
const added = [statement(4, 'first'), statement(5, 'second')];
const result = insertAfterPrologue(body, added, isDirective);
show('after the prologue', result);
console.log(String(result === body));
added.push(statement(6, 'later'));
show('the array is spread, not kept', body);
show('no prologue', insertAfterPrologue([statement(2, 'call')], [statement(7, 'helper')], isDirective));
show('all prologue', insertAfterPrologue([statement(1, 'use strict')], [statement(8, 'tail')], isDirective));
show('empty target', insertAfterPrologue([], [statement(9, 'only')], isDirective));
show('nothing to add', insertAfterPrologue([statement(1, 'use strict')], [], isDirective));
show('undefined to add', insertAfterPrologue([statement(2, 'call')], undefined, isDirective));
