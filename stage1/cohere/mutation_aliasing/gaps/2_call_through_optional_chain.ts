// A method called through ?.
class Node {
    readonly edges: number[] = [];
}
const nodes = new Map<number, Node>([[1, new Node()]]);
nodes.get(1)?.edges.push(7);
nodes.get(2)?.edges.push(8);
console.log(`${nodes.get(1)?.edges.length ?? 0}`);
