package csv_test

import (
	"csv-go"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type sHeaderBody struct {
	String     string
	Int        int
	Float      float64
	Complex    complex64
	Bool       bool
	Uint       uint
	PointerInt *int
}

func (s *sHeaderBody) MarshalCSV() []string {
	return []string{
		fmt.Sprint(s.String),
		fmt.Sprint(s.Int),
		fmt.Sprint(s.Float),
		// fmt.Sprint(s.Complex),
		fmt.Sprint(s.Bool),
		fmt.Sprint(s.Uint),
		fmt.Sprint(*s.PointerInt),
	}
}

func (s *sHeaderBody) Header() []string {
	return []string{
		"Custom Header String",
		"Custom Header Int",
		"Custom Header Float",
		// "Custom Header Complex",
		"Custom Header Bool",
		"Custom Header Uint",
		"Custom Header PointerInt",
	}
}

func TestWriterCustomHeaderAndMarshaller(t *testing.T) {
	strBuf := strings.Builder{}
	w := csv.NewWriter[*sHeaderBody](&strBuf)
	elems := []*sHeaderBody{{
		String:     "string",
		Int:        1,
		Float:      1.1,
		Complex:    1.1 + 1.1i,
		Bool:       true,
		Uint:       1,
		PointerInt: new(int),
	}}

	w.WriteCSV(elems)
	result := strBuf.String()

	assert.Equal(t,
		"Custom Header String,Custom Header Int,Custom Header Float,Custom Header Bool,Custom Header Uint,Custom Header PointerInt\n"+
			"string,1,1.1,true,1,0\n",
		result,
	)
}

func TestArbitraryStruct(t *testing.T) {
	type s struct {
		String     string
		Int        int
		Float      float64
		complex    complex64
		Bool       bool
		Uint       uint
		PointerInt *int
	}
	strBuf := strings.Builder{}
	w := csv.NewWriter[*s](&strBuf)
	elems := []*s{{
		String:     "string",
		Int:        1,
		Float:      1.1,
		complex:    1.1 + 1.1i,
		Bool:       true,
		Uint:       1,
		PointerInt: new(int),
	}}

	w.WriteCSV(elems)
	result := strBuf.String()

	assert.Equal(t,
		"String,Int,Float,Bool,Uint,PointerInt\n"+
			"string,1,1.1,true,1,0\n",
		result,
	)
}

func TestStructWithTags(t *testing.T) {
	type s struct {
		String     string `csv:"string"`
		Int        int    `csv:"weird tag"`
		Float      float64
		complex    complex64
		Bool       bool
		Uint       uint
		PointerInt *int `csv:"-"`
	}
	strBuf := strings.Builder{}
	w := csv.NewWriter[*s](&strBuf)
	elems := []*s{{
		String:     "string",
		Int:        1,
		Float:      1.1,
		complex:    1.1 + 1.1i,
		Bool:       true,
		Uint:       1,
		PointerInt: new(int),
	}}

	w.WriteCSV(elems)
	result := strBuf.String()

	assert.Equal(t,
		"string,weird tag,Float,Bool,Uint\n"+
			"string,1,1.1,true,1\n",
		result,
	)
}

// TODO: WIP
// func TestMap(t *testing.T) {
// 	strBuf := strings.Builder{}
// 	w := csv.NewWriter[map[string]any](&strBuf)
// 	elems := []map[string]any{{
// 		"String":     "string",
// 		"Int":        1,
// 		"Float":      1.1,
// 		"complex":    1.1 + 1.1i,
// 		"Bool":       true,
// 		"Uint":       1,
// 		"PointerInt": new(int),
// 	}}

// 	w.WriteCSV(elems)
// 	result := strBuf.String()

// 	assert.Equal(t,
// 		"Custom Header String,Custom Header Int,Custom Header Float,Custom Header Bool,Custom Header Uint,Custom Header PointerInt\n"+
// 			"string,1,1.1,true,1,0\n",
// 		result,
// 	)
// }
