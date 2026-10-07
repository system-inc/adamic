The Node host uses the single @types/node 25.3.3 installation at
stage3/api/node_modules/@types/node. Install stage3/api's locked dependencies
before checking sources that statically import node:* modules. No declarations
are copied, generated or vendored into internal/load. A missing installation or
wrong version is a loader error. Ordinary Adamic sources keep the existing prelude.

Host lowerers identify declarations with load.IsNodeLibrary and register only
implemented runtime members with lower.RegisterNodeLibraryMembers during init.
Names are node:<declaring module>.<member>, for example node:fs.statSync,
node:fs.size and node:fs.isFile. Import aliases resolve to their declarations.
Unregistered runtime uses fail up front with NotYet naming the member, including
function values, properties and constructors. Type-only references are allowed.
Registration does not implement a member; its unit must supply the lowering and
backend runtime. Unsupported overloads still need a named NotYet in that lowerer.
