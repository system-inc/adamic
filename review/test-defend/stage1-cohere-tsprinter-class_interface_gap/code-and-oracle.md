CODE UNDER TEST: Expressions family: stage1/cohere/tsprinter/expressions.ts formatExpression and Expressions printer, with Documents and Parser dependencies. Prefix gap: internal/lower/expression.go lowering.prefix, lowering.increment and native/JavaScript emitters. Number: internal/lower/library_math_number.go lowering.libraryNumber and represented Number conversion runtime.

ORACLE: Expressions executes Go cohere and Prettier 3.9.6 and embedded Prettier, comparing exact text with pinned allowance records. Prefix and Number run Node alongside native and JavaScript backends and require literal 2 and 17 respectively, plus leak checks.

Aim P1 at prefix update used as a value, unlike document loops incremented as statements. Aim N1 at literal-string AST conversion, unlike computed Number operands in document/main transports. Aim E1 at standalone parenthesized string formatting; document main never calls formatExpression. All aim declarations preceded their respective runs.
