package lower

// resolve and relative can throw while reading cwd; join and dirname are lexical for proven strings.
func nodePathThrows(member string) bool { return member == "resolve" || member == "relative" }
