// Cohere Markdown preprocess's six complete non-MDX passes on an owned numeric tree.
import { panic } from 'adamic';
import type { AstArena } from './astArena.ts';
import { AstWalk } from './astWalk.ts';
import { bracketContent, isIndented, quoteRaw, trimHTML } from './astSource.ts';
import { markAlignedList } from './astLists.ts';
import { splitText } from './splitText.ts';
export class AstPreprocessor {
    readonly arena: AstArena;
    readonly source: string;
    readonly tabWidth: number;
    readonly openParagraphs = new Set<number>();
    readonly riskyPositions = new Set<number>();
    error = '';
    constructor(arena: AstArena, source: string, tabWidth: number) {
        this.arena = arena;
        this.source = source;
        this.tabWidth = tabWidth;
    }
    raw(id: number): string {
        const node = this.arena.node(id);
        if(!node.hasRaw) this.error = "markdown: Cannot read properties of undefined (reading 'includes')";
        return node.raw;
    }
    merge(id: number): void {
        const node = this.arena.node(id);
        if(!node.parent) return;
        const children: number[] = [];
        for(const child of node.children) {
            const previous = children[children.length - 1];
            if(
                previous !== undefined &&
                this.arena.node(previous).kind === 'text' &&
                this.arena.node(child).kind === 'text'
            ) {
                const merged = this.arena.add('text');
                this.arena.positionFrom(merged, previous);
                this.arena.node(merged).position = this.arena.nextPosition;
                this.arena.nextPosition++;
                this.arena.node(merged).end = this.arena.node(child).end;
                this.arena.node(merged).endLine = this.arena.node(child).endLine;
                this.arena.node(merged).endColumn = this.arena.node(child).endColumn;
                this.arena.node(merged).literal = true;
                this.arena.node(merged).value = this.arena.node(previous).value + this.arena.node(child).value;
                children[children.length - 1] = merged;
            }
            else children.push(child);
        }
        node.children = children;
    }
    markRisk(walk: AstWalk): void {
        const node = this.arena.node(walk.id);
        if(node.kind === 'text') {
            const raw = this.raw(walk.id);
            if(raw.includes('[[')) {
                for(const parent of walk.ancestors)
                    if(this.arena.node(parent).kind === 'paragraph') this.openParagraphs.add(parent);
            }
            if(!raw.includes(']]')) return;
        }
        else if(node.kind !== 'wikiLink') return;
        for(const parent of walk.ancestors) {
            if(this.arena.node(parent).kind === 'paragraph' && this.openParagraphs.has(parent))
                this.riskyPositions.add(this.arena.node(parent).position);
        }
    }
    sentence(walk: AstWalk): number {
        const node = this.arena.node(walk.id);
        if(node.kind !== 'text') return walk.id;
        let text = this.raw(walk.id);
        let paragraph = -1;
        let paragraphDepth = -1;
        for(let index = walk.ancestors.length - 1; index >= 0; index--) {
            const parent = walk.ancestors[index] ?? panic('sentence ancestor');
            if(this.arena.node(parent).kind === 'paragraph') {
                paragraph = parent;
                paragraphDepth = index;
                break;
            }
        }
        if(paragraph >= 0) {
            for(let index = 0; index < paragraphDepth; index++) {
                if(this.arena.node(walk.ancestors[index] ?? panic('quote ancestor')).kind === 'blockquote') {
                    text = quoteRaw(text, node.value);
                    break;
                }
            }
            if(walk.parent >= 0 && this.arena.node(walk.parent).kind === 'paragraph')
                text = trimHTML(
                    text,
                    walk.index === 0,
                    walk.index === this.arena.node(walk.parent).children.length - 1,
                );
        }
        const risky = paragraph >= 0 && this.riskyPositions.has(this.arena.node(paragraph).position);
        const replacement = this.arena.add(risky ? 'text' : 'sentence');
        this.arena.positionFrom(replacement, walk.id);
        if(risky) {
            this.arena.node(replacement).literal = true;
            this.arena.node(replacement).value = text;
            return replacement;
        }
        this.arena.node(replacement).parent = true;
        for(const token of splitText(text)) {
            const child = this.arena.add(token.type);
            this.arena.node(child).literal = true;
            this.arena.node(child).value = token.value;
            this.arena.node(child).wordKind = token.kind;
            this.arena.node(child).cj = token.cj;
            this.arena.node(child).leading = token.leading;
            this.arena.node(child).trailing = token.trailing;
            this.arena.node(replacement).children.push(child);
        }
        return replacement;
    }
    run(root: number): number {
        let result = root;
        for(let pass = 0; pass < 6; pass++) {
            if(pass === 5) {
                const risk = new AstWalk(this.arena, result);
                while(risk.next() && this.error === '') {
                    this.markRisk(risk);
                    risk.descend(risk.id);
                }
            }
            const walk = new AstWalk(this.arena, result);
            while(walk.next() && this.error === '') {
                const node = this.arena.node(walk.id);
                let replacement = walk.id;
                if(pass === 0 && node.kind === 'text') {
                    node.raw = this.source.slice(node.start, node.end);
                    node.hasRaw = true;
                }
                else if(pass === 1) this.merge(walk.id);
                else if(pass === 2 && node.kind === 'code')
                    node.indented = isIndented(this.source.slice(node.start, node.end));
                else if(pass === 3 && (node.kind === 'image' || node.kind === 'imageReference')) {
                    const alt = bracketContent(this.source, node.start, node.end);
                    node.hasOriginalAlt = alt !== undefined;
                    node.originalAlt = alt ?? '';
                }
                else if(pass === 4) markAlignedList(this.arena, walk, this.source, this.tabWidth);
                else if(pass === 5) replacement = this.sentence(walk);
                walk.descend(replacement);
            }
            result = walk.root;
        }
        return result;
    }
}
