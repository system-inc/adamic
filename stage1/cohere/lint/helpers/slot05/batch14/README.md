# Jump routing helper

`make_break.a` ports Go cohere's `Builder.makeBreak`. The generic node and block values retain identity. `JumpState` owns ordered jump records, each with a mutable broken flag. Current block reachability selects the early return. Exact label text selects the innermost matching target; an unlabeled break skips nonbreakable records, while a labeled break may reach one. The selected target is marked before calling link. Every reachable invocation calls makeUnreachable, including a reachable invocation with no matching target.

AST text, graph linking and allocation/unreachable transitions remain explicit external dependencies. The source does not implement those helpers, a full CFG, or a regex matcher. Go rune/byte offsets and diagnostic positions are irrelevant to this helper.

The private oracle overlays dependency call names in the actual pinned Go method. Wrappers call the real Go dependencies while recording order, arguments and broken flags visible during linking. The .a driver supplies callback effects from the Go observation and compares the helper's own routing, callback order, final current block and target state. It also keeps an old jump-list view and checks that broken marking remains visible there. Parser-generated label names plus nil, empty, unmatched, Unicode and repeated label controls are tested against depths zero through five, all breakability patterns, initial broken flags, nil destinations and loop aliases. Graph implementation parity is outside this contract.

Go cohere is pinned to 715ba94f3608a6500086b1076ce5cb7e51b836db. All four consumers' test-file string literals are read, including configuration and expected-text literals, parsed by the real parser, and used to derive label/node controls. This is bounded helper behavior, not a claim that these rules are ported. The Go method's assumptions require a present current block and a contiguous private jump stack. Arbitrary callback effects, missing array slots, aliasing the same mutable jump object into multiple stack slots, invalid UTF-8 and lone surrogate labels are not covered.

pushJump and popJump were withdrawn after a concurrent ownership refresh proved slot 03's earlier claim. Their preliminary checks are not delivered or counted. One helper is retained in this directory; two replacement slots await the branch's full landing gate and publication.

```
source /workspace/adamic-tools/env.sh
ADAMIC_SLOT05_BATCH14_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch14/evidence" go test ./stage1/cohere/lint/helpers/slot05/batch14 -count=1 -v -timeout=20m > /tmp/lint05-batch14-retained.log 2>&1
```
