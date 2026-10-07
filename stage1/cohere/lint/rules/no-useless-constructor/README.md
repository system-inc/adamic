# no-useless-constructor

`kinds` lists MethodDeclaration beyond the Go rule's listeners (Constructor). The stage 1 parser keeps a quoted
`'constructor'` member as a MethodDeclaration where typescript-go makes it a Constructor, so the rule listens there too and
re-gates on the name. Remove the extra kind when the parser makes the same node Go does.
