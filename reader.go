package csv

import (
	"bufio"
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrCannotUnmarshalUnknownTypeWithoutHeader = errors.New("cannot unmarshal unknown type without header")

	errFieldNotFound = errors.New("field not found")
)

type reader struct {
	reader     *bufio.Reader
	WithHeader bool
	Comma      rune
	UseCRLF    bool
}

type Parser interface {
	// Parse string and return the a new instance of the struct with the parsed value
	Parse(string) (interface{}, error)
}

type UnmarshalerWithoutHeader interface {
	UnmarshalCSV(values []string) error
}

type Unmarshaler interface {
	UnmarshalCSVWithHeader(values, names []string) error
}

// Creates a new reader from an io.Reader. Default separator is ',',
// default UseCRLF is false, and default WithHeader is true.
func NewReader(r io.Reader) *reader {
	return &reader{
		reader:     bufio.NewReader(r),
		Comma:      ',',
		WithHeader: true,
	}
}

func (r reader) ReadCSV(arr interface{}) error {
	const EOL = '\n'
	var attrs []string
	sliceType := reflect.TypeOf(arr).Elem()
	result := reflect.MakeSlice(sliceType, 0, 20)
	elemType := sliceType.Elem()
	if r.WithHeader {

		header, err := r.reader.ReadString(EOL)
		if err != nil {
			return err
		}
		if r.UseCRLF {
			header = strings.TrimSuffix(header, "\r\n")
		} else {
			header = strings.TrimSuffix(header, "\n")
		}
		header = strings.TrimSuffix(header, string(r.Comma))
		attrs = strings.Split(header, string(r.Comma))
	}
	for {
		line, err := r.reader.ReadString(EOL)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return err
			}
		}
		if r.UseCRLF {
			line = strings.TrimSuffix(line, "\r\n")
		} else {
			line = strings.TrimSuffix(line, "\n")
		}
		line = strings.TrimSuffix(line, string(r.Comma))
		values := strings.Split(line, string(r.Comma))
		item := reflect.New(elemType)
		if !r.WithHeader {
			if unm, ok := item.Interface().(UnmarshalerWithoutHeader); ok {
				if err := unm.UnmarshalCSV(values); err != nil {
					return err
				}
			} else {
				return ErrCannotUnmarshalUnknownTypeWithoutHeader
			}
		}
		if unm, ok := item.Interface().(Unmarshaler); ok {
			if err := unm.UnmarshalCSVWithHeader(values, attrs); err != nil {
				return err
			}
		} else {
			fields := reflect.VisibleFields(elemType)
			for i, attr := range attrs {
				field, err := getField(fields, attr)
				if err != nil {
					if errors.Is(err, errFieldNotFound) {
						continue
					}
					return err
				}
				if err := setField(&item, field, values[i]); err != nil {
					return err
				}
			}
		}
		result = reflect.Append(result, item.Elem())
	}
	reflect.ValueOf(arr).Elem().Set(result)
	return nil
}

func getField(fields []reflect.StructField, attr string) (int, error) {
	// first look for tag
	for i, field := range fields {
		if field.Tag.Get("csv") == attr {
			return i, nil
		}
	}
	// then look for name
	for i, field := range fields {
		if field.Name == attr || strings.EqualFold(field.Name, attr) {
			// if field is tagged, we ignore the name
			if _, ok := field.Tag.Lookup("csv"); !ok {
				return i, nil
			}
		}
	}
	return -1, errFieldNotFound
}

func setField(item *reflect.Value, field int, value string) error {
	if parser, ok := item.Elem().Field(field).Interface().(Parser); ok {
		newItem, err := parser.Parse(value)
		if err != nil {
			return err
		}
		if item.Elem().Field(field).Type().Kind() == reflect.Pointer {
			item.Elem().Field(field).Elem().Set(reflect.ValueOf(newItem))
		} else {
			item.Elem().Field(field).Set(reflect.ValueOf(newItem))
		}
	}
	fieldKind := item.Elem().Field(field).Kind()
	isPointer := false
	if fieldKind == reflect.Pointer {
		isPointer = true
		fieldKind = item.Elem().Field(field).Elem().Kind()
	}
	switch fieldKind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = strings.TrimSpace(value)
		if val, err := strconv.ParseInt(value, 10, 64); err != nil {
			return err
		} else {
			if !isPointer {
				item.Elem().Field(field).SetInt(val)
			} else {
				item.Elem().Field(field).Set(reflect.ValueOf(val))
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = strings.TrimSpace(value)
		if val, err := strconv.ParseUint(value, 10, 64); err != nil {
			return err
		} else {
			if !isPointer {
				item.Elem().Field(field).SetUint(val)
			} else {
				item.Elem().Field(field).Set(reflect.ValueOf(val))
			}
		}
	case reflect.Bool:
		value = strings.TrimSpace(value)
		if val, err := strconv.ParseBool(value); err != nil {
			return err
		} else {
			if !isPointer {
				item.Elem().Field(field).SetBool(val)
			} else {
				item.Elem().Field(field).Set(reflect.ValueOf(val))
			}
		}
	case reflect.Float32, reflect.Float64:
		value = strings.TrimSpace(value)
		if val, err := strconv.ParseFloat(value, 64); err != nil {
			return err
		} else {
			if !isPointer {
				item.Elem().Field(field).SetFloat(val)
			} else {
				item.Elem().Field(field).Set(reflect.ValueOf(val))
			}
		}
	case reflect.Complex64, reflect.Complex128:
		value = strings.TrimSpace(value)
		if val, err := strconv.ParseComplex(value, 128); err != nil {
			return err
		} else {
			if !isPointer {
				item.Elem().Field(field).SetComplex(val)
			} else {
				item.Elem().Field(field).Set(reflect.ValueOf(val))
			}
		}
	case reflect.String:
		val := strings.TrimPrefix(strings.TrimSuffix(value, `"`), `"`)
		if !isPointer {
			item.Elem().Field(field).SetString(val)
		} else {
			item.Elem().Field(field).Set(reflect.ValueOf(val))
		}
	}
	return nil
}
