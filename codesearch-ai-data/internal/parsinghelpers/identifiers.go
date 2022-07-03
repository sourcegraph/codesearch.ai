package parsinghelpers

import sitter "github.com/smacker/go-tree-sitter"

func FindNamedIdentifier(node *sitter.Node, code []byte) string {
	if node == nil {
		return ""
	}

	identifier := ""
	for i := 0; i < int(node.NamedChildCount()); i++ {
		namedChildNode := node.NamedChild(i)
		namedChildNodeType := namedChildNode.Type()
		if namedChildNodeType == "name" ||
			namedChildNodeType == "identifier" ||
			namedChildNodeType == "property_identifier" ||
			namedChildNodeType == "constant" ||
			namedChildNodeType == "field_identifier" {
			identifier = namedChildNode.Content(code)
			break
		}
	}
	return identifier
}
