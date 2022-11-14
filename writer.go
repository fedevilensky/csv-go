package csv

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

var (
	ErrNoCommasAllowedInHeader = errors.New("no commas allowed in header")
	ErrNoCommasAllowedInBody   = errors.New("no commas allowed in body")
)

type writer struct {
	writer     *bufio.Writer
	WithHeader bool
	Comma      rune
	UseCRLF    bool
}

type Stringer interface {
	String() string
}

type HeaderMarshaler interface {
	// Header should return a slice of strings, which will be used as the header of the CSV file.
	// No comma allowed, comma being what you set in the writer (by default ',').
	Header() []string
}

type BodyMarshaler interface {
	// MarshalCSV should return a slice of strings, where each string is a value.
	// No Comma is allowed, comma being what you set in the writer (by default ',').
	MarshalCSV() []string
}

// Creates a new writer from an io.Writer. Default separator is ',', default
// UseCRLF is false, and default WithHeader is true.
func NewWriter(w io.Writer) *writer {
	return &writer{
		writer:     bufio.NewWriter(w),
		Comma:      ',',
		WithHeader: true,
	}
}

func (w writer) WriteCSV(arr []interface{}) error {
	var EOL string
	if w.UseCRLF {
		EOL = "\r\n"
	} else {
		EOL = "\n"
	}

	if w.WithHeader {
		if err := w.writeHeader(arr); err != nil {
			return err
		}
		if _, err := w.writer.WriteString(EOL); err != nil {
			return err
		}
	}

	for _, elem := range arr {
		if err := w.writeElem(elem); err != nil {
			return err
		}
		if _, err := w.writer.WriteString(EOL); err != nil {
			return err
		}
	}

	return nil
}

func (w writer) writeElem(e interface{}) error {
	if mar, ok := e.(BodyMarshaler); ok {
		for i, value := range mar.MarshalCSV() {
			if i > 0 {
				if _, err := w.writer.WriteRune(w.Comma); err != nil {
					return err
				}
			}
			if strings.ContainsRune(value, w.Comma) {
				return ErrNoCommasAllowedInBody
			}
			if _, err := w.writer.WriteString(value); err != nil {
				return err
			}
		}
		return nil
	}

	elemType := reflect.TypeOf(e)
	elem := reflect.ValueOf(e)
	if elemType.Kind() == reflect.Pointer {
		elemType = elemType.Elem()
		elem = elem.Elem()
	}
	fields := reflect.VisibleFields(elemType)
	for i, field := range fields {
		if strings.ContainsRune(field.Name, w.Comma) {
			return ErrNoCommasAllowedInBody
		}
		str, err := getString(elem.FieldByName(field.Name))
		if err != nil {
			return err
		}
		if _, err := w.writer.WriteString(str); err != nil {
			return err
		}
		if i < len(fields)-1 {
			if _, err := w.writer.WriteRune(w.Comma); err != nil {
				return err
			}
		}
	}
	return nil
}

func getString(v reflect.Value) (string, error) {
	// if v has String method, will use it to get the string
	if stringer, ok := v.Interface().(Stringer); ok {
		return stringer.String(), nil
	}
	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprint(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprint(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return fmt.Sprint(v.Float()), nil
	case reflect.Bool:
		return fmt.Sprint(v.Bool()), nil
	case reflect.Pointer:
		return getString(v.Elem())
	}
	return "", errors.New("unsupported type")
}

func (w writer) writeHeader(arr []interface{}) error {
	elem := reflect.New(reflect.TypeOf(arr).Elem())
	if mar, ok := elem.Interface().(HeaderMarshaler); ok {
		headerNames := mar.Header()
		for i, name := range headerNames {
			if i > 0 {
				if _, err := w.writer.WriteRune(w.Comma); err != nil {
					return err
				}
			}
			if strings.ContainsRune(name, w.Comma) {
				return ErrNoCommasAllowedInHeader
			}
			if _, err := w.writer.WriteString(name); err != nil {
				return err
			}

		}
	} else {
		i := 0
		for _, field := range reflect.VisibleFields(elem.Type()) {
			if field.Anonymous {
				continue
			}
			if i > 0 {
				if _, err := w.writer.WriteRune(w.Comma); err != nil {
					return err
				}
			}
			if _, err := w.writer.WriteString(field.Name); err != nil {
				return err
			}
			i++
		}
	}
	return nil
}

// WriteCSVElems is the same as WriteCSV, but you do not have to pass a slice. This is useful
// when you just have a couple of values.
func (w writer) WriteCSVElems(elems ...interface{}) error {
	return w.WriteCSV(elems)
}
