import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageNoConstEnumSuffix, messageNoInterfaceSuffix, messageNoTypeAliasSuffix } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(node: ParseNode, index: number): void {
        const name = this.context.name(index);
        if(name < 0) {
            return;
        }
        const text = this.context.node(name).text;
        const suffixes =
            node.kind === 'InterfaceDeclaration'
                ? ['Interface', 'Properties', 'Options']
                : ['Type', 'Properties', 'Interface', 'Options'];
        const needs =
            node.kind === 'VariableDeclaration'
                ? this.context.node(name).kind === 'Identifier' &&
                  this.constEnum(this.context.initializer(index)) &&
                  !text.endsWith('Kind')
                : !suffixes.some((suffix) => text.endsWith(suffix));
        if(!needs) {
            return;
        }
        if(node.kind === 'VariableDeclaration') {
            this.report(name, 'noConstEnumSuffix', messageNoConstEnumSuffix(text));
        }
        else if(node.kind === 'InterfaceDeclaration') {
            this.report(name, 'noInterfaceSuffix', messageNoInterfaceSuffix(text));
        }
        else {
            this.report(name, 'noTypeAliasSuffix', messageNoTypeAliasSuffix(text));
        }
    }
    report(name: number, id: string, message: string): void {
        this.context.report(name, 'nexus/consistency-require-type-suffix', id, message, '', '', '');
    }
    // constEnum is whether the initializer is an `as const` object whose every key names its own string value.
    constEnum(index: number): boolean {
        if(index < 0 || this.context.node(index).kind !== 'AsExpression') {
            return false;
        }
        const children = this.context.node(index).children;
        const expression = children[0] ?? -1;
        const type = children[1] ?? -1;
        if(expression < 0 || type < 0 || this.context.node(type).kind !== 'TypeReference') {
            return false;
        }
        const typeName = this.context.node(type).children[0] ?? -1;
        if(
            typeName < 0 ||
            this.context.node(typeName).text !== 'const' ||
            this.context.node(expression).kind !== 'ObjectLiteralExpression' ||
            this.context.node(expression).children.length === 0
        ) {
            return false;
        }
        for(const property of this.context.node(expression).children) {
            if(this.context.node(property).kind !== 'PropertyAssignment') {
                return false;
            }
            const key = this.context.node(property).children[0] ?? -1;
            const value = this.context.node(property).children[1] ?? -1;
            if(
                key < 0 ||
                value < 0 ||
                this.context.node(key).kind !== 'Identifier' ||
                this.context.node(value).kind !== 'StringLiteral' ||
                this.context.node(key).text !== this.context.node(value).text
            ) {
                return false;
            }
        }
        return true;
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
