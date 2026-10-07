package ast

// Oracle-only constructor for the typed nil value the Go binding guard accepts.
// A shallow-copied node preserves the parsed kind and parent; the original stays intact.
func (n *Node) AdamicNilBindingData() { n.data = (*BindingElement)(nil) }
