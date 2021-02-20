package signature

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
)

func OperationSignature(schema *ast.Schema, operation string, operationName string) (string, error) {
	// Parse the query (force a string so we don't reuse an existing document)
	document, err := parser.ParseQuery(&ast.Source{Input: operation})
	if err != nil {
		return "", err
	}

	// Pre-walker
	dropUnusedOperations(document, operationName)

	// Walker
	seenFragments, detectFragmentUsage := fragmentUsed()
	events := &validator.Events{}
	events.OnValue(hideLiterals)
	events.OnFragmentSpread(detectFragmentUsage)
	validator.Walk(schema, document, events)

	// Post-walker
	dropUnusedFragments(document, seenFragments)
	return prettyPrint(document), nil
}

func OperationHash(operation string) string {
	hash := sha256.Sum256([]byte(operation))
	return hex.EncodeToString(hash[:])
}
