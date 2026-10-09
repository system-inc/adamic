// Package boundaries holds cohere's port of eslint-plugin-boundaries, which checks which parts of a
// codebase may import which.
//
// Only `boundaries/dependencies` is here, and only for the option shapes a config this tree lints
// actually writes. Base, the backend framework under api-phi-health, uses it twice: once to keep its
// own layers pointing one way (the command line may reach api and foundation, foundation may reach
// api, api reaches nothing), and once to stop a project module importing a worker. Every other shape
// the upstream schema accepts is refused by name when the config is read, so a config can never load
// clean while asking for a decision this port does not make.
//
// The authority is the installed build, eslint-plugin-boundaries 7.2.0 with @boundaries/elements
// 3.1.1 in api-phi-health's node_modules. Neither ships its test corpus, so the fixtures here were
// taken by driving that build through ESLint with Base's own configuration on planted imports at real
// paths in api-phi-health, and every verdict in `dependencies_test.go` is one that build returned.
package boundaries
