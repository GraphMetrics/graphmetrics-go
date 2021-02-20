package signature

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
)

func TestTransforms_DropUnusedOperations(t *testing.T) {
	operation := `
query FirstOperation {
	field(id: "1", x: 1, y: 2.3)	
}
query SecondOperation {
	field(id: "1", x: 1, y: 2.3)	
}
`
	expected := `query SecondOperation {
	field(id: "1", x: 1, y: 2.3)
}
`

	_, document, _ := givenOperation(operation)
	dropUnusedOperations(document, "SecondOperation")

	actual := prettyPrint(document)
	assert.Equal(t, expected, actual)
}

func TestTransforms_HideLiterals(t *testing.T) {
	operation := `
fragment Test on Query {
	secondField(input: { id: "2" })
}
query {
	a: field(id: "1", x: 1, y: 2.3)
	b: field(id: $id, x: $x, y: $y)
	...Test
	thirdField(input: [{ id: "3" }])
	fourthField {
		fieldTwo(id: "5")
	}
}
`
	expected := `query {
	a: field(id: "", x: 0, y: 0)
	b: field(id: $id, x: $x, y: $y)
	... Test
	thirdField(input: [])
	fourthField {
		fieldTwo(id: "")
	}
}
fragment Test on Query {
	secondField(input: {})
}
`

	schema, document, events := givenOperation(operation)
	events.OnValue(hideLiterals)
	validator.Walk(schema, document, events)

	actual := prettyPrint(document)
	assert.Equal(t, expected, actual)
}

func TestTransforms_SeenFragments(t *testing.T) {
	operation := `
fragment Test on MyInterface {
	fieldTwo(id: "5")	
}
query {
	fourthField {
		... {
			fieldOne
		}
		... Test
		... on MyType {
			fieldThree
		}
	}
}
`
	schema, document, events := givenOperation(operation)
	seenFragments, detectFragmentUsage := fragmentUsed()
	events.OnFragmentSpread(detectFragmentUsage)
	validator.Walk(schema, document, events)

	assert.Equal(t, seenFragments, map[string]bool{"Test": true})
}

func TestTransforms_DropUnusedFragments(t *testing.T) {
	operation := `
fragment Test on Query {
	field(id: "1", x: 1, y: 2.3)	
}
fragment Unused on Query {
	secondField(input: $input)
}
query {
	...Test
}
`
	expected := `query {
	... Test
}
fragment Test on Query {
	field(id: "1", x: 1, y: 2.3)
}
`

	_, document, _ := givenOperation(operation)
	dropUnusedFragments(document, map[string]bool{"Test": true})

	actual := prettyPrint(document)
	assert.Equal(t, expected, actual)
}

func givenOperation(operation string) (*ast.Schema, *ast.QueryDocument, *validator.Events) {
	schema, _ := validator.LoadSchema(&ast.Source{Input: schema})
	document, _ := parser.ParseQuery(&ast.Source{Input: operation})
	events := &validator.Events{}
	return schema, document, events
}

var schema = `
scalar String
scalar Int
scalar ID
scalar Float

input MyFilter {
	n: Int
}

interface MyInterface {
	fieldThree: Float
}

type MyType implements MyInterface{
	fieldOne: String
	fieldTwo(id: ID): Int
	fieldThree: Float
}

input MyInput {
	id: ID
}

type Query {
	field(id: ID, x: Int, y: Float): String
	secondField(input: MyInput): ID
	thirdField(input: [ID]): Int
	fourthField: MyInterface
}

schema {
	query: Query
}
`
