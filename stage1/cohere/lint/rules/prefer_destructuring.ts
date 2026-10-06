import { panic } from 'adamic';
import type { RuleContext } from './context.ts';
import { Finding } from '../finding.ts';
import { initializer } from './declaration_initializer.ts';
import { first, last, name, parent } from './shared.ts';

function commentsInside(context: RuleContext, start: number, end: number): number {
    let count = 0;
    for(const comment of context.comments) {
        if(comment.start >= start && comment.end <= end) {
            count++;
        }
    }
    return count;
}
export function preferDestructuring(context: RuleContext, index: number): void {
    const node = context.node(index);
    const declarator = node.kind === 'VariableDeclaration';
    let left: number;
    let right: number;
    if(declarator) {
        const list = parent(context, index);
        if(list >= 0 && ['4', '6'].includes(context.node(list).semantic)) {
            return;
        }
        left = name(context, index);
        right = initializer(context, index);
    }
    else {
        if(context.node(node.children[1] ?? panic('operator')).kind !== 'EqualsToken') {
            return;
        }
        left = first(context, index);
        right = node.children[2] ?? -1;
    }
    if(right < 0) {
        return;
    }
    right = context.unwrap(right);
    const member = context.node(right);
    if(
        !['PropertyAccessExpression', 'ElementAccessExpression'].includes(member.kind) ||
        context.hasKind(right, 'QuestionDotToken')
    ) {
        return;
    }
    const computed = member.kind === 'ElementAccessExpression';
    const property = last(context, right);
    const receiver = first(context, right);
    if(context.node(receiver).kind === 'SuperKeyword' || context.node(property).kind === 'PrivateIdentifier') {
        return;
    }
    const key = context.node(property);
    let array = computed && key.kind === 'NumericLiteral' && key.text !== '';
    if(array) {
        for(const character of key.text.split('')) {
            if(character < '0' || character > '9') {
                array = false;
            }
        }
    }
    const option = declarator ? 'variabledeclarator' : 'assignmentexpression';
    const configured = context.settings.values.size !== 0;
    if(context.settings.read(`${option}.${array ? 'array' : 'object'}`, configured ? 'false' : 'true') !== 'true') {
        return;
    }
    if(!array && context.settings.read('enforceforrenamedproperties', 'false') !== 'true') {
        if(
            left < 0 ||
            context.node(left).kind !== 'Identifier' ||
            context.node(left).text !== key.text ||
            (computed ? key.kind !== 'StringLiteral' : key.kind !== 'Identifier')
        ) {
            return;
        }
    }
    let replacement = '';
    let repair = '';
    if(
        declarator &&
        !computed &&
        left >= 0 &&
        context.node(left).kind === 'Identifier' &&
        key.kind === 'Identifier' &&
        context.node(left).text === key.text
    ) {
        const object = context.unwrap(receiver);
        if(
            commentsInside(context, context.start(index), node.end) <=
            commentsInside(context, context.start(object), context.node(object).end)
        ) {
            let text = context.text(object);
            if(
                context.node(object).kind === 'BinaryExpression' &&
                context.node(context.node(object).children[1] ?? panic('operator')).kind === 'CommaToken'
            ) {
                text = `(${text})`;
            }
            replacement = `{${key.text}} = ${text}`;
            repair = 'fix';
        }
    }
    const message = `Use ${array ? 'array' : 'object'} destructuring. Pulling one name out of an object or array by hand repeats the name on both sides, so a rename touches two places and the two can drift. Destructuring writes it once.`;
    context.record(
        new Finding(
            'prefer-destructuring',
            'preferDestructuring',
            message,
            context.start(index),
            node.end,
            repair,
            replacement,
            '',
        ),
    );
}

export function visit(context: RuleContext, index: number): void {
    if(
        ['VariableDeclaration', 'BinaryExpression'].includes(context.node(index).kind) &&
        context.enabled('prefer-destructuring')
    ) {
        preferDestructuring(context, index);
    }
}
