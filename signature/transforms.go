package signature

import (
	"bytes"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/formatter"
	"github.com/vektah/gqlparser/v2/validator"
)

func dropUnusedOperations(document *ast.QueryDocument, operationName string) {
	k := 0
	for _, o := range document.Operations {
		if o.Name == operationName {
			document.Operations[k] = o
			k++
		}
	}
	if k > 0 {
		document.Operations = document.Operations[:k]
	}
}

func dropUnusedFragments(document *ast.QueryDocument, seenFragments map[string]bool) {
	k := 0
	for _, f := range document.Fragments {
		if seenFragments[f.Name] {
			document.Fragments[k] = f
			k++
		}
	}
	if k > 0 {
		document.Fragments = document.Fragments[:k]
	}
}

func hideLiterals(walker *validator.Walker, value *ast.Value) {
	switch value.Kind {
	case ast.IntValue:
		value.Raw = "0"
		break
	case ast.FloatValue:
		value.Raw = "0"
		break
	case ast.StringValue:
		value.Raw = ""
		break
	case ast.ListValue:
		value.Children = []*ast.ChildValue{}
		break
	case ast.ObjectValue:
		value.Children = []*ast.ChildValue{}
		break
	}
}

func fragmentUsed() (map[string]bool, func(walker *validator.Walker, fragmentSpread *ast.FragmentSpread)) {
	seenFragments := map[string]bool{}
	return seenFragments, func(walker *validator.Walker, fragmentSpread *ast.FragmentSpread) {
		seenFragments[fragmentSpread.Name] = true
	}
}

func prettyPrint(document *ast.QueryDocument) string {
	var buf bytes.Buffer
	formatter.NewFormatter(&buf).FormatQueryDocument(document)
	return buf.String()
}
