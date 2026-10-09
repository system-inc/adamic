class Attribute {
    readonly namespace: string | boolean | undefined;
    constructor(namespace: string | boolean | undefined) {
        this.namespace = namespace;
    }
}
const node = new Attribute(true);
console.log(`${node.namespace}`);
