package lower

import "github.com/microsoft/TypeScript/tsc/shim/checker"

// viewsReaching is the program-wide question "which checked views can put a value at this type": every view the
// lowering made (viewSources) whose target, or a member of it, is assignable to t by the checker, in the order the
// views were made. A checked view checks at a read, never at its cast, so a value it admits sits at t unread until
// something reads it. A view with no target to judge could put anything anywhere, so it counts as reaching every
// type, failing closed, since each admission leaning on this answer trusts what it leaves out; nothing builds one
// today. Asked once lowering is done, when every view is made: by checkArrayArmViews, for the arms Array.isArray's
// brand decides (#9z2ksv7), and by Workers' family 17 (#1atvfp5).
func (l *lowering) viewsReaching(t *checker.Type) []viewSource {
	var reaching []viewSource
	for _, view := range l.viewSources {
		if view.target == nil {
			reaching = append(reaching, view)
			continue
		}
		for _, member := range castMembers(view.target) {
			if l.checker.IsTypeAssignableTo(member, t) {
				reaching = append(reaching, view)
				break
			}
		}
	}
	return reaching
}
