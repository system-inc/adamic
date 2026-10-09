import { Lexer } from '/workspace/adamic/stage1/cohere/yaml/lexer.ts';
console.log(JSON.stringify(new Lexer().lex('[a, -x]')));
