import { panic } from 'adamic';
import { Parser } from '../../../../typescript/parser/parser.ts';
import type { RuleContext } from '../../context.ts';
import { SuggestionEdit as Edit, Suggestion } from '../../suggestions.a';
const noEdits: Edit[] = [];
class Entity {
    readonly character: string;
    readonly alternatives: string[];
    readonly hasAlternatives: boolean;
    constructor(character: string, alternatives: string[], hasAlternatives: boolean) {
        this.character = character; this.alternatives = alternatives; this.hasAlternatives = hasAlternatives;
    }
}
export class Rule {
    readonly context: RuleContext;
    readonly entities: Entity[] = [];
    readonly plans = new Map<number, Suggestion[]>();
    constructor(context: RuleContext) {
        this.context = context;
        const settings = context.settings;
        const raw = settings.text;
        const absent: string[] = [];
        const configured = settings.list('forbid', absent);
        if(raw === '' || raw === 'null' || configured === absent) {
            this.entities.push(new Entity('>', ['&gt;'], true));
            this.entities.push(new Entity('"', ['&quot;', '&ldquo;', '&#34;', '&rdquo;'], true));
            this.entities.push(new Entity("'", ['&apos;', '&lsquo;', '&#39;', '&rsquo;'], true));
            this.entities.push(new Entity('}', ['&#125;'], true));
        }
        else {
        // Reparse option data to distinguish string entries from objects and null.
        const parser = new Parser(`(${raw});`, 'rule options');
        const statement = parser.node(parser.file()).children[0] ?? panic('missing options');
        const outer = parser.node(statement).children[0] ?? panic('missing expression');
        const object = parser.node(outer).children[0] ?? panic('missing object');
        for(const property of parser.node(object).children) {
            const parts = parser.node(property).children;
            const key = parser.node(parts[0] ?? panic('missing key')).text.toLowerCase();
            if(key !== 'forbid') { continue; }
            const array = parser.node(parts[1] ?? panic('missing value'));
            for(const index of array.children) {
                const entry = parser.node(index);
                if(entry.kind === 'StringLiteral') { this.entities.push(new Entity(entry.text, [], false)); }
                else if(entry.kind === 'ObjectLiteralExpression') {
                    let character = ''; let hasCharacter = false; let hasAlternatives = false;
                    const alternatives: string[] = [];
                    for(const field of entry.children) {
                        const children = parser.node(field).children;
                        const name = parser.node(children[0] ?? panic('missing field name')).text.toLowerCase();
                        const value = parser.node(children[1] ?? panic('missing field value'));
                        if(name === 'char' && value.kind === 'StringLiteral') { character = value.text; hasCharacter = true; }
                        if(name === 'alternatives' && value.kind === 'ArrayLiteralExpression') {
                            hasAlternatives = true;
                            for(const alternative of value.children) { alternatives.push(parser.node(alternative).text); }
                        }
                    }
                    if(hasCharacter && hasAlternatives) { this.entities.push(new Entity(character, alternatives, true)); }
                }
            }
        }
        }
    }
    suggestions(start: number): Suggestion[] { return this.plans.get(start) ?? []; }
    visit(index: number, parent: number): void {
        const node = this.context.node(index);
        if(node.kind !== 'JsxText') { return; }
        const raw = this.context.source.slice(node.pos, node.end);
        for(let offset = 0; offset < raw.length; offset++) {
            for(const entity of this.entities) {
                // The Go rule compares a single UTF-8 byte, so non-ASCII entries never match.
                if(entity.character.length !== 1 || (entity.character.charCodeAt(0) ?? 0) > 127 || raw[offset] !== entity.character) { continue; }
                const start = node.pos + offset;
                const id = entity.hasAlternatives ? 'unescapedEntityAlts' : 'unescapedEntity';
                const message = entity.hasAlternatives ? `\`${entity.character}\` can be escaped with ${entity.alternatives.map(value => '\x60' + value + '\x60').join(', ')}.` : `HTML entity \`${entity.character}\` is written raw in JSX text and must be escaped.`;
                const plans: Suggestion[] = [];
                if(entity.hasAlternatives) {
                    for(const alternative of entity.alternatives) {
                        plans.push(new Suggestion('replaceWithAlt', `Replace with \`${alternative}\`.`, [new Edit(node.pos, node.end, raw.slice(0, offset) + alternative + raw.slice(offset + 1))]));
                    }
                }
                this.context.reportRange(start, start + 1, 'react/no-unescaped-entities', id, message, noEdits, plans);
                this.plans.set(this.context.findings.length - 1, plans);
            }
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
