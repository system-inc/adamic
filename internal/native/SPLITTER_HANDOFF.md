Stable emitter identity handoff to developer tools

No splitter or grouping implementation is changed on codex/stable-emitter-names.
The unchanged splitter at 14e8372816b1d3bf5bcc37ca57336e46d41b7a41 still gives
18 changed body/state units and 29 effective invalidations for adding a function.
Both temporary edits now give one changed unit and an unchanged shared header.

Required grouping change:

1. Group source functions by SourceIdentity.Module, not by encounter order or batches
   of sixteen. Use one persistent unit filename per relative module path, obtained
   from a deterministic escaped path plus digest. An added function joins that
   module's existing unit. Include its region variants and declaration-owned adapters.
2. The C emitter writes `// adamic-module "relative/path.a"` immediately before each
   source function definition and at each module boundary in main. The quoted value
   follows Go strconv.Quote syntax; use strconv.Unquote. An empty module denotes a
   synthesized helper. Do not infer module identity from a truncated readable symbol.
   IR users can use Function.Source and Program.MainModules directly.
3. Source-less library helpers use semantic content identities. Group these by the
   stable symbol (or an explicitly stable digest bucket), independently of their
   position in the function list. Their identical definitions are emitted once.
4. Separate main's marked ranges into module initialization units, plus a small
   coordinator retaining the original module order, adamic_start, final global
   releases, and return. The uncovered IR Main prefix is forwarder initialization.
   Marked ranges currently share a scope: check that extracted locals and cleanup
   have no cross-range lifetime before moving them. Fail explicitly if unsupported;
   never rearrange initialization or duplicate global state to obtain a cache hit.

Required dependency change:

1. Keep only the runtime/type ABI in a common header. Do not place every generated
   function prototype, global extern, cache, class, shape and literal there.
2. Compute each unit's referenced-symbol closure from identifier tokens, excluding
   literals/comments. Write only the prototypes/externs needed by that unit into its
   own declaration header. Resolve references from constant initializers and adapter
   bodies too. Preserve one external definition for every shared object.
3. Keep function-local caches with the owning function's module and module-main
   caches with that module. Globals belong to their source module. Place data used
   exclusively by one module there; keep genuinely shared content-addressed data in
   independently named units. Adding a declaration must not change other modules'
   declaration headers. Do not append all new data to one state.c that every edit
   recompiles. Preserve descriptor and closure pointer identity where the runtime
   observes it.
4. Hash each unit's actual preprocessed contents, referenced declarations, transitive
   runtime headers, ordered flags, compiler identity and cache format. An ABI or
   borrow/region/throw proof change must rebuild affected callers. Identity stability
   is not permission to retain an object whose generated body or dependencies changed.

Acceptance to add to developer tools:

Use the three grammar.ts overlays saved in stable_emitter_evidence/unit_changes_probe.go.txt.
Assert that each changes only the grammar module unit and leaves other unit inputs
and preprocessed keys identical. Include added/deleted units and changed headers in
that count. Reintroduce the whole-program temporary counter: the saved probe's
function-edit assertion currently fails with 18 changed units. Run Node/output,
counted and sanitized oracle parity in split and single-unit modes after changing
ownership, initialization extraction or dependency headers.
