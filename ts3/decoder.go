package ts3

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

// Decoder 负责将 TS3 的原始文本响应解码为 Go 结构体
type Decoder struct{}

// NewDecoder 创建一个新的解码器
func NewDecoder() *Decoder {
	return &Decoder{}
}

// Decode 将 response 字符串解析并填充到 v 指向的结构体或切片中
func (d *Decoder) Decode(response string, v interface{}) error {
	// 1. 验证 v 是否为指针
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("ts3: Decode requires a non-nil pointer")
	}

	// 获取指针指向的实际元素
	elem := rv.Elem()

	// 2. 解析原始字符串为 []map[string]string 结构
	// TS3 的数据是以 "|" 分隔行，空格分隔键值对
	maps := parseRawResponse(response)

	// 3. 根据目标类型进行填充
	switch elem.Kind() {
	case reflect.Slice:
		// 如果目标是切片 (e.g. []OnlineClient)
		return d.decodeSlice(maps, elem)
	case reflect.Struct:
		// 如果目标是单个结构体 (e.g. VersionInfo)
		// 如果返回了多行数据但目标只是单体，通常取第一行
		if len(maps) > 0 {
			return d.decodeStruct(maps[0], elem)
		}
		return nil
	default:
		return errors.New("ts3: Decode target must be a struct or a slice of structs")
	}
}

// parseRawResponse 将原始字符串拆解为键值对映射列表
func parseRawResponse(response string) []map[string]string {
	if len(response) == 0 {
		return nil
	}

	// 1. 按 "|" 分隔多行数据
	lines := strings.Split(response, "|")
	result := make([]map[string]string, 0, len(lines))

	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		rowMap := make(map[string]string)
		// 2. 按空格分隔属性
		parts := strings.Split(line, " ")
		for _, part := range parts {
			// 3. 按 "=" 分隔键和值
			kv := strings.SplitN(part, "=", 2)
			key := kv[0]
			val := ""
			if len(kv) > 1 {
				val = Unescape(kv[1]) // 关键：在此处反转义
			}
			rowMap[key] = val
		}
		result = append(result, rowMap)
	}
	return result
}

// decodeSlice
func (d *Decoder) decodeSlice(maps []map[string]string, sliceVal reflect.Value) error {
	elemType := sliceVal.Type().Elem()

	for _, dataMap := range maps {
		newElem := reflect.New(elemType).Elem()

		if err := d.decodeStruct(dataMap, newElem); err != nil {
			return err
		}
		sliceVal.Set(reflect.Append(sliceVal, newElem))
	}
	return nil
}

// decodeStruct
func (d *Decoder) decodeStruct(data map[string]string, structVal reflect.Value) error {
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)

		tag := fieldType.Tag.Get("ts3")
		if tag == "" {
			continue
		}

		valStr, ok := data[tag]
		if !ok {
			continue
		}

		if err := setField(field, valStr); err != nil {
			return err
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

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value == "" {
			value = "0"
		}
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value == "" {
			value = "0"
		}
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)

	case reflect.Bool:
		field.SetBool(value == "1")

	case reflect.Float32, reflect.Float64:
		if value == "" {
			value = "0.0"
		}
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)

	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.Int {
			if value == "" {
				return nil
			}
			strParts := strings.Split(value, ",")
			slice := reflect.MakeSlice(field.Type(), len(strParts), len(strParts))
			for i, part := range strParts {
				valInt, err := strconv.Atoi(part)
				if err != nil {
					return err
				}
				slice.Index(i).SetInt(int64(valInt))
			}
			field.Set(slice)
		}
	}
	return nil
}
