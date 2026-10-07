import { syntaxKinds as listenerKinds } from './listeners.a';
export const syntaxKinds: readonly number[] = listenerKinds;
import type { RuleContext } from '../../context.ts';
import { message } from './messages.a';
const name = 'base/consistency-no-bare-throw';
const constructors: readonly string[] = ['Error', 'TypeError', 'RangeError', 'SyntaxError', 'ReferenceError', 'EvalError', 'URIError'];
const captures: readonly string[] = ['/BaseError.ts', '/BaseLog.ts', '/CreateBaseErrors.ts', '/WriteUnhandledErrorEnvelope.ts', '/IsolateErrorHandlers.ts'];
const paths: readonly string[] = ['/command-line/', '/foundation/orm/schema/', '/foundation/orm/metadata/',
    '/foundation/internal/metadata/', '/account/metadata/', '/base/source/client/', '/base/source/api/',
    '/nexus/source/', '/nexus/code-quality/', '/test/', '/tests/', '/testing/'];
export class Rule {
    readonly context: RuleContext;
    readonly exempt: boolean;
    constructor(context: RuleContext) {
        this.context = context;
        const path = context.parser.path.replaceAll('\\', '/');
        this.exempt = path.endsWith('.test.ts') || captures.some((suffix) => path.endsWith(suffix)) || paths.some((part) => path.includes(part));
    }
    visit(index: number): void {
        if(this.exempt) return;
        const context = this.context;
        const thrown = context.node(index).children[0] ?? -1;
        if(thrown < 0 || context.node(thrown).kind !== 'NewExpression') return;
        const callee = context.node(thrown).children[0] ?? -1;
        if(callee < 0 || context.node(callee).kind !== 'Identifier') return;
        const constructor = context.node(callee).text;
        if(!constructors.includes(constructor)) return;
        context.report(thrown, name, 'bareThrow', message(constructor), '', '', '');
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
