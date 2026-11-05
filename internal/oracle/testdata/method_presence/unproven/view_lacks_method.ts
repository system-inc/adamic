// A checked cast lets in an object whose tag says Circle and which has no area: Node says its area is
// "undefined", where a constant presence would say "function". The cast checks only the tag, and a
// view reads its fields as they come, so nothing about the receiver's construction proves the method.
interface Circle {
    readonly kind: 'circle';
    area(): number;
}
function shape(raw: { readonly kind: string }): string {
    const circle = raw as Circle;
    return typeof circle.area;
}
const real: Circle = {
    kind: 'circle',
    area(): number {
        return 3;
    },
};
console.log(shape({ kind: ['circ', 'le'].join('') }));
console.log(shape(real));
