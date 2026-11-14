# Paired latent census

| Measure | Raw findings | Distinct reads |
| --- | ---: | ---: |
| Lowered (no remaining finding) | 0 | 0 |
| Moved by reason | 0 | 0 |
| New findings | 0 | 0 |
| New sites | 0 | 0 |
| Units whose first stop moved | 0 | 0 |
| Regressions | 0 | 0 |
| before NotYet | 40 | 40 |
| before Refused | 1372 | 1267 |
| after NotYet | 40 | 40 |
| after Refused | 1372 | 1267 |

| Reason | Raw before | Raw after | Reads before | Reads after | Reads lowered |
| --- | ---: | ---: | ---: | ---: | ---: |
| Object enumeration with a getter or accessor | 3 | 3 | 3 | 3 | 0 |
| Object enumeration with getter or accessor descriptors | 2 | 2 | 2 | 2 | 0 |
| a ClassExpression | 1 | 1 | 1 | 1 | 0 |
| a CompilerOptions seen as CompilerOptions (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a cast that makes a viewed readonly array writable | 3 | 3 | 3 | 3 | 0 |
| a cast the runtime can't check | 379 | 379 | 379 | 379 | 0 |
| a class used from a function that may run before the class's declaration (its temporal dead zone) | 3 | 3 | 3 | 3 | 0 |
| a function taking IndexSignatureDeclaration \| undefined seen as one taking SignatureDeclaration \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking Node seen as one taking Node \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking SourceFile seen as one taking SourceFile \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking Symbol seen as one taking Symbol \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking SymbolFlags seen as one taking SymbolFlags \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking SyntaxKind.MethodSignature \| SyntaxKind.MethodDeclaration \| SyntaxKind.Constructor \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.CallSignature \| SyntaxKind.ConstructSignature \| ... 6 more ... \| SyntaxKind.JSDocFunctionType seen as one taking SyntaxKind (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking TypeFacts.None seen as one taking number (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking Visitor seen as one taking Visitor<TIn, Node \| undefined> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (cloneNode would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createPropertyDeclaration would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createTemplateMiddle would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createTemplateTail would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (reportPrivateInBaseOfClassExpression would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (throwIfCancellationRequested would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a presence test of a method whose presence the receiver's construction doesn't prove; a view or host value has to read it | 1 | 1 | 1 | 1 | 0 |
| a read narrowed to undefined of a property a call may write | 6 | 6 | 6 | 6 | 0 |
| a read of a void value, as a returned value | 1 | 1 | 1 | 1 | 0 |
| a structural relation deeper than 64 type pairs while checking promise-returning functions seen as void-returning | 18 | 18 | 18 | 18 | 0 |
| a value of type (ClassDeclaration \| EnumDeclaration \| ExportAssignment \| ExportDeclaration \| FunctionDeclaration \| ... 5 more ... \| VariableStatement)[] seen as Statement[], which can write Statement where ClassDeclaration \| EnumDeclaration \| ExportAssignment \| ExportDeclaration \| FunctionDeclaration \| ... 5 more ... \| VariableStatement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (readonly [() => IntrinsicType, __String])[] seen as (readonly [() => Type, __String])[], which can write readonly [() => Type, __String] where readonly [() => IntrinsicType, __String] is read | 2 | 2 | 2 | 2 | 0 |
| a value of type AccessorDeclaration[] seen as T \| (T \| AccessorDeclaration)[] \| AccessorDeclaration, which can write T \| AccessorDeclaration where AccessorDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type AnonymousType seen as AnonymousType, which can write AnonymousType \| undefined where GenericType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type AnonymousType \| DeferredTypeReference seen as AnonymousType, which can write AnonymousType \| undefined where GenericType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type AnonymousType \| GenericType seen as AnonymousType, which can write AnonymousType \| undefined where GenericType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type AssignmentPattern \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type AutoAccessorPropertyDeclaration \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type AwaitedTypeInstantiation seen as Type \| undefined, which can write Symbol \| undefined where Symbol is read | 1 | 1 | 1 | 1 | 0 |
| a value of type AwaitedTypeInstantiation \| Type seen as Type, which can write Symbol \| undefined where Symbol is read | 1 | 1 | 1 | 1 | 0 |
| a value of type BigIntLiteral \| ComputedPropertyName \| Identifier \| JsxNamespacedName \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral seen as Type, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type BigIntLiteral \| ComputedPropertyName \| Identifier \| JsxNamespacedName \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type BigIntLiteral \| ComputedPropertyName \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Block \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type CallExpression \| undefined seen as NodeArray<Expression> \| undefined, whose readonly field transformFlags becomes writable: a readonly field may hold something narrower than TransformFlags, which a write of TransformFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type CallLikeExpression \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ClassDeclaration \| ClassExpression \| InferTypeNode \| InterfaceDeclaration \| JSDocCallbackTag \| ... 4 more ... \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Declaration seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Declaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 6 | 6 | 6 | 6 | 0 |
| a value of type Declaration \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Declaration[] seen as (ArrayTypeNode \| Declaration \| NodeWithTypeArguments \| TupleTypeNode)[], which can write ArrayTypeNode \| Declaration \| NodeWithTypeArguments \| TupleTypeNode where Declaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Declaration[] seen as Node[], which can write Node where Declaration is read | 7 | 7 | 7 | 7 | 0 |
| a value of type Diagnostic seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithDetachedLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithLocation seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 78 | 78 | 78 | 78 | 0 |
| a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where SourceFile is read | 53 | 53 | 53 | 53 | 0 |
| a value of type DiagnosticWithLocation \| undefined seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 4 | 4 | 4 | 4 | 0 |
| a value of type ElementAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type EntityNameExpression \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ExportDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Expression \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 5 | 5 | 5 | 5 | 0 |
| a value of type Expression \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type FlowNode seen as FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 5 | 5 | 5 | 5 | 0 |
| a value of type FlowNode seen as FlowNode[] \| FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 2 | 2 | 2 | 2 | 0 |
| a value of type FlowNode \| undefined seen as false \| FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 6 | 6 | 6 | 6 | 0 |
| a value of type GenericType \| ResolvedType seen as ObjectType, which can write SymbolTable \| undefined where SymbolTable is read | 3 | 3 | 3 | 3 | 0 |
| a value of type GetAccessorDeclaration \| MethodDeclaration \| PropertyAssignment \| SetAccessorDeclaration \| ShorthandPropertyAssignment \| SpreadAssignment \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type IndexInfo \| undefined seen as IndexSignatureDeclaration \| undefined, which can write number \| undefined where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IndexInfo[] seen as IndexInfo[], which can write IndexInfo where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IndexSignatureDeclaration seen as Mutable<IndexSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type IndexSignatureDeclaration[] seen as Declaration[] \| undefined, which can write Declaration where IndexSignatureDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IntrinsicType[] seen as TypeParameter[], which can write TypeParameter where IntrinsicType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IntrinsicType[] \| undefined seen as TypeParameter[] \| undefined, which can write TypeParameter where IntrinsicType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IntrinsicType[] \| undefined seen as TypeParameter[], which can write TypeParameter where IntrinsicType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IterationTypes[] seen as (IterationTypes \| undefined)[], which can write IterationTypes \| undefined where IterationTypes is read | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocSignature \| SignatureDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocTypeExpression \| undefined seen as Signature \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SignatureFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JsxAttributeLike \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<__String, TransientSymbol> seen as SymbolTable \| undefined, which can write Symbol where TransientSymbol is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, [VariableDeclarationList, VariableDeclaration[]]> seen as Map<string, [CatchClause \| VariableDeclarationList, VariableDeclaration[]]>, which can write [CatchClause \| VariableDeclarationList, VariableDeclaration[]] where [VariableDeclarationList, VariableDeclaration[]] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type MemberName \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type MultiMap<__String, [Type, number]> seen as Map<__String, (number \| Type)[][]>, which can write (number \| Type)[][] where [Type, number][] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Node seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node \| undefined seen as NodeLinks \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of NodeCheckFlags would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Node \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NodeArray<Node> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NodeBuilderContext seen as SyntacticTypeNodeBuilderContext, which can write Required<Pick<SymbolTracker, "reportInferenceFallback">> where SymbolTrackerImpl is read | 11 | 11 | 11 | 11 | 0 |
| a value of type NumberLiteralType seen as LiteralType, which can write string \| number \| PseudoBigInt where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ParameterDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ParameterPropertyDeclaration | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyAccessExpression \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyName \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ResolvedType seen as ObjectType, which can write SymbolTable \| undefined where SymbolTable is read | 17 | 17 | 17 | 17 | 0 |
| a value of type ResolvedType \| TypeReference seen as ObjectType, which can write SymbolTable \| undefined where SymbolTable is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Set<__String> seen as Set<__String \| undefined>, which can write __String \| undefined where __String is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Signature \| undefined seen as Type \| undefined, which can write TypeFlags where SignatureFlags is read | 3 | 3 | 3 | 3 | 0 |
| a value of type SourceFile \| undefined seen as NodeLinks \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of NodeCheckFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type SourceFile \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type StringLiteral seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type StringLiteralType seen as LiteralType, which can write string \| number \| PseudoBigInt where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type StringLiteralType[] seen as Type[], which can write Type where StringLiteralType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Symbol seen as Declaration \| undefined, which can write number \| undefined where number is read | 2 | 2 | 2 | 2 | 0 |
| a value of type Symbol seen as Type \| undefined, which can write TypeFlags where SymbolFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Symbol seen as Type, which can write TypeFlags where SymbolFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Symbol \| undefined seen as Declaration \| undefined, which can write number \| undefined where number is read | 3 | 3 | 3 | 3 | 0 |
| a value of type Symbol \| undefined seen as GenericType \| undefined, which can write TypeFlags where SymbolFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Symbol \| undefined seen as Node \| undefined, which can write number \| undefined where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Symbol \| undefined seen as SourceFile \| undefined, which can write number \| undefined where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Symbol \| undefined seen as Type \| undefined, which can write TypeFlags where SymbolFlags is read | 10 | 10 | 10 | 10 | 0 |
| a value of type Symbol[] seen as Symbol[], which can write Symbol where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type SyntheticExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type T seen as T, a type parameter whose constraint EntityNameOrEntityNameExpression can be written, so it can write what T can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type T seen as T, a type parameter whose constraint Node can be written, so it can write what T can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type TransientSymbol[] seen as Symbol[], which can write Symbol where TransientSymbol is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Type seen as TypeNode, which can write number \| undefined where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Type \| undefined seen as Expression \| undefined, which can write number \| undefined where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Type \| undefined seen as JSDocTypeExpression \| undefined, which can write number \| undefined where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Type \| undefined seen as Signature \| undefined, which can write SignatureFlags where TypeFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Type \| undefined seen as Symbol \| undefined, which can write SymbolFlags where TypeFlags is read | 4 | 4 | 4 | 4 | 0 |
| a value of type Type \| undefined seen as TypeNode \| undefined, which can write number \| undefined where number is read | 3 | 3 | 3 | 3 | 0 |
| a value of type TypeCheckerHost seen as Program, whose readonly field redirectTargetsMap becomes writable: a readonly field may hold something narrower than RedirectTargetsMap, which a write of MultiMap<Path, string> would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type TypeNode \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 7 | 7 | 7 | 7 | 0 |
| a value of type TypeParameterDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type TypeVariable[] seen as (Type \| undefined)[], which can write Type \| undefined where TypeVariable is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Type[] seen as (Type \| undefined)[], which can write Type \| undefined where Type is read | 1 | 1 | 1 | 1 | 0 |
| a value of type [() => Type, __String][] seen as (readonly [() => Type, __String])[], which can write readonly [() => Type, __String] where [() => Type, __String] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type [node: Node] seen as unknown[], which can write unknown where Node is read | 2 | 2 | 2 | 2 | 0 |
| a value of type false \| FlowNode \| undefined seen as false \| FlowAssignment \| FlowLabel \| FlowReduceLabel \| FlowStart \| FlowSwitchClause \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 2 | 2 | 2 | 2 | 0 |
| a value of type never[] seen as BaseType[], which can write BaseType where never is read | 14 | 14 | 14 | 14 | 0 |
| a value of type never[] seen as ClassElement[], which can write ClassElement where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as Declaration[], which can write Declaration where never is read | 2 | 2 | 2 | 2 | 0 |
| a value of type never[] seen as IndexInfo[], which can write IndexInfo where never is read | 7 | 7 | 7 | 7 | 0 |
| a value of type never[] seen as Signature[], which can write Signature where never is read | 8 | 8 | 8 | 8 | 0 |
| a value of type never[] seen as Symbol[] \| undefined, which can write Symbol where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as Symbol[], which can write Symbol where never is read | 12 | 12 | 12 | 12 | 0 |
| a value of type never[] seen as TypeParameter[] \| undefined, which can write TypeParameter where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as TypeParameter[], which can write TypeParameter where never is read | 3 | 3 | 3 | 3 | 0 |
| a value of type never[] seen as Type[] \| undefined, which can write Type where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as Type[], which can write Type where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as VarianceFlags[], which can write VarianceFlags where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type readonly TypeParameter[] \| undefined seen as TypeParameterDeclaration[] \| undefined, which can write TypeParameterDeclaration where TypeParameter is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { introducesError: boolean; node: Identifier \| PropertyAccessEntityNameExpression; sym?: never; } \| { introducesError: boolean; node: Identifier \| PropertyAccessEntityNameExpression; sym: Symbol \| undefined; } seen as { introducesError: boolean; node: LeftHandSideExpression; }, which can write LeftHandSideExpression where Identifier \| PropertyAccessEntityNameExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { noInferenceFallback?: boolean \| undefined; enclosingDeclaration: ModuleDeclaration; enclosingFile: SourceFile \| undefined; flags: NodeBuilderFlags; ... 28 more ...; out: WriterContextOut; } seen as NodeBuilderContext, which can write Node \| undefined where ModuleDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| an argument to a rest parameter that isn't an array, which T's type argument can be inferred from, judged against its constraint unknown[] | 1 | 1 | 1 | 1 | 0 |
| an object refinement using an open numeric enum as a literal tag | 429 | 429 | 324 | 324 | 0 |
| an unproven predicate argument for parameter callback (argument "isFunctionOrSourceFile") | 3 | 3 | 3 | 3 | 0 |
| an unproven predicate argument for parameter callback (argument "l => isImportTypeNode(l) \|\| isExportDeclaration(l) \|\| isImportDeclaration(l) \|\| isJSDocImportTag(l)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter callback (argument "n => isInterfaceDeclaration(n) \|\| isTypeAliasDeclaration(n) \|\| isTypeLiteralNode(n)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter f (argument "d => isClassDeclaration(d) \|\| isClassExpression(d)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter nodeTest (argument "(n): n is FunctionDeclaration \| VariableDeclaration => isFunctionDeclaration(n) \|\| isVariableDeclaration(n)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter nodeTest (argument "isDeclaration") | 2 | 2 | 2 | 2 | 0 |
| an unproven predicate argument for parameter nodeTest (argument "isExportSpecifier") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter nodeTest (argument "isFunctionDeclaration") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter nodeTest (argument "isPropertyAccessExpression") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter predicate (argument "d => isVariableDeclaration(d)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter predicate (argument "m => m.kind === SyntaxKind.ExportKeyword") | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from Declaration to NamedDeclaration: optional field name has no proven compatible presence/type | 2 | 2 | 2 | 2 | 0 |
| an unproven relation from DiagnosticRelatedInformation to Diagnostic: optional field reportsUnnecessary has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from InstantiableType \| UnionOrIntersectionType to TypeParameter: optional field constraint has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from NodeArray<BindingElement> \| NodeArray<Expression> \| NodeArray<ArrayBindingElement> to NodeArray<Node>: optional field concat.parameter.element.slice.element.propertyName has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from ObjectType to AnonymousType: optional field target has no proven compatible presence/type | 2 | 2 | 2 | 2 | 0 |
| an unproven relation from StructuredType to AnonymousType: optional field target has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from StructuredType to ObjectType: optional field members has no proven compatible presence/type | 3 | 3 | 3 | 3 | 0 |
| an unproven relation from Type to SyntheticDefaultModuleType: optional field syntheticType has no proven compatible presence/type | 2 | 2 | 2 | 2 | 0 |
| an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | 24 | 24 | 24 | 24 | 0 |
| an unproven relation from Type to TypeVariable: optional field constraint has no proven compatible presence/type | 2 | 2 | 2 | 2 | 0 |
| an unproven relation from UnionOrIntersectionType to UnionType: optional field resolvedReducedType has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot ElementFlags.Optional | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot ModifierFlags.None | 5 | 5 | 5 | 5 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Block | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExportAssignment | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExpressionWithTypeArguments | 2 | 2 | 2 | 2 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Identifier | 2 | 2 | 2 | 2 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.JSDocTypeExpression | 7 | 7 | 7 | 7 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.MethodSignature \| SyntaxKind.MethodDeclaration \| SyntaxKind.Constructor \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.CallSignature \| SyntaxKind.ConstructSignature \| ... 6 more ... \| SyntaxKind.JSDocFunctionType | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Parameter | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.QuestionToken | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.SourceFile | 7 | 7 | 7 | 7 | 0 |
| an unproven value assigned to a numeric literal or enum member slot Ternary | 2 | 2 | 2 | 2 | 0 |
| an unproven value assigned to a numeric literal or enum member slot Ternary.False | 1 | 1 | 1 | 1 | 0 |
| optional property constraint in TypeParameter absent from structural source InterfaceType, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property constraint in TypeParameter absent from structural source Type, which can hide fields | 12 | 12 | 12 | 12 | 0 |
| optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property id in ArrayBindingPattern absent from structural source never, which can hide fields | 3 | 3 | 3 | 3 | 0 |
| optional property id in BigIntLiteral absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property id in Expression absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property id in JSDocThisTag absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property id in LateBoundName absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property id in NamedExports absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property id in Node absent from structural source never, which can hide fields | 46 | 46 | 46 | 46 | 0 |
| optional property id in PrivateIdentifier absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property localSymbol in Declaration absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property reportsUnnecessary in Diagnostic absent from structural source DiagnosticRelatedInformation, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property skipLogging in ErrorOutputContainer absent from structural source { errors?: Diagnostic[] \| undefined; }, which can hide fields | 6 | 6 | 6 | 6 | 0 |
| this outside a method | 1 | 1 | 1 | 1 | 0 |
| writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds | 3 | 3 | 3 | 3 | 0 |
| writing a possibly absent optional own field | 1 | 1 | 1 | 1 | 0 |

## Raw Findings

Moved by reason:


New findings:

None.

Units whose first stop moved:

None.

Panics: 0 before, 0 after.

## Distinct Reads

Moved by reason:


New findings:

None.

Units whose first stop moved:

None.

Panics: 0 before, 0 after.

The JSON retains every lowered site, move, new finding, unit stop and panic.
