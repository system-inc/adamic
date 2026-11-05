interface Queue<T> { enqueue(...items: T[]): void; dequeue(): T; isEmpty(): boolean; }
export function createQueue<T>(items?: readonly T[]): Queue<T> {
    const elements: ({ value: T } | undefined)[] = items?.map(value => ({ value })) ?? [];
    let headIndex = 0;
    function isEmpty() { return headIndex === elements.length; }
    function enqueue(...items: T[]) {
        for (const value of items) elements.push({ value });
    }
    function dequeue(): T {
        if (isEmpty()) throw new Error("Queue is empty");
        const slot = elements[headIndex];
        if (slot === undefined) throw new Error("Queue invariant violated");
        const result = slot.value;
        elements[headIndex] = undefined;
        headIndex++;
        if (headIndex > 100 && headIndex > (elements.length >> 1)) {
            const newLength = elements.length - headIndex;
            elements.copyWithin(0, headIndex);
            elements.splice(newLength);
            headIndex = 0;
        }
        return result;
    }
    return { enqueue, dequeue, isEmpty };
}
