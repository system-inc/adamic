// Match cohere's ASCII guard, including its position-zero shebang exception.
export function canBeginAt(text: string, position: number): boolean {
    if(position < 0 || position > text.length) {
        return false;
    }
    if(position === 0 && text.startsWith('#!')) {
        return true;
    }
    for(let index = position; index < text.length; index++) {
        const code = text.charCodeAt(index);
        if(code === 47) {
            return true;
        }
        if(code !== 32 && (code < 9 || code > 13)) {
            return false;
        }
    }
    return false;
}
