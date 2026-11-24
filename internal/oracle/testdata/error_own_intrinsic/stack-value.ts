function read(value: unknown, key: string): unknown {
    if (value === null || value === undefined) {
        return undefined;
    }
    return Reflect.get(Object(value), key, value);
}
console.log(typeof read(new Error("x"), "stack"));
