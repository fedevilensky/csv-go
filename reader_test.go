package csv_test

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"csv-go"
)

func TestUnmarshalReflection(t *testing.T) {
	type s struct {
		String     string
		Int        int
		Float      float64
		Complex    complex64
		Bool       bool
		Uint       uint
		PointerInt *int
	}
	file, err := os.Open("./testdata/s.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var res []s
	if err := csv.NewReader(file).ReadCSV(&res); err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatal("expected 2 rows, but got ", len(res))
	}
	t.Log(res[0])
	t.Log(res[1])
}

type sTagged struct {
	TaggedString  string    `csv:"String"`
	TaggedInt     int       `csv:"Int"`
	TaggedFloat   float64   `csv:"Float"`
	TaggedComplex complex64 `csv:"Complex"`
	TaggedBool    bool      `csv:"Bool"`
	TaggedUint    uint      `csv:"Uint"`
	PointerInt    *int      `csv:"_"`
}

func TestUnmarshalReflectionTagged(t *testing.T) {
	file, err := os.Open("./testdata/s.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var res []sTagged
	if err := csv.NewReader(file).ReadCSV(&res); err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatal("expected 2 rows, but got ", len(res))
	}

	for _, r := range res {
		if r.PointerInt != nil {
			t.Fatal("PointerInt should be nil")
		}
		if r.TaggedString == "" {
			t.Fatal("taggedString should not be empty")
		}
	}
}

type sUnmarshal struct {
	string     string
	int        int
	float      float64
	complex    complex64
	bool       bool
	uint       uint64
	pointerint *int64
}

func (s *sUnmarshal) UnmarshalCSV(values []string) error {
	s.string = values[0]
	s.int, _ = strconv.Atoi(strings.TrimSpace(values[1]))
	cmp, _ := strconv.ParseComplex(strings.TrimSpace(values[2]), 64)
	s.complex = complex64(cmp)
	s.bool, _ = strconv.ParseBool(strings.TrimSpace(values[3]))
	s.uint, _ = strconv.ParseUint(strings.TrimSpace(values[4]), 10, 64)
	pint, _ := strconv.ParseInt(strings.TrimSpace(values[5]), 10, 64)
	s.pointerint = &pint
	s.float, _ = strconv.ParseFloat(strings.TrimSpace(values[6]), 64)
	return nil
}

func (s *sUnmarshal) UnmarshalCSVWithHeader(values, attributeNames []string) error {
	for i, attributeName := range attributeNames {
		switch attributeName {
		case "String":
			s.string = values[i]
		case "Int":
			s.int, _ = strconv.Atoi(strings.TrimSpace(values[i]))
		case "Float":
			s.float, _ = strconv.ParseFloat(strings.TrimSpace(values[i]), 64)
		case "Complex":
			cmp, _ := strconv.ParseComplex(strings.TrimSpace(values[i]), 64)
			s.complex = complex64(cmp)
		case "Bool":
			s.bool, _ = strconv.ParseBool(strings.TrimSpace(values[i]))
		case "Uint":
			s.uint, _ = strconv.ParseUint(strings.TrimSpace(values[i]), 10, 64)
		case "PointerInt":
			pint, _ := strconv.ParseInt(strings.TrimSpace(values[i]), 10, 64)
			s.pointerint = &pint
		}
	}

	return nil
}

func TestUnmarshalWithHeader(t *testing.T) {
	file, err := os.Open("./testdata/s.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var res []sUnmarshal
	if err := csv.NewReader(file).ReadCSV(&res); err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatal("expected 2 rows, but got ", len(res))
	}

	for _, r := range res {
		if r.string == "" {
			t.Fatal("taggedString should not be empty")
		}
	}
}

func TestUnmarshalWithoutHeader(t *testing.T) {
	file, err := os.Open("./testdata/s_no_header.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var res []sUnmarshal
	csvReader := csv.NewReader(file)
	csvReader.WithHeader = false
	if err := csvReader.ReadCSV(&res); err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatal("expected 2 rows, but got ", len(res))
	}

	for _, r := range res {
		if r.string == "" {
			t.Fatal("taggedString should not be empty")
		}
	}
}

type sCustom struct {
	Custom custom
}

type custom struct {
	string string
}

func (c custom) Parse(s string) (interface{}, error) {
	return custom{string: s}, nil
}

func TestUnmarshalCustomStruct(t *testing.T) {
	file, err := os.Open("./testdata/s_custom.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var res []sCustom
	if err := csv.NewReader(file).ReadCSV(&res); err != nil {
		t.Fatal(err)
	}
	if len(res) != 3 {
		t.Fatal("expected 3 rows, but got ", len(res))
	}

	for _, r := range res {
		if r.Custom.string == "" {
			t.Fatal("taggedString should not be empty")
		}
	}
}
