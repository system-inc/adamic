import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { assignment } from '../func-name-matching/rule.ts';
import { record } from '../typescript-no-non-null-asserted-optional-chain/diagnostic.ts';
const wrong = 'This name is designated as the alias for `this`, and it is holding something else. A reader who knows the convention will read it as the captured context and be wrong, which is worse than an ordinary misleading name because the convention is what made it trustworthy.';
const unexpected = 'This captures `this` under a name that is not the designated alias. Every capture in the codebase should read the same way, so a reader recognises the pattern instead of working out what each new name means.';
function scope(kind: string): boolean {
    return ['FunctionDeclaration','FunctionExpression','MethodDeclaration','GetAccessor','SetAccessor','Constructor'].includes(kind);
}
export class Rule {
    readonly context: RuleContext;
    readonly aliases: string[];
    constructor(context: RuleContext) {
        this.context = context;
        const aliases = context.settings.list('aliases',['that']);
        this.aliases = aliases.length === 0 ? ['that'] : aliases;
    }
    report(index: number, id: string): void {
        record(this.context,index,'consistent-this',id,id === 'unexpectedAlias' ? unexpected : wrong,[]);
    }
    check(index: number, name: string, value: number, operator: string): void {
        const isThis = value >= 0 && this.context.node(value).kind === 'ThisKeyword';
        if(this.aliases.includes(name)) {
            if(!isThis || operator !== 'EqualsToken') { this.report(index,'aliasNotAssignedToThis'); }
        }
        else if(isThis) { this.report(index,'unexpectedAlias'); }
    }
    nested(index: number, root: number): boolean {
        let ancestor = this.context.parents[index] ?? -1;
        while(ancestor >= 0 && ancestor !== root) {
            const kind = this.context.node(ancestor).kind;
            const parent = this.context.parents[ancestor] ?? -1;
            if(kind === 'Block') {
                if(!(parent === root && scope(this.context.node(root).kind))) { return true; }
            }
            if(['SwitchStatement','CaseBlock','CatchClause','WithStatement'].includes(kind)) { return true; }
            ancestor = parent;
        }
        return false;
    }
    walk(index: number, root: number, only: boolean, declared: number[], rescued: Map<string, boolean>, nestedScopes: number[]): void {
        const node = this.context.node(index);
        if(scope(node.kind)) {
            if(only) { this.judge(index); } else { nestedScopes.push(index); }
            return;
        }
        if(node.kind === 'ArrowFunction') {
            for(const child of node.children) { this.walk(child,root,true,declared,rescued,nestedScopes); }
            return;
        }
        if(node.kind === 'VariableDeclaration') {
            const name = this.context.node(node.children[0] ?? panic('missing binding'));
            const last = node.children[node.children.length - 1] ?? panic('missing declaration child');
            // Declaration type nodes are not initializers. '=' is omitted by the parser.
            let initializer = -1;
            const text = this.context.source.slice(name.end,this.context.node(last).pos);
            if(last !== node.children[0] && text.includes('=')) { initializer = last; }
            if(name.kind === 'Identifier') {
                if(initializer >= 0) { this.check(index,name.text,initializer,'EqualsToken'); }
                if(!only && this.aliases.includes(name.text)) {
                    declared.push(index);
                    if(initializer >= 0) { rescued.set(name.text,true); }
                }
            }
        }
        else if(node.kind === 'BinaryExpression') {
            const left = this.context.node(node.children[0] ?? panic('missing left'));
            const operator = this.context.node(node.children[1] ?? panic('missing operator')).kind;
            const right = node.children[2] ?? panic('missing right');
            if(assignment(operator) && left.kind === 'Identifier') {
                this.check(index,left.text,right,operator);
                if(!only && this.aliases.includes(left.text) && operator === 'EqualsToken' && this.context.node(right).kind === 'ThisKeyword' && !this.nested(index,root)) { rescued.set(left.text,true); }
            }
        }
        for(const child of node.children) { this.walk(child,root,only,declared,rescued,nestedScopes); }
    }
    judge(index: number): void {
        const declared: number[] = [];
        const rescued = new Map<string, boolean>();
        const nestedScopes: number[] = [];
        for(const child of this.context.node(index).children) { this.walk(child,index,false,declared,rescued,nestedScopes); }
        for(const declaration of declared) {
            const name = this.context.node(this.context.node(declaration).children[0] ?? panic('missing alias')).text;
            if(!rescued.has(name)) { this.report(declaration,'aliasNotAssignedToThis'); }
        }
        for(const child of nestedScopes) { this.judge(child); }
    }
    visit(index: number, parent: number): void { this.judge(index); }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
