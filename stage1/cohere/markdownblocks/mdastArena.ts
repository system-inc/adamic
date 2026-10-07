// Full mdast construction fields, with numeric child edges instead of owning cycles.
import { panic } from 'adamic';
import { encode } from './codec.ts';
import type { ParsedFrontMatter } from './parseFrontMatter.ts';
import { MdastNode } from './mdastNode.ts';
export class MdastArena {
    readonly nodes: MdastNode[] = [];
    node(id: number): MdastNode {
        return this.nodes[id] ?? panic('missing constructed mdast node');
    }
    add(type: string, parent: boolean, literal: boolean): number {
        const id = this.nodes.length;
        const node = new MdastNode(type);
        node.parent = parent;
        node.literal = literal;
        this.nodes.push(node);
        return id;
    }
    attachFrontMatter(front: ParsedFrontMatter): void {
        if(!front.present) return;
        const id = this.add('frontMatter', false, false);
        const node = this.node(id);
        node.frontMatter = front;
        node.start.line = 1;
        node.start.column = 1;
        const lines = front.raw.split('\n');
        node.end.line = lines.length;
        node.end.column = (lines[lines.length - 1] ?? panic('front matter last line')).length + 1;
        node.end.offset = front.raw.length;
        this.node(0).children.splice(0, 0, id);
    }
    toString(id: number): string {
        const output: string[] = [];
        const stack: number[] = [id];
        while(stack.length > 0) {
            const node = this.node(stack.pop() ?? panic('mdast text stack'));
            if(node.literal) output.push(node.value);
            else if(node.hasAlt && node.alt !== '') output.push(node.alt);
            else if(node.parent) {
                for(let index = node.children.length - 1; index >= 0; index--)
                    stack.push(node.children[index] ?? panic('mdast child'));
            }
        }
        return output.join('');
    }
    canonical(root: number): string {
        const output: string[] = [];
        const stack: number[] = [root];
        while(stack.length > 0) {
            const node = this.node(stack.pop() ?? panic('mdast canonical stack'));
            const fields: string[] = [
                node.type,
                `${node.parent ? node.children.length : -1}`,
                `${node.literal}`,
                `${node.valueNull}`,
                node.value,
                `${node.depth}`,
                `${node.ordered}`,
                node.hasStartValue ? `+${node.startValue}` : '-',
                `${node.spread}`,
                node.hasChecked ? `${node.checked}` : '-',
                node.hasLang ? `+${node.lang}` : '-',
                node.hasMeta ? `+${node.meta}` : '-',
                node.url,
                node.hasTitle ? `+${node.title}` : '-',
                node.hasAlt ? `+${node.alt}` : '-',
                node.hasLabel ? `+${node.label}` : '-',
                node.identifier,
                node.referenceType,
                node.align.join(','),
                `${node.start.line},${node.start.column},${node.start.offset},${node.end.line},${node.end.column},${node.end.offset}`,
                node.frontMatter === undefined ? '-' : `+${node.frontMatter.language}`,
                node.frontMatter === undefined || node.frontMatter.explicitLanguage === ''
                    ? '-'
                    : `+${node.frontMatter.explicitLanguage}`,
                node.frontMatter?.value ?? '',
                node.frontMatter?.startDelimiter ?? '',
                node.frontMatter?.endDelimiter ?? '',
                node.frontMatter?.raw ?? '',
            ];
            const row: string[] = [];
            for(const field of fields) row.push(encode(field));
            output.push(row.join('\t'));
            for(let index = node.children.length - 1; index >= 0; index--)
                stack.push(node.children[index] ?? panic('mdast child'));
        }
        return output.join('\n');
    }
}
