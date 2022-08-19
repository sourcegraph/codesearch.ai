package parsinghelpers

import sitter "github.com/smacker/go-tree-sitter"

func IsIdentifierType(nodeType string) bool {
	return nodeType == "name" ||
		nodeType == "identifier" ||
		nodeType == "property_identifier" ||
		nodeType == "constant" ||
		nodeType == "field_identifier" ||
		nodeType == "type_identifier"
}

func FindNamedIdentifier(node *sitter.Node, code []byte) string {
	if node == nil {
		return ""
	}

	identifier := ""
	for i := 0; i < int(node.NamedChildCount()); i++ {
		namedChildNode := node.NamedChild(i)
		namedChildNodeType := namedChildNode.Type()
		if IsIdentifierType(namedChildNodeType) {
			identifier = namedChildNode.Content(code)
			break
		}
	}
	return identifier
}
