#!/usr/bin/env python3
"""Instrument an existing Adamic entry overlay for #rgrtt32."""
import json
from pathlib import Path
import sys

repository, overlay = (Path(arg).resolve() for arg in sys.argv[1:])
manifest = overlay / "overlay.json"
metadata = json.loads(manifest.read_text())
replacements = metadata["Replace"]


def edit(name, changes):
    original = repository / "internal/lower" / name
    source = Path(replacements.get(str(original), original)).read_text()
    for before, after in changes:
        if source.count(before) != 1:
            raise ValueError(f"{name}: expected one occurrence of {before!r}")
        source = source.replace(before, after)
    target = overlay / f"rgrtt32_{name}"
    target.write_text(source)
    replacements[str(original)] = str(target)


edit("class.go", [
    ("""func (l *lowering) setProperty(target *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {""",
     """func (l *lowering) setProperty(target *ast.Node, valueNode *ast.Node) (statements []ir.Statement, err error) {
    if l.traceRgrtt32Site(target) {
        defer func() { l.traceRgrtt32Store(target, statements, err) }()
    }"""),

])

edit("interface_cast.go", [
    ("""func (l *lowering) judgeViewAccess(part *ast.Node, known bool) (bool, error) {""",
     """func (l *lowering) judgeViewAccess(part *ast.Node, known bool) (checked bool, failure error) {
    if l.traceRgrtt32Site(part) {
        defer func() { l.traceRgrtt32Judgment(part, known, checked, failure) }()
    }"""),
])

edit("diagnostics.go", [
    ("""return &NotYet{Where: l.program.Where(node), What: what}""",
     """failure := &NotYet{Where: l.program.Where(node), What: what}
    if what == "writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds" && l.traceRgrtt32Site(node) {
        l.traceRgrtt32Refusal(node, failure)
    }
    return failure"""),
])

edit("latent_hook.go", [
    ("""\t\t\tlatentFindingOwner, latentFindingPhase = program.Where(node), "lowering"
\t\t\tunit :=""",
     """\t\t\tlatentFindingOwner, latentFindingPhase = program.Where(node), "lowering"
    if !traceRgrtt32Select(node) {
        continue
    }
    traceRgrtt32Writes = nil
    \t\t\tunit :="""),
    ("""l.latentFinishUnit(&unit, node)""",
     """l.traceRgrtt32Completed(node)
    l.latentFinishUnit(&unit, node)"""),
])

helper = overlay / "rgrtt32_probe.go"
helper.write_text(r"""
package lower

import (
    "fmt"
    "os"
    "runtime/debug"
    "strings"

    "github.com/microsoft/TypeScript/tsc/shim/ast"
    "github.com/system-inc/adamic/internal/ir"
)

var traceRgrtt32Enabled = os.Getenv("ADAMIC_RGRRT32_TRACE") == "1"
var traceRgrtt32Where = os.Getenv("ADAMIC_RGRRT32_WHERE")
var traceRgrtt32Writes []ir.SetProperty
var traceRgrtt32Target *ast.Node

func traceRgrtt32Select(node *ast.Node) bool {
    selected := os.Getenv("ADAMIC_RGRRT32_UNITS")
    if !traceRgrtt32Enabled || selected == "" {
        return true
    }
    name := node.Name()
    return name != nil && strings.Contains(","+selected+",", ","+name.Text()+",")
}

func (l *lowering) traceRgrtt32Site(node *ast.Node) bool {
    return traceRgrtt32Enabled && node != nil &&
        node.Kind == ast.KindPropertyAccessExpression &&
        node.Name().Text() == "valueDeclaration" &&
        strings.HasSuffix(l.program.Where(node), traceRgrtt32Location())
}

func traceRgrtt32Location() string {
    if traceRgrtt32Where != "" {
        return traceRgrtt32Where
    }
    return "/src/compiler/utilities.ts:4315:9"
}

func (l *lowering) traceRgrtt32Store(target *ast.Node, statements []ir.Statement, err error) {
    where := l.program.Where(target)
    for _, statement := range statements {
        if store, ok := statement.(ir.SetProperty); ok {
            fmt.Fprintf(os.Stderr, "RGRRT32 store unit=%s site=%s ir=%#v checked=%t\n",
                latentFindingOwner, where, store, l.result.CheckedField(store.ViewReceiverTypeID, store.Name))
            traceRgrtt32Writes = append(traceRgrtt32Writes, store)
            return
        }
    }
    fmt.Fprintf(os.Stderr, "RGRRT32 store unit=%s site=%s ir=%#v error=%v\n",
        latentFindingOwner, where, statements, err)
}

func (l *lowering) traceRgrtt32Judgment(target *ast.Node, known, checked bool, failure error) {
    receiver := l.viewReceiver(l.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression))
    id := int(receiver.Id())
    fmt.Fprintf(os.Stderr, "RGRRT32 judgment unit=%s site=%s receiverTypeID=%d known=%t judgedChecked=%t checked=%t viewOrigins=%d error=%v\n",
        latentFindingOwner, l.program.Where(target), id, known, checked,
        l.result.CheckedField(id, target.Name().Text()), len(l.result.ViewOrigins), failure)
}

func (l *lowering) traceRgrtt32Refusal(target *ast.Node, failure *NotYet) {
    receiver := l.viewReceiver(l.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression))
    id := int(receiver.Id())
    fmt.Fprintf(os.Stderr, "RGRRT32 refusal unit=%s site=%s receiverTypeID=%d checked=%t ir=unavailable reach=unavailable viewOrigins=%d error=%v\n%s",
        latentFindingOwner, failure.Where, id, l.result.CheckedField(id, target.Name().Text()),
        len(l.result.ViewOrigins), failure, debug.Stack())
}

func (l *lowering) traceRgrtt32FindTarget() *ast.Node {
    if traceRgrtt32Target != nil {
        return traceRgrtt32Target
    }
    var visit func(*ast.Node) bool
    visit = func(node *ast.Node) bool {
        if l.traceRgrtt32Site(node) {
            traceRgrtt32Target = node
            return true
        }
        return node.ForEachChild(visit)
    }
    for _, source := range l.program.Files() {
        if !strings.HasSuffix(l.program.FileName(source), "/src/compiler/utilities.ts") {
            continue
        }
        for _, declaration := range source.Statements.Nodes {
            if declaration.Kind == ast.KindFunctionDeclaration && declaration.Name() != nil &&
                declaration.Name().Text() == "setValueDeclaration" && visit(declaration) {
                return traceRgrtt32Target
            }
        }
    }
    return nil
}

func (l *lowering) traceRgrtt32Completed(owner *ast.Node) {
    if !traceRgrtt32Enabled {
        return
    }
    if os.Getenv("ADAMIC_RGRRT32_UNITS") != "" {
        if target := l.traceRgrtt32FindTarget(); target != nil {
            receiver := l.viewReceiver(l.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression))
            id := int(receiver.Id())
            name := ""
            if owner.Name() != nil {
                name = owner.Name().Text()
            }
            fmt.Fprintf(os.Stderr, "RGRRT32 completed unit=%s name=%s site=%s stores=%d receiverTypeID=%d checked=%t viewOrigins=%d\n",
                latentFindingOwner, name, l.program.Where(target), len(traceRgrtt32Writes), id,
                l.result.CheckedField(id, target.Name().Text()), len(l.result.ViewOrigins))
        }
    }
    if len(traceRgrtt32Writes) == 0 {
        return
    }
    defer func() {
        if crash := recover(); crash != nil {
            fmt.Fprintf(os.Stderr, "RGRRT32 reach unit=%s probe-panic=%v\n", latentFindingOwner, crash)
        }
    }()
    // Allocation numbering rewrites IR, so do it on a disposable census snapshot.
    probe := l.latentSnapshot()
    graph, viewed, viewUnknown := probe.viewedAllocations(probe.result, true)
    for _, store := range traceRgrtt32Writes {
        reach := graph.ReachingAllocations(store.Object)
        intersection := []int{}
        for _, site := range reach.Sites {
            if viewed[site] {
                intersection = append(intersection, site)
            }
        }
        fmt.Fprintf(os.Stderr, "RGRRT32 reach unit=%s writeSite=%d receiver=%#v receiverSites=%v receiverUnknown=%t reasons=%q viewedReceiverSites=%v viewOrigins=%d viewOriginsUnknown=%t checked=%t\n",
            latentFindingOwner, store.Site, store.Object, reach.Sites, reach.Unknown, reach.Reasons,
            intersection, len(probe.result.ViewOrigins), viewUnknown,
            probe.result.CheckedField(store.ViewReceiverTypeID, store.Name))
    }
}
""")
replacements[str(repository / "internal/lower/rgrtt32_probe.go")] = str(helper)
manifest.write_text(json.dumps(metadata, indent=2) + "\n")
