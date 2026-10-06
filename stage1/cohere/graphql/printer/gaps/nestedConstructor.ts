import { Inner } from './nestedInner.ts';
// Positions differ between modules. This padding isolates that condition.
class Holder {
    readonly inner: Inner;
    constructor(label: string) {
        this.inner = new Inner(label);
    }
}
const holder = new Holder('ready');
console.log(holder.inner.show());
