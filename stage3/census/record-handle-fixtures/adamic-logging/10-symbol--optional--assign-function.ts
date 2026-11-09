// Logging-only adaptation for Adamic's console.log(message: string).
// Self-contained excerpts from microsoft/TypeScript tag v6.0.3.
// Only required Type/Symbol fields are retained. Dependency leaves are opaque:
// their contents do not affect rejection of a property-free function.
export {};
interface MapLike<T> { [index: string]: T; }
type SymbolId = number;
type __String = (string & { __escapedIdentifier: void; }) | (void & { __escapedIdentifier: void; }) | InternalSymbolName;
declare const enum InternalSymbolName {
    Call = "__call", // Call signatures
    Constructor = "__constructor", // Constructor implementations
    New = "__new", // Constructor signatures
    Index = "__index", // Index signatures
    ExportStar = "__export", // Module export * declarations
    Global = "__global", // Global self-reference
    Missing = "__missing", // Indicates missing symbol
    Type = "__type", // Anonymous type literal symbol
    Object = "__object", // Anonymous object literal declaration
    JSXAttributes = "__jsxAttributes", // Anonymous JSX attributes object literal declaration
    Class = "__class", // Unnamed class expression
    Function = "__function", // Unnamed function expression
    Computed = "__computed", // Computed property name declaration with dynamic name
    Resolving = "__resolving__", // Indicator symbol used to mark partially resolved type aliases
    ExportEquals = "export=", // Export assignment symbol
    Default = "default", // Default export symbol (technically not wholly internal, but included here for usability)
    This = "this",
    InstantiationExpression = "__instantiationExpression", // Instantiation expressions
    ImportAttributes = "__importAttributes",
}
declare const enum SymbolFlags {
    None                    = 0,
    FunctionScopedVariable  = 1 << 0,   // Variable (var) or parameter
    BlockScopedVariable     = 1 << 1,   // A block-scoped variable (let or const)
    Property                = 1 << 2,   // Property or enum member
    EnumMember              = 1 << 3,   // Enum member
    Function                = 1 << 4,   // Function
    Class                   = 1 << 5,   // Class
    Interface               = 1 << 6,   // Interface
    ConstEnum               = 1 << 7,   // Const enum
    RegularEnum             = 1 << 8,   // Enum
    ValueModule             = 1 << 9,   // Instantiated module
    NamespaceModule         = 1 << 10,  // Uninstantiated module
    TypeLiteral             = 1 << 11,  // Type Literal or mapped type
    ObjectLiteral           = 1 << 12,  // Object Literal
    Method                  = 1 << 13,  // Method
    Constructor             = 1 << 14,  // Constructor
    GetAccessor             = 1 << 15,  // Get accessor
    SetAccessor             = 1 << 16,  // Set accessor
    Signature               = 1 << 17,  // Call, construct, or index signature
    TypeParameter           = 1 << 18,  // Type parameter
    TypeAlias               = 1 << 19,  // Type alias
    ExportValue             = 1 << 20,  // Exported value marker (see comment in declareModuleMember in binder)
    Alias                   = 1 << 21,  // An alias for another symbol (see comment in isAliasSymbolDeclaration in checker)
    Prototype               = 1 << 22,  // Prototype property (no source representation)
    ExportStar              = 1 << 23,  // Export * declaration
    Optional                = 1 << 24,  // Optional property
    Transient               = 1 << 25,  // Transient symbol (created during type check)
    Assignment              = 1 << 26,  // Assignment treated as declaration (eg `this.prop = 1`)
    ModuleExports           = 1 << 27,  // Symbol for CommonJS `module` of `module.exports`
    All = -1,

    Enum = RegularEnum | ConstEnum,
    Variable = FunctionScopedVariable | BlockScopedVariable,
    Value = Variable | Property | EnumMember | ObjectLiteral | Function | Class | Enum | ValueModule | Method | GetAccessor | SetAccessor,
    Type = Class | Interface | Enum | EnumMember | TypeLiteral | TypeParameter | TypeAlias,
    Namespace = ValueModule | NamespaceModule | Enum,
    Module = ValueModule | NamespaceModule,
    Accessor = GetAccessor | SetAccessor,

    // Variables can be redeclared, but can not redeclare a block-scoped declaration with the
    // same name, or any other value that is not a variable, e.g. ValueModule or Class
    FunctionScopedVariableExcludes = Value & ~FunctionScopedVariable,

    // Block-scoped declarations are not allowed to be re-declared
    // they can not merge with anything in the value space
    BlockScopedVariableExcludes = Value,

    ParameterExcludes = Value,
    PropertyExcludes = None,
    EnumMemberExcludes = Value | Type,
    FunctionExcludes = Value & ~(Function | ValueModule | Class),
    ClassExcludes = (Value | Type) & ~(ValueModule | Interface | Function), // class-interface mergability done in checker.ts
    InterfaceExcludes = Type & ~(Interface | Class),
    RegularEnumExcludes = (Value | Type) & ~(RegularEnum | ValueModule), // regular enums merge only with regular enums and modules
    ConstEnumExcludes = (Value | Type) & ~ConstEnum, // const enums merge only with const enums
    ValueModuleExcludes = Value & ~(Function | Class | RegularEnum | ValueModule),
    NamespaceModuleExcludes = 0,
    MethodExcludes = Value & ~Method,
    GetAccessorExcludes = Value & ~SetAccessor,
    SetAccessorExcludes = Value & ~GetAccessor,
    AccessorExcludes = Value & ~Accessor,
    TypeParameterExcludes = Type & ~TypeParameter,
    TypeAliasExcludes = Type,
    AliasExcludes = Alias,

    ModuleMember = Variable | Function | Class | Interface | Enum | Module | TypeAlias | Alias,

    ExportHasLocal = Function | Class | Enum | ValueModule,

    BlockScoped = BlockScopedVariable | Class | Enum,

    PropertyOrAccessor = Property | Accessor,

    ClassMember = Method | Accessor | Property,

    /** @internal */
    ExportSupportsDefaultModifier = Class | Function | Interface,

    /** @internal */
    ExportDoesNotSupportDefaultModifier = ~ExportSupportsDefaultModifier,

    /** @internal */
    // The set of things we consider semantically classifiable.  Used to speed up the LS during
    // classification.
    Classifiable = Class | Enum | TypeAlias | Interface | TypeParameter | Module | Alias,

    /** @internal */
    LateBindingContainer = Class | Interface | TypeLiteral | ObjectLiteral | Function,
}
interface Symbol {
 flags: SymbolFlags;
 escapedName: __String;
 id: SymbolId;
 mergeId: number;
 constEnumOnlyModule: boolean | undefined;
}


// The value whose static receiver type is being tested is the record itself.
type Receiver = () => string | undefined;
type Holding = MapLike<Symbol>;
const h: Holding = {};
const source: Receiver = () => "called";
const assigned = Object.assign(h, source);
const receiver: Receiver = assigned;
console.log(Object.is(receiver, h) ? "same record: true" : "same record: false");
console.log("typeof receiver: " + typeof receiver);
try {
 receiver();
 console.log("call returned");
} catch (error) {
 if (!(error instanceof TypeError)) throw error;
 console.log("call threw: TypeError");
}
