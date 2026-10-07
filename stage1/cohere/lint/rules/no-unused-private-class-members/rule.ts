import type { RuleContext } from '../../context.ts';
function assignment(context: RuleContext, index: number): boolean {
    return (
        context.kind(index) === 'BinaryExpression' &&
        [
            'EqualsToken',
            'PlusEqualsToken',
            'MinusEqualsToken',
            'AsteriskEqualsToken',
            'AsteriskAsteriskEqualsToken',
            'SlashEqualsToken',
            'PercentEqualsToken',
            'LessThanLessThanEqualsToken',
            'GreaterThanGreaterThanEqualsToken',
            'GreaterThanGreaterThanGreaterThanEqualsToken',
            'AmpersandEqualsToken',
            'BarEqualsToken',
            'CaretEqualsToken',
            'BarBarEqualsToken',
            'AmpersandAmpersandEqualsToken',
            'QuestionQuestionEqualsToken',
        ].includes(context.kind(context.child(index, 1)))
    );
}
function destructuring(context: RuleContext, index: number): boolean {
    let current = index;
    while(context.parent(current) >= 0) {
        const parent = context.parent(current);
        if(
            [
                'ArrayLiteralExpression',
                'ObjectLiteralExpression',
                'PropertyAssignment',
                'SpreadElement',
                'ParenthesizedExpression',
            ].includes(context.kind(parent))
        ) {
            current = parent;
            continue;
        }
        return assignment(context, parent) && context.child(parent, 0) === current;
    }
    return false;
}
function writeOnly(context: RuleContext, access: number): boolean {
    const parent = context.parent(access);
    const kind = context.kind(parent);
    const discarded = context.kind(context.parent(parent)) === 'ExpressionStatement';
    if(kind === 'BinaryExpression') {
        return (
            assignment(context, parent) &&
            context.child(parent, 0) === access &&
            (context.kind(context.child(parent, 1)) === 'EqualsToken' || discarded)
        );
    }
    if(kind === 'PrefixUnaryExpression') {
        return ['PlusPlusToken', 'MinusMinusToken'].includes(context.node(parent).operator) && discarded;
    }
    if(kind === 'PostfixUnaryExpression') {
        return discarded;
    }
    if(['ForInStatement', 'ForOfStatement'].includes(kind)) {
        return context.child(parent, context.kind(context.child(parent, 0)) === 'AwaitKeyword' ? 1 : 0) === access;
    }
    if(['ArrayLiteralExpression', 'SpreadElement'].includes(kind)) {
        return destructuring(context, parent);
    }
    if(kind === 'PropertyAssignment') {
        return context.property(parent) === access && destructuring(context, parent);
    }
    return false;
}
function used(context: RuleContext, index: number, name: string, accessor: boolean): boolean {
    if(context.kind(index) === 'PrivateIdentifier') {
        if(context.node(index).text !== name) {
            return false;
        }
        const parent = context.parent(index);
        if(context.memberName(parent) === index) {
            return false;
        }
        return context.kind(parent) !== 'PropertyAccessExpression' || accessor || !writeOnly(context, parent);
    }
    if(['ClassDeclaration', 'ClassExpression'].includes(context.kind(index))) {
        for(const member of context.members(index)) {
            const named = context.memberName(member);
            if(context.kind(named) === 'PrivateIdentifier' && context.node(named).text === name) {
                return false;
            }
        }
    }
    for(const child of context.node(index).children) {
        if(used(context, child, name, accessor)) {
            return true;
        }
    }
    return false;
}
export function noUnusedPrivateClassMembers(context: RuleContext, index: number): void {
    if(!['ClassDeclaration', 'ClassExpression'].includes(context.kind(index))) {
        return;
    }
    const members = context.members(index);
    const seen = new Set<string>();
    for(const member of members) {
        const name = context.memberName(member);
        if(context.kind(name) !== 'PrivateIdentifier') {
            continue;
        }
        const text = context.node(name).text;
        if(seen.has(text)) {
            continue;
        }
        seen.add(text);
        let accessor = false;
        for(const sibling of members) {
            const named = context.memberName(sibling);
            if(
                named >= 0 &&
                context.node(named).text === text &&
                ['GetAccessor', 'SetAccessor'].includes(context.kind(sibling))
            ) {
                accessor = true;
            }
        }
        let read = false;
        for(const sibling of members) {
            if(used(context, sibling, text, accessor)) {
                read = true;
            }
        }
        if(!read) {
            context.reportNode(
                name,
                'no-unused-private-class-members',
                'noUnusedPrivateClassMember',
                `'${text}' is declared but nothing in this class ever reads it. A private member is visible only inside its own class body, so a value no line here reads back can never be observed and the code that maintains it is dead too. Read it, or remove the member and its writes.`,
            );
        }
    }
}

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(index: number): void {
        noUnusedPrivateClassMembers(this.context, index);
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
