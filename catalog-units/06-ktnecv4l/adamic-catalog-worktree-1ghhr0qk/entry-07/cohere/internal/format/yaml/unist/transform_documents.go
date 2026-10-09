package unist

// Ported from yaml-unist-parser 3.2.0, dist/transforms/: documents.mjs, document.mjs, document-head.mjs
// and document-body.mjs.

import (
	"fmt"

	"github.com/system-inc/cohere/internal/format/yaml/compose"
	"github.com/system-inc/cohere/internal/format/yaml/cst"
)

// documentData is the object transformDocuments collects for each document.
type documentData struct {
	tokensBeforeBody []*cst.Token
	cstNode          *cst.Token
	node             *compose.Document
	tokensAfterBody  []*cst.Token
	documentEnd      *cst.Token
}

// getPointText is utils/get-point-text.mjs.
func getPointText(at *point) string {
	return fmt.Sprintf("%d:%d", at.line, at.column)
}

// transformDocuments is upstream's transformDocuments: the stream's top-level tokens sorted into each
// document's tokens before its body (directives and the comments among them), its CST document and its
// end marker with the comments before it.
func (context *context) transformDocuments(parsedDocuments []*compose.Document, cstTokens []*cst.Token) []*Node {
	if len(parsedDocuments) == 0 {
		return []*Node{}
	}
	var documents []*documentData
	bufferComments := []*cst.Token{}
	tokensBeforeBody := []*cst.Token{}
	var currentDocumentData *documentData
	createDocumentData := func(cstNode *cst.Token) *documentData {
		data := &documentData{
			tokensBeforeBody: append(append([]*cst.Token{}, tokensBeforeBody...), bufferComments...),
			cstNode:          cstNode,
			tokensAfterBody:  []*cst.Token{},
		}
		// parsedDocuments[documents.length]: undefined past the end, which only an unexpected document
		// token reaches, and that throws first.
		if len(documents) < len(parsedDocuments) {
			data.node = parsedDocuments[len(documents)]
		}
		documents = append(documents, data)
		tokensBeforeBody = tokensBeforeBody[:0]
		bufferComments = bufferComments[:0]
		return data
	}
	for _, token := range tokens(cstTokens) {
		if token.Type == "document" {
			if len(documents) >= len(parsedDocuments) {
				throwError("Unexpected 'document' token at %s", getPointText(context.transformOffset(token.Offset)))
			}
			currentDocumentData = createDocumentData(token)
			continue
		}
		if token.Type == "comment" {
			bufferComments = append(bufferComments, token)
			continue
		}
		if token.Type == "directive" {
			tokensBeforeBody = append(tokensBeforeBody, bufferComments...)
			tokensBeforeBody = append(tokensBeforeBody, token)
			bufferComments = bufferComments[:0]
			continue
		}
		if token.Type == "doc-end" {
			if currentDocumentData == nil || currentDocumentData.documentEnd != nil {
				throwError("Unexpected 'doc-end' token at %s", getPointText(context.transformOffset(token.Offset)))
			}
			currentDocumentData.tokensAfterBody = append([]*cst.Token{}, bufferComments...)
			bufferComments = bufferComments[:0]
			currentDocumentData.documentEnd = token
			continue
		}
	}
	if len(tokensBeforeBody) > 0 {
		firstToken := tokensBeforeBody[0]
		throwError("Unexpected '%s' token at %s", firstToken.Type, getPointText(context.transformOffset(firstToken.Offset)))
	}
	if len(bufferComments) > 0 {
		if currentDocumentData == nil {
			currentDocumentData = createDocumentData(nil)
		}
		if len(bufferComments) > 0 {
			currentDocumentData.tokensAfterBody = append(currentDocumentData.tokensAfterBody, bufferComments...)
			bufferComments = bufferComments[:0]
		}
	}
	result := make([]*Node, len(documents))
	for index, document := range documents {
		result[index] = context.transformDocument(document)
	}
	return result
}

// transformDocument is upstream's transformDocument.
func (context *context) transformDocument(document *documentData) *Node {
	documentHead, docStart, tokensBeforeBody := context.transformDocumentHead(document.tokensBeforeBody, document.cstNode, document.node)
	documentBody, documentEndPoint, documentTrailingComment := context.transformDocumentBody(docStart, tokensBeforeBody,
		document.cstNode, document.node, document.tokensAfterBody, document.documentEnd)
	return context.createDocument(context.createPosition(context.position(documentHead).start, documentEndPoint), docStart != nil,
		document.documentEnd != nil, documentHead, documentBody, documentTrailingComment)
}

// transformDocumentHead is upstream's transformDocumentHead: the directives, the comments that end the
// head, and its trailing comment. It returns the --- token and the tokens between it and the body.
func (context *context) transformDocumentHead(tokensBeforeBody []*cst.Token, cstNode *cst.Token, document *compose.Document) (*Node, *cst.Token, []*cst.Token) {
	directives, endCommentCandidates := context.categorizeHeadNodes(tokensBeforeBody)
	betweenTokens := []*cst.Token{}
	var docStart *cst.Token
	if cstNode != nil {
		for _, token := range tokens(cstNode.Start) {
			betweenTokens = append(betweenTokens, token)
			if docStart == nil && token.Type == "doc-start" {
				for _, t := range betweenTokens {
					if t.Type == "comment" {
						comment := context.transformComment(t)
						endCommentCandidates = append(endCommentCandidates, comment)
					}
				}
				betweenTokens = []*cst.Token{}
				docStart = token
			}
		}
	}
	position := context.getHeadPosition(directives, document, docStart)
	var trailingComment *Node
	if docStart != nil && len(betweenTokens) > 0 {
		lastToken := betweenTokens[0]
		if lastToken.Type == "comment" {
			if context.transformOffset(lastToken.Offset).line == position.end.line {
				trailingComment = context.transformComment(lastToken)
				betweenTokens = betweenTokens[1:]
			}
		}
	}
	endComments := []*Node{}
	if docStart != nil {
		endComments = endCommentCandidates
	}
	return context.createDocumentHead(position, directives, endComments, trailingComment), docStart, betweenTokens
}

// categorizeHeadNodes is document-head.mjs's categorizeNodes: a comment on a directive's line is its
// trailing comment, and the comments after the last directive are candidates for the head's end
// comments.
func (context *context) categorizeHeadNodes(tokensBeforeBody []*cst.Token) (directives []*Node, endCommentCandidates []*Node) {
	directives = []*Node{}
	endCommentCandidates = []*Node{}
	var lastDirective *Node
	for _, token := range tokensBeforeBody {
		if token.Type == "comment" {
			node := context.transformComment(token)
			if lastDirective != nil && context.position(lastDirective).end.line == context.position(node).start.line &&
				lastDirective.TrailingComment == nil {
				lastDirective.TrailingComment = node
				context.position(lastDirective).end = context.position(node).end
			} else {
				endCommentCandidates = append(endCommentCandidates, node)
			}
		} else {
			node := context.transformDirective(token)
			directives = append(directives, node)
			lastDirective = node
			endCommentCandidates = []*Node{}
		}
	}
	return directives, endCommentCandidates
}

// getHeadPosition is document-head.mjs's getPosition.
func (context *context) getHeadPosition(directives []*Node, document *compose.Document, docStart *cst.Token) *position {
	var start, end int
	switch {
	case docStart != nil:
		start, end = docStart.Offset, docStart.Offset+len(docStart.Source)
	case document.Contents != nil:
		start, end = document.Contents.Range[0], document.Contents.Range[0]
	default:
		start, end = document.Range[0], document.Range[0]
	}
	if len(directives) != 0 {
		start = context.position(directives[0]).start.offset
	}
	return context.transformRange(start, end)
}

// transformDocumentBody is upstream's transformDocumentBody: the body, the point where the document
// ends, and the comment on the line of the ... marker.
func (context *context) transformDocumentBody(docStart *cst.Token, tokensBeforeBody []*cst.Token, cstNode *cst.Token, document *compose.Document, tokensAfterBody []*cst.Token, docEnd *cst.Token) (*Node, *point, *Node) {
	documentTrailingComment, endComments, propTokens := context.categorizeBodyNodes(tokensBeforeBody, cstNode, tokensAfterBody, docEnd)
	hasContent := false
	if contents := document.Contents; contents != nil {
		hasContent = contents.Range[0] < contents.Range[1]
		if !hasContent {
			for _, token := range propTokens {
				if token.Type == "tag" || token.Type == "anchor" {
					hasContent = true
					break
				}
			}
		}
	}
	var content *Node
	if hasContent {
		content = context.transformNode(document.Contents, propTokens)
	}
	if !hasContent {
		for _, token := range context.extractComments(propTokens) {
			throwError("Unexpected token type in empty document body: %s", token.Type)
		}
	}
	position, documentEndPoint := context.getBodyPosition(docStart, document, content, docEnd)
	return context.createDocumentBody(position, content, endComments), documentEndPoint, documentTrailingComment
}

// categorizeBodyNodes is document-body.mjs's categorizeNodes.
func (context *context) categorizeBodyNodes(tokensBeforeBody []*cst.Token, cstNode *cst.Token, tokensAfterBody []*cst.Token, docEnd *cst.Token) (documentTrailingComment *Node, endComments []*Node, propTokens []*cst.Token) {
	endComments = []*Node{}
	documentTrailingComments := []*Node{}
	propTokens = []*cst.Token{}
	for _, token := range tokensBeforeBody {
		if maybeContentPropertyToken(token) {
			propTokens = append(propTokens, token)
			continue
		}
		throwError("Unexpected token type: %s", token.Type)
	}
	for _, token := range context.extractComments(tokensAfterBody) {
		throwError("Unexpected token type: %s", token.Type)
	}
	var docEndPoint *point
	if docEnd != nil {
		docEndPoint = context.transformOffset(docEnd.Offset)
	}
	if cstNode != nil {
		// tokens(cstNode.end, docEnd?.end)
		var docEndEnd []*cst.Token
		if docEnd != nil {
			docEndEnd = docEnd.End
		}
		for _, token := range tokens(cstNode.End, docEndEnd) {
			if token.Type == "comment" {
				comment := context.transformComment(token)
				if docEndPoint != nil {
					if docEndPoint.line == context.position(comment).start.line {
						documentTrailingComments = append(documentTrailingComments, comment)
					} else if context.position(comment).start.line < docEndPoint.line {
						endComments = append(endComments, comment)
					}
				} else {
					endComments = append(endComments, comment)
				}
				continue
			}
			throwError("Unexpected token type: %s", token.Type)
		}
	}
	if len(documentTrailingComments) > 1 {
		throwError("Unexpected multiple document trailing comments at %s",
			getPointText(context.position(documentTrailingComments[1]).start))
	}
	// getLast(documentTrailingComments) || null
	if len(documentTrailingComments) > 0 {
		documentTrailingComment = documentTrailingComments[len(documentTrailingComments)-1]
	}
	return documentTrailingComment, endComments, propTokens
}

// getBodyPosition is document-body.mjs's getPosition.
func (context *context) getBodyPosition(docStart *cst.Token, document *compose.Document, content *Node, docEnd *cst.Token) (*position, *point) {
	var origEnd int
	if docEnd != nil {
		origEnd = max(0, docEnd.Offset-1)
	} else {
		// findCharIndex(context.text, document.range[2], /\S/u) ?? context.text.length
		origEnd = findCharIndex(context.text, document.Range[2], isNotWhitespace)
		if origEnd == -1 {
			origEnd = len(context.text)
		}
	}
	// context.text[origEnd - 1] === "\r"
	if origEnd-1 >= 0 && origEnd-1 < len(context.text) && context.text[origEnd-1] == '\r' {
		origEnd--
	}
	origStart := origEnd
	if content != nil {
		origStart = context.position(content).start.offset
	}
	if docStart != nil {
		docStartEnd := docStart.Offset + len(docStart.Source) + 1
		if origStart < docStartEnd && docStartEnd <= origEnd {
			origStart = docStartEnd
		}
	}
	position := context.transformRange(origStart, origEnd)
	documentEndPoint := position.end
	if docEnd != nil {
		documentEndPoint = context.transformOffset(docEnd.Offset + len(docEnd.Source))
	}
	return position, documentEndPoint
}
