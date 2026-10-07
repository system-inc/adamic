# symbol-ancestry

Raw binder metadata; no rule decisions are returned. The default query follows an import alias, while `symbol-ancestry\nown` retains the alias itself. Other suffixes fail. It registers with one physical switch line; no shared generator or test harness changes are required.

Fields use the existing UTF-16-length framing. Version 1, mode `symbol-ancestry`, and a presence flag begin the response. If present, symbol identity, symbol flags and declaration count follow. Each declaration supplies filename, node kind, start/end, declaration-file flag, default-library flag, external-module flag and ancestor count. Each ancestor supplies kind, name text, node flags and raw module keyword (such as `GlobalKeyword`, `NamespaceKeyword` or `ModuleKeyword`; empty for a nonmodule node). Parent order is immediate parent through SourceFile.

The native decoder is `symbol_ancestry.a`. The timer rule requests raw aliases because production Go does not follow a timer import. The Go test holds global augmentation and the two alias modes against checker declarations. Disabling suffix refusal and alias resolution independently fails that test.
