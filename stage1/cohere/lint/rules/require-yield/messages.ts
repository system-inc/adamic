export const messageYield =
    "This generator function contains no `yield`, so calling it returns an iterator that finishes immediately with the function's return value and never produces anything. A caller writing `for(const item of generate())` gets zero iterations. Either the asterisk is left over from a refactor and the function should be a plain one, or a `yield` is missing.";
