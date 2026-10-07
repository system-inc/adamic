package lower

// resolve can throw while reading cwd; join and dirname are purely lexical for proven strings.
func nodePathThrows(member string) bool { return member == "resolve" }
