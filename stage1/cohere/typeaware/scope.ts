import type { Binding } from './binding.ts';
export class Scope {
    readonly container: number;
    readonly bindings: Binding[];
    constructor(container: number, bindings: Binding[]) {
        this.container = container;
        this.bindings = bindings;
    }
}
