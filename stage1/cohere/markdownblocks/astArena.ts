// Acyclic mdast storage: child edges and shared position identities are numeric.
import { panic } from 'adamic';
export interface AstNodeInterface {
    kind: string;
    value: string;
    raw: string;
    hasRaw: boolean;
    literal: boolean;
    parent: boolean;
    ordered: boolean;
    indented: boolean;
    aligned: boolean;
    originalAlt: string;
    hasOriginalAlt: boolean;
    hasPosition: boolean;
    position: number;
    start: number;
    end: number;
    startLine: number;
    endLine: number;
    startColumn: number;
    endColumn: number;
    wordKind: string;
    cj: boolean;
    leading: boolean;
    trailing: boolean;
    children: number[];
}
export class AstArena {
    readonly nodes: AstNodeInterface[] = [];
    nextPosition = 0;
    node(id: number): AstNodeInterface {
        return this.nodes[id] ?? panic('missing mdast node');
    }
    add(kind: string): number {
        const id = this.nodes.length;
        this.nodes.push({
            kind,
            value: '',
            raw: '',
            hasRaw: false,
            literal: false,
            parent: false,
            ordered: false,
            indented: false,
            aligned: false,
            originalAlt: '',
            hasOriginalAlt: false,
            hasPosition: false,
            position: -1,
            start: 0,
            end: 0,
            startLine: 0,
            endLine: 0,
            startColumn: 0,
            endColumn: 0,
            wordKind: '',
            cj: false,
            leading: false,
            trailing: false,
            children: [],
        });
        return id;
    }
    reservePosition(position: number): void {
        if(position >= this.nextPosition) this.nextPosition = position + 1;
    }
    positionFrom(target: number, source: number): void {
        this.node(target).hasPosition = this.node(source).hasPosition;
        this.node(target).position = this.node(source).position;
        this.node(target).start = this.node(source).start;
        this.node(target).end = this.node(source).end;
        this.node(target).startLine = this.node(source).startLine;
        this.node(target).endLine = this.node(source).endLine;
        this.node(target).startColumn = this.node(source).startColumn;
        this.node(target).endColumn = this.node(source).endColumn;
    }
}
