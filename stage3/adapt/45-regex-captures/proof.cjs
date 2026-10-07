"use strict";

// Conservative participation proof for the regular-expression forms reviewed
// in this unit. Unsupported constructs decline, rather than imply presence.
function prove(pattern) {
    let position = 0;
    let captures = 0;
    const union = (left, right) => new Set([...left, ...right]);
    const intersection = (left, right) => new Set([...left].filter(value => right.has(value)));
    function disjunction() {
        let mandatory = sequence();
        while (pattern[position] === "|") {
            position++;
            mandatory = intersection(mandatory, sequence());
        }
        return mandatory;
    }
    function sequence() {
        let mandatory = new Set();
        while (position < pattern.length && pattern[position] !== ")" && pattern[position] !== "|") {
            let current = atom();
            const quantifier = pattern[position];
            if (quantifier === "?" || quantifier === "*" || quantifier === "+") {
                position++;
                if (quantifier !== "+") current = new Set();
                if (pattern[position] === "?") position++;
            }
            else if (quantifier === "{") {
                const match = /^\{(\d+)(?:,(\d*)?)?\}/.exec(pattern.slice(position));
                if (!match) throw new Error("unsupported brace syntax");
                position += match[0].length;
                if (Number(match[1]) === 0) current = new Set();
                if (pattern[position] === "?") position++;
            }
            mandatory = union(mandatory, current);
        }
        return mandatory;
    }
    function atom() {
        const character = pattern[position++];
        if (character === "(") {
            let capture;
            if (pattern[position] === "?") {
                if (pattern.slice(position, position + 2) !== "?:") throw new Error("unsupported group assertion or name");
                position += 2;
            }
            else capture = ++captures;
            const mandatory = disjunction();
            if (pattern[position++] !== ")") throw new Error("unclosed group");
            if (capture) mandatory.add(capture);
            return mandatory;
        }
        if (character === "[") {
            let closed = false;
            while (position < pattern.length) {
                const item = pattern[position++];
                if (item === "\\") position++;
                else if (item === "]") { closed = true; break; }
            }
            if (!closed) throw new Error("unclosed character class");
        }
        else if (character === "\\") {
            const escape = pattern[position++];
            if (!escape || /[1-9kpPuUx]/.test(escape)) throw new Error("unsupported escape or backreference");
        }
        else if ("*+?{}".includes(character)) throw new Error("unexpected quantifier");
        return new Set();
    }
    try {
        // Validate syntax separately; construction never executes a match.
        new RegExp(pattern);
        const mandatory = disjunction();
        if (position !== pattern.length) throw new Error("unexpected closing group");
        return { supported: true, captures, mandatory: [...mandatory].sort((a, b) => a - b) };
    }
    catch (error) {
        return { supported: false, captures, mandatory: [], reason: error.message };
    }
}

module.exports = { prove };
