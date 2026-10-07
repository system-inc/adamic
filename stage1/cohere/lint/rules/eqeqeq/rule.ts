import { panic } from 'adamic';
import { description, suggestionDescription } from './messages.ts';
import type { RuleContext } from '../../context.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(index: number): void {
        const children = this.context.node(index).children;
        if(children.length !== 3) {
            return;
        }
        const left = children[0] ?? panic('missing left');
        const operator = children[1] ?? panic('missing operator');
        const right = children[2] ?? panic('missing right');
        const actual = this.context.source.slice(this.context.start(operator), this.context.node(operator).end);
        const isNull =
            this.context.node(this.context.unwrap(left)).kind === 'NullKeyword' ||
            this.context.node(this.context.unwrap(right)).kind === 'NullKeyword';
        const hasTypeOf =
            this.context.node(this.context.unwrap(left)).kind === 'TypeOfExpression' ||
            this.context.node(this.context.unwrap(right)).kind === 'TypeOfExpression';
        const type = this.context.literalType(left);
        const sameType = type !== '' && type === this.context.literalType(right);
        let expected: string;
        if(actual === '===' || actual === '!==') {
            if(this.context.nullPolicy !== 'Never' || !isNull) {
                return;
            }
            expected = actual === '===' ? '==' : '!=';
        }
        else if(actual === '==' || actual === '!=') {
            if(this.context.mode === 'Smart' && (hasTypeOf || sameType || isNull)) {
                return;
            }
            if(this.context.nullPolicy !== 'Always' && isNull) {
                return;
            }
            expected = actual === '==' ? '===' : '!==';
        }
        else {
            return;
        }
        const message = description(actual, expected);
        const suggestion = suggestionDescription(actual, expected);
        this.context.report(
            operator,
            'eqeqeq',
            'unexpected',
            message,
            hasTypeOf || sameType ? 'fix' : 'suggestion',
            expected,
            hasTypeOf || sameType ? '' : suggestion,
        );
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
