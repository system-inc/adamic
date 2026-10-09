| ID | Origin registry.go line | Change | Failing rows |
|---|---:|---|---|
| M1 | 52 | `if module != "" { => if false {` | TestAdamicRuleModule |
| M2 | 125 | `decoder.DisallowUnknownFields() => // decoder.DisallowUnknownFields()` | TestDescriptorRejections |
| M3 | 136 | `names[d.Name] = true => names[d.Name] = false` | TestDescriptorRejections |
| M4 | 137 | `if adapters[d.Oracle] { => if false {` | TestDuplicateOracleAdapter |
| M5 | 168 | `if d.Visit == "" \|\| len(d.Kinds) == 0 { => if d.Visit == "" && len(d.Kinds) == 0 {` | TestDescriptorRejections |
| M6 | 178 | `if !identifier.MatchString(kind) \|\| seen[kind] \|\| !kinds[kind] { => if !identifier.MatchString(kind) \|\| seen[kind] {` | TestDescriptorRejections |
| M7 | 193 | `if adapter.Name.Name != "main" \|\| !functions[d.Oracle] \|\| !functions[d.Oracle+"Options"] { => if adapter.Name.Name != "main" {` | TestDescriptorRejections |
| M8 | 226 | `if a.Order != b.Order { => if a.Order == b.Order {` |  |
| M9 | 310 | `arguments += ", parent" => arguments += ", index"` |  |
| M10 | 245 | `if d.Module == "" { => if d.Module != "" {` | TestAdamicRuleModule |
| M11 | 350 | `err == nil && bytes.Equal(existing, file.data) => err == nil && !bytes.Equal(existing, file.data)` | TestDeterministicRegeneration |
| M12 | 92 | `sort.Strings(paths) => // sort.Strings(paths)` |  |
| P1 | 335 | `empty Generate` | TestDuplicateOracleAdapter, TestDeterministicRegeneration, TestAdamicRuleModule, TestDescriptorRejections |
| P2 | 240 | `empty Render` | TestAdamicRuleModule |
