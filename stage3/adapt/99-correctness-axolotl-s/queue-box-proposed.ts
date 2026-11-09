interface Queue<T> { enqueue(...items: T[]): void; dequeue(): T; isEmpty(): boolean; }
export function createQueue<T>(items?: readonly T[]): Queue<T> {
    const elements: ({ value: T } | undefined)[] = [];
    if (items) for (const value of items) elements.push({ value });
    let headIndex = 0;

    function isEmpty() {
        return headIndex === elements.length;
    }

    function enqueue(...items: T[]) {
        for (const value of items) elements.push({ value });
    }

    function dequeue(): T {
        if (isEmpty()) {
            throw new Error("Queue is empty");
        }

        const box = elements[headIndex];
        if (box === undefined) throw new Error("Queue invariant violated");
        elements[headIndex] = undefined; // Don't keep referencing dequeued item
        headIndex++;

        // If more than half of the queue is empty, copy the remaining elements to the
        // front and shrink the array (unless we'd be saving fewer than 100 slots)
        if (headIndex > 100 && headIndex > (elements.length >> 1)) {
            const newLength = elements.length - headIndex;
            elements.copyWithin(/*target*/ 0, /*start*/ headIndex);

            elements.splice(newLength);
            headIndex = 0;
        }

        return box.value;
    }

    return {
        enqueue,
        dequeue,
        isEmpty,
    };
}
