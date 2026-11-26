function createSet() {
    return {
        kind: "set" as const,
        [Symbol.iterator]: () => {
            let done = false;
            return {
                next: () => {
                    const result = { value: 1, done };
                    done = true;
                    return result;
                },
            };
        },
    };
}
export const held = createSet();
