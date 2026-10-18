package javascript

// Preserve the first stored adapter when equal root keys update a collection.
// Returning raw roots from keys/iteration would discard its checking contract.
const viewCallableIdentityCollections = `
class AdamicViewMap extends Map {
    constructor(entries = []) { super(); if (entries != null) for (const [key, value] of entries) this.set(key, value); }
    set(key, value) {
        const root = adamicViewAdapterUnderlying(key);
        if (super.has(root)) super.get(root).value = value;
        else super.set(root, {key: root === 0 ? 0 : key, value});
        return this;
    }
    get(key) { return super.get(adamicViewAdapterUnderlying(key))?.value; }
    has(key) { return super.has(adamicViewAdapterUnderlying(key)); }
    delete(key) { return super.delete(adamicViewAdapterUnderlying(key)); }
    *keys() { for (const entry of super.values()) yield entry.key; }
    *values() { for (const entry of super.values()) yield entry.value; }
    *entries() { for (const entry of super.values()) yield [entry.key, entry.value]; }
    [Symbol.iterator]() { return this.entries(); }
    forEach(callback, receiver) { super.forEach(entry => callback.call(receiver, entry.value, entry.key, this)); }
}
class AdamicViewSet extends Set {
    #originals;
    constructor(values = []) { super(); this.#originals = new Map(); if (values != null) for (const value of values) this.add(value); }
    add(value) {
        const root = adamicViewAdapterUnderlying(value);
        if (!super.has(root)) { super.add(root); this.#originals.set(root, root === 0 ? 0 : value); }
        return this;
    }
    has(value) { return super.has(adamicViewAdapterUnderlying(value)); }
    delete(value) { const root = adamicViewAdapterUnderlying(value); this.#originals.delete(root); return super.delete(root); }
    clear() { this.#originals.clear(); super.clear(); }
    *values() { for (const root of super.values()) yield this.#originals.get(root); }
    keys() { return this.values(); }
    *entries() { for (const value of this.values()) yield [value, value]; }
    [Symbol.iterator]() { return this.values(); }
    forEach(callback, receiver) { super.forEach(root => { const value = this.#originals.get(root); callback.call(receiver, value, value, this); }); }
}
`
