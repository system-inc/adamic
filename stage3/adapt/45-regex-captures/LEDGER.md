## parseRange whitespace split allocation (adaptation 45 amendment)

Adamic's Nov 10 05:40Z ruling, #b5w3ycg row 4: write `range.split(/\s+/)`
at the call, retaining the capture-free `as string[]` proof. This allocates a
new RegExp object per executed split where upstream shared one. `split` ignores
lastIndex, and the pattern has no flags, so no state carries between calls.
The shared declaration is removed because its only other occurrence was this
call; leaving it would violate upstream's noUnusedLocals.
