import { SchemaPattern } from './schemaPattern.ts';
export class SchemaTag {
    tag: string;
    format: string;
    resolve: string;
    defaultKey: boolean;
    pattern: SchemaPattern;
    constructor(tag: string, format: string, resolve: string, defaultKey: boolean, pattern: string) {
        this.tag = tag;
        this.format = format;
        this.resolve = resolve;
        this.defaultKey = defaultKey;
        this.pattern = new SchemaPattern(pattern);
    }
}
