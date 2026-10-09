| ID | Origin file:line | Change | Failing rows |
|---|---|---|---|
| M01 | internal/regexp/matcher.go:426 | b.steps >= b.limit → b.steps > b.limit | TestMatcherStepLimitBoundary |
| M02 | internal/regexp/matcher.go:327 | stateful := p.flags.Global \|\| p.flags.Sticky → stateful := p.flags.Global && p.flags.Sticky | TestMatcherExecution family, TestMatcherOct6LoopsNode |
| M03 | internal/regexp/matcher.go:385 | r.LastIndex = uint64(state.pos) → r.LastIndex = uint64(state.pos + 1) | TestMatcherExecution family, TestMatcherOct6LoopsNode |
| M04 | internal/regexp/matcher.go:399 | r.LastIndex = 0 → r.LastIndex = 1 | TestMatcherExecution family, TestMatcherOct6LoopsNode |
| M05 | internal/regexp/matcher.go:450 | alternative.pc = i.y 			stack = append(stack, alternative) 			s.pc = i.x → alternative.pc = i.x 			stack = append(stack, alternative) 			s.pc = i.y | TestMatcherExecution family |
| M06 | internal/regexp/matcher.go:537 | for _, id := range i.clear { 					state.caps[2*id] = -1 					state.caps[2*id+1] = -1 				} → drop whole loop | TestMatcherExecution family |
| M07 | internal/regexp/matcher.go:497 | failed = ok == i.negative → failed = ok != i.negative | TestMatcherExecution family, TestMatcherOct6LoopsNode, TestMatcherPropertyProviderStrings |
| M08 | internal/regexp/matcher.go:218 | if !f.DotAll { → if f.DotAll { | TestMatcherExecution family |
| M09 | internal/regexp/matcher.go:544 | if i.greedy { → if !i.greedy { | TestMatcherExecution family, TestMatcherOct6LoopsNode |
| M10 | internal/regexp/matcher.go:589 | return rune(c), pos + 1, true → return rune(c), pos + 2, true | TestMatcherExecution family, TestMatcherOct6LoopsNode, TestMatcherPropertyProviderStrings, TestMatcherProviderSnapshot, TestMatcherStepLimitBoundary |
| M11 | internal/regexp/canonicalize.go:13 | if !f.IgnoreCase { → if f.IgnoreCase { | TestMatcherExecution family, TestMatcherOct6LoopsNode, TestMatcherStepLimit, TestMatcherStepLimitBoundary |
| M12 | internal/regexp/sets.go:175 | set.strings[index] = slices.Clone(text) → set.strings[index] = text | TestMatcherProviderSnapshot |
| M13 | internal/regexp/sets.go:285 | contains == intersection → contains != intersection | TestMatcherExecution family, TestMatcherPropertyProviderStrings |
| M14 | internal/regexp/sets.go:293 | return i < len(s.ranges) && s.ranges[i].From <= c → return i < len(s.ranges) && s.ranges[i].From < c | TestMatcherExecution family, TestMatcherOct6LoopsNode, TestMatcherPropertyProviderStrings, TestMatcherStepLimit, TestMatcherStepLimitBoundary |
| M15 | internal/regexp/parser.go:110 | if f.Unicode && f.UnicodeSets { → if f.Unicode && !f.UnicodeSets { | TestFlags, TestMatcherExecution family, TestMatcherOct6LoopsNode, TestMatcherPropertyProviderStrings, TestNodeAgreement, TestParse |
| M16 | internal/regexp/parser.go:752 | limit = 2 → limit = 3 | TestMatcherExecution family |
| M17 | internal/regexp/sets.go:71 | b[k].From - 1 → b[k].From | TestMatcherExecution family |
| M18 | internal/regexp/matcher.go:389 | if p.anchored \|\| p.flags.Sticky { → if p.anchored && p.flags.Sticky { | TestMatcherExecution family, TestMatcherOct6LoopsNode |
