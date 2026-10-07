import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { message } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    readonly always: boolean;
    constructor(context: RuleContext) {
        this.context = context;
        this.always = context.settings.read('option', 'except-parens') === 'always';
    }
    test(index: number): number {
        const node = this.context.node(index);
        if(node.kind === 'IfStatement' || node.kind === 'WhileStatement' || node.kind === 'ConditionalExpression') {
            return node.children[0] ?? -1;
        }
        if(node.kind === 'DoStatement') {
            return node.children[1] ?? -1;
        }
        if(node.kind === 'ForStatement') {
            // The parser omits absent header clauses. Scan only gaps between children,
            // so semicolons inside initializers cannot be mistaken for clause boundaries.
            let cursor = this.context.start(index);
            let semicolons = 0;
            for(const child of node.children) {
                const childStart = this.context.start(child);
                this.context.scanner.pos = cursor;
                while(this.context.scanner.scan() !== 'EndOfFile' && this.context.scanner.start < childStart) {
                    if(this.context.scanner.kind === 'SemicolonToken') {
                        semicolons++;
                    }
                }
                if(semicolons === 1) {
                    return child;
                }
                cursor = this.context.node(child).end;
            }
        }
        return -1;
    }
    assignment(index: number): boolean {
        if(index < 0 || this.context.node(index).kind !== 'BinaryExpression') {
            return false;
        }
        const operator = this.context.node(index).children[1] ?? panic('missing operator');
        return [
            'EqualsToken', 'PlusEqualsToken', 'MinusEqualsToken', 'AsteriskEqualsToken',
            'AsteriskAsteriskEqualsToken', 'SlashEqualsToken', 'PercentEqualsToken',
            'LessThanLessThanEqualsToken', 'GreaterThanGreaterThanEqualsToken',
            'GreaterThanGreaterThanGreaterThanEqualsToken', 'AmpersandEqualsToken',
            'BarEqualsToken', 'CaretEqualsToken', 'AmpersandAmpersandEqualsToken',
            'BarBarEqualsToken', 'QuestionQuestionEqualsToken',
        ].includes(this.context.node(operator).kind);
    }
    report(index: number): void {
        const operator = this.context.node(index).children[1] ?? panic('missing operator');
        this.context.report(operator, 'no-cond-assign', 'condAssign', message, '', '', '');
    }
    visit(index: number): void {
        if(this.always) {
            if(!this.assignment(index)) {
                return;
            }
            const assignment = this.context.node(index);
            let ancestor = this.context.parents[index] ?? -1;
            while(ancestor >= 0) {
                const node = this.context.node(ancestor);
                if(node.kind === 'ArrowFunction' || node.kind === 'Block') {
                    return;
                }
                const test = this.test(ancestor);
                if(test >= 0 && this.context.node(test).pos <= assignment.pos && assignment.end <= this.context.node(test).end) {
                    this.report(index);
                    return;
                }
                ancestor = this.context.parents[ancestor] ?? -1;
            }
            return;
        }
        let test = this.test(index);
        if(test >= 0 && this.context.node(index).kind === 'ConditionalExpression') {
            test = this.context.unwrap(test);
        }
        if(this.assignment(test)) {
            this.report(test);
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
