import type { Rules } from './rules.ts';
import type { Types } from './types.ts';
import type { TypeFact } from './type_fact.ts';
import { types } from './facts.ts';
import { CheckerFacts } from './checker_facts.ts';
import { Diagnostic } from './diagnostic.ts';
import { stringLike, numberLike, booleanLike, bigintLike } from './flags.ts';

export class Casts {
    readonly rules: Rules;
    readonly facts: CheckerFacts;
    constructor(rules: Rules) { this.rules = rules; this.facts = new CheckerFacts(rules); }
    unsafe(graph: Types, type: TypeFact, seen: number[]): boolean {
        if((type.flags & 1) !== 0) { return true; }
        if(seen.includes(type.id)) { return false; }
        seen.push(type.id);
        return type.arguments.some((id) => this.unsafe(graph, graph.type(id), seen));
    }
    category(type: TypeFact): string {
        if((type.flags & stringLike) !== 0) { return 'string'; }
        if((type.flags & numberLike) !== 0) { return 'number'; }
        if((type.flags & booleanLike) !== 0) { return 'boolean'; }
        if((type.flags & bigintLike) !== 0) { return 'bigint'; }
        if((type.flags & (512 | 16384)) !== 0) { return 'symbol'; }
        if((type.flags & (4 | 16)) !== 0) { return 'undefined'; }
        if((type.flags & 8) !== 0) { return 'null'; }
        return 'object';
    }
    unit(type: TypeFact): boolean { return (type.flags & (1024 | 2048 | 8192 | 32768 | 4 | 8)) !== 0; }
    derives(index: number, type: TypeFact, symbol: number, seen: number[]): boolean {
        const bases = types(this.rules.ask(index, `base-shapes\n${type.id}`), 'base-shapes');
        for(const id of bases.roots) {
            const base = bases.type(id);
            const own = this.facts.metadata(index, base).classSymbol();
            if(own === 0 || seen.includes(own)) { continue; }
            if(own === symbol) { return true; }
            seen.push(own);
            if(this.derives(index, base, symbol, seen)) { return true; }
        }
        return false;
    }
    distinguish(index: number, left: TypeFact, right: TypeFact): boolean {
        const category = this.category(left);
        if(category !== this.category(right)) { return true; }
        if(category !== 'object') { return this.unit(left) && this.unit(right); }
        const l = this.facts.metadata(index, left);
        const r = this.facts.metadata(index, right);
        if(l.isClass() !== r.isClass()) { return true; }
        if(l.isClass() && r.isClass() && l.classSymbol() !== r.classSymbol() &&
            !this.derives(index, left, r.classSymbol(), []) && !this.derives(index, right, l.classSymbol(), [])) { return true; }
        const lp = this.facts.properties(index, left);
        const rp = this.facts.properties(index, right);
        for(let at = 0; at < lp.names.length; at++) {
            const match = rp.names.indexOf(lp.names[at] ?? '');
            if(match < 0) { continue; }
            const lt = lp.types.root(at);
            const rt = rp.types.root(match);
            if(this.unit(lt) && this.unit(rt) && !this.facts.compare(index, lt.id, rt.id, true)) { return true; }
        }
        return false;
    }
    checked(index: number, source: Types, target: Types): boolean {
        if((source.root().flags & 134217728) === 0) { return false; }
        const members = target.parts(target.root());
        const kept: number[] = [];
        const dropped: number[] = [];
        for(const id of source.root().parts) {
            if(members.some((member) => this.facts.compare(index, member, id, true))) { kept.push(id); }
            else { dropped.push(id); }
        }
        if(kept.length === 0 || kept.length !== members.length) { return false; }
        return kept.every((keep) => dropped.every((drop) => this.distinguish(index, source.type(keep), source.type(drop))));
    }
    run(): void {
        for(let index = 0; index < this.rules.parser.nodes.length; index++) {
            if((this.rules.parents[index] ?? -1) < 0) { continue; }
            const node = this.rules.parser.node(index);
            if(!['AsExpression', 'TypeAssertionExpression'].includes(node.kind)) { continue; }
            const expression = node.kind === 'AsExpression' ? node.children[0] ?? -1 : node.children[1] ?? -1;
            const annotation = node.kind === 'AsExpression' ? node.children[1] ?? -1 : node.children[0] ?? -1;
            const typeNode = this.rules.parser.node(annotation);
            if(typeNode.kind === 'TypeReference' && typeNode.children.length === 1 &&
                this.rules.parser.node(typeNode.children[0] ?? -1).text === 'const') { continue; }
            const source = types(this.rules.ask(expression, 'raw-shape'), 'raw-shape');
            const target = types(this.rules.ask(annotation, 'annotation-shape'), 'annotation-shape');
            const s = source.root(); const t = target.root();
            if(s.id === t.id) { continue; }
            if(!this.unsafe(source, s, [])) {
                if(this.facts.compare(index, s.id, t.id) || this.checked(index, source, target)) { continue; }
            }
            else if((t.flags & 3) !== 0 || (t.array && t.arguments.length === 1 &&
                (target.type(t.arguments[0] ?? 0).flags & 2) !== 0)) { continue; }
            const sourceName = this.rules.name(s, index);
            const targetName = this.rules.name(t, index);
            const message = `'${sourceName}' can't be cast to '${targetName}' here: nothing at runtime could confirm it, so the cast is trusted, not proven, and every use after it inherits the guess. Return or declare '${targetName}' where the value comes from, or build one from fields you have narrowed with typeof. A cast from a union to some of its members is allowed when a discriminant property, typeof or instanceof can tell them apart, since Adamic compiles that to a check.`;
            this.rules.findings.push(new Diagnostic('adamic/no-unchecked-cast', 'uncheckedCast', message,
                this.rules.byte(this.rules.start(node)), this.rules.byte(node.end), ''));
        }
    }
}
