class Inner {
    readonly value = 1;
    read(): number {
        return this.value;
    }
    touch(): number {
        return this.read();
    }
}
class Outer {
    readonly inner: Inner;
    readonly label: string;
    constructor() {
        this.inner = new Inner();
        this.inner.touch();
        this.label = 'ready';
    }
}
console.log(`${new Outer().inner.value}`);
