# output-symbol

Version 1 raw binder ancestry, extending this wave's separate `symbol-ancestry` protocol without changing it. Header: version, `output-symbol`, symbol-present. A present symbol has identity, raw symbol flags, declaration count. Each declaration has file, kind, byte start/end, raw node flags, declaration-file/default-library/external-module flags, ancestor count. Each immediate-parent-to-source ancestor has kind, textual name (empty for nontextual names), name kind, raw node flags, module keyword. Destructuring and computed names keep their kind without calling `Node.Text` on those AST shapes.

The default resolves aliases. `output-symbol\nown` preserves the original symbol. Other suffixes are refused. Native code checks platform identity, including the identifier spelling `NodeJS`; a string-named module with that text is a different declaration.
