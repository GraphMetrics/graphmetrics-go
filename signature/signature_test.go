package signature

import (
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator"

	"github.com/stretchr/testify/assert"
)

func TestTransforms_StableSignature(t *testing.T) {
	operation := `
fragment Test on Query {
	field(id: "1", x: 1, y: 2.3)	
}
fragment OtherTest on Query {
	secondField(input: $input)
}
query MyQuery {
	...Test
}
query MyOtherQuery {
	...Test
}
`
	expected := `query MyQuery {
	... Test
}
fragment Test on Query {
	field(id: "", x: 0, y: 0)
}
`
	schema, _ := validator.LoadSchema(&ast.Source{Input: schema})
	signature, _ := OperationSignature(schema, operation, "MyQuery")

	assert.Equal(t, expected, signature)
}
