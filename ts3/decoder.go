package ts3

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Decoder converts TS3 ServerQuery response text into Go structs/slices.
type Decoder struct{}

// NewDecoder creates a Decoder instance.
func NewDecoder() *Decoder {
	return &Decoder{}
}

// Decode parses response into v, where v must be a non-nil pointer to struct
// or slice of structs.
func (d *Decoder) Decode(response string, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("ts3: Decode requires a non-nil pointer")
	}

	elem := rv.Elem()
	rows := parseRawResponse(response)

	switch elem.Kind() {
	case reflect.Struct:
		if len(rows) == 0 {
			return nil
		}
		return d.decodeStruct(rows[0], elem)
	case reflect.Slice:
		return d.decodeSlice(rows, elem)
	default:
		return errors.New("ts3: Decode target must be a struct or slice of structs")
	}
}

func parseRawResponse(response string) []map[string]string {
	response = strings.TrimSpace(response)
	if response == "" {
		return nil
	}

	lines := strings.Split(response, "|")
	out := make([]map[string]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		row := make(map[string]string)
		for _, item := range strings.Fields(line) {
			kv := strings.SplitN(item, "=", 2)
			key := kv[0]
			val := ""
			if len(kv) == 2 {
				val = Unescape(kv[1])
			}
			row[key] = val
		}
		out = append(out, row)
	}
	return out
}

func (d *Decoder) decodeSlice(rows []map[string]string, target reflect.Value) error {
	elemType := target.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return errors.New("ts3: Decode slice target must be []struct")
	}

	decoded := reflect.MakeSlice(target.Type(), 0, len(rows))
	for i, row := range rows {
		item := reflect.New(elemType).Elem()
		if err := d.decodeStruct(row, item); err != nil {
			return fmt.Errorf("ts3: decode row %d: %w", i, err)
		}
		decoded = reflect.Append(decoded, item)
	}
	target.Set(decoded)
	return nil
}

func (d *Decoder) decodeStruct(data map[string]string, target reflect.Value) error {
	t := target.Type()
	for i := 0; i < target.NumField(); i++ {
		field := target.Field(i)
		structField := t.Field(i)

		tag := structField.Tag.Get("ts3")
		if tag == "" {
			continue
		}

		raw, ok := data[tag]
		if !ok {
			continue
		}

		if err := setField(field, raw); err != nil {
			return fmt.Errorf("field %s (%s): %w", structField.Name, tag, err)
		}
	}
	return nil
}

func setField(field reflect.Value, value string) error {
	if !field.CanSet() {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value == "" {
			value = "0"
		}
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(n)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value == "" {
			value = "0"
		}
		n, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(n)
		return nil

	case reflect.Bool:
		switch strings.ToLower(value) {
		case "", "0", "false", "no", "off":
			field.SetBool(false)
			return nil
		case "1", "true", "yes", "on":
			field.SetBool(true)
			return nil
		default:
			return fmt.Errorf("invalid bool value %q", value)
		}

	case reflect.Float32, reflect.Float64:
		if value == "" {
			value = "0"
		}
		n, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(n)
		return nil

	case reflect.Slice:
		return setSliceField(field, value)
	}

	return nil
}

func setSliceField(field reflect.Value, value string) error {
	if value == "" {
		field.Set(reflect.MakeSlice(field.Type(), 0, 0))
		return nil
	}

	switch field.Type().Elem().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parts := strings.Split(value, ",")
		out := reflect.MakeSlice(field.Type(), 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			n, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				return err
			}
			item := reflect.New(field.Type().Elem()).Elem()
			item.SetInt(n)
			out = reflect.Append(out, item)
		}
		field.Set(out)
		return nil

	case reflect.String:
		parts := strings.Split(value, ",")
		out := reflect.MakeSlice(field.Type(), 0, len(parts))
		for _, part := range parts {
			item := reflect.New(field.Type().Elem()).Elem()
			item.SetString(strings.TrimSpace(part))
			out = reflect.Append(out, item)
		}
		field.Set(out)
		return nil
	}

	return nil
}
