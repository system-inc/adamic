import type { Comment } from './comment.ts';
export function sortByPosition(comments: Comment[]): void {
    for(let outer = 1; outer < comments.length; outer++) {
        const current = comments[outer];
        if(current === undefined) {
            continue;
        }
        let inner = outer - 1;
        while(inner >= 0 && (comments[inner]?.start ?? -1) > current.start) {
            const previous = comments[inner];
            if(previous !== undefined) {
                // eslint-disable-next-line nexus/correctness-no-caller-data-mutation -- Go's sortByPosition sorts its slice in place; this API keeps that contract.
                comments[inner + 1] = previous;
            }
            inner--;
        }
        // eslint-disable-next-line nexus/correctness-no-caller-data-mutation -- This is the final write in the explicitly in-place sort.
        comments[inner + 1] = current;
    }
}
