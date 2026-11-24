class Custom extends Error { field = 1; }
console.log(String(Object.prototype.hasOwnProperty.call(new Custom(), "field")));
