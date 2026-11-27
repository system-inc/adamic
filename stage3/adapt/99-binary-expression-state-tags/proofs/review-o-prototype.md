# Review-o's reported prototype-accessor observation

Source: review-o, read 13 of the round-3 head, “Decision needed” in the
review quoted in the round-4 brief.

> In Node an inherited numeric setter can still intercept a push. A VM-local prototype observer forwarding writes saw the original function and adapted string. When it substituted an unknown state for the incoming tag, the original completed and the adapted loop reached the VM timeout.

Adamic's Nov 26 ruling on #bbssdcy puts accessors installed on
`Array.prototype` outside the ordinary-array contract. This witness
documents the excluded observation and timeout. The dispatch stays
as ruled.

The original probe and its two raw outputs sit beside this file:
`array-prototype-probe.patch`, `prototype-observe.txt` and
`prototype-corrupt.txt`, byte for byte as review-o read 13 left them.
This file records the observation they show.
