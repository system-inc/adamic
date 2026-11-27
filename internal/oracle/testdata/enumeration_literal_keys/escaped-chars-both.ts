import "./support/iterator.ts";
const hashTable = { "#": 1 };
const escapedCharsMap = new Map(Object.entries({
    "\t": "\\t",
    "\v": "\\v",
    "\f": "\\f",
    "\b": "\\b",
    "\r": "\\r",
    "\n": "\\n",
    "\\": "\\\\",
    '"': '\\"',
    "'": "\\'",
    "`": "\\`",
    "\u2028": "\\u2028", // lineSeparator
    "\u2029": "\\u2029", // paragraphSeparator
    "\u0085": "\\u0085", // nextLine
    "\r\n": "\\r\\n", // special case for CRLFs in backticks
}));
console.log(String(escapedCharsMap.get("\t")));
