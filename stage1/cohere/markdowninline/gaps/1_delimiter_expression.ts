// Original matching syntax. The slice uses its finite scanner instead.
console.log('a*b'.replaceAll(/(\\+|^|.)(\*+|_+)($|.)/g, '$1\\$2$3'));
