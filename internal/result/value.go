package result

import (
	"fmt"
	"reflect"
	"strconv"
)

// Kind is the normalized type of a Redis reply.
type Kind int

const (
	KindNil Kind = iota
	KindString
	KindInt
	KindFloat
	KindBool
	KindArray
	KindMap
)

// Pair is one key/value entry in a map-like reply.
type Pair struct {
	Key   string
	Value Value
}

// Value is a terminal-friendly representation of a RESP reply.
type Value struct {
	Kind  Kind
	Text  string
	Int   int64
	Float float64
	Bool  bool
	Items []Value
	Pairs []Pair
}

func Normalize(raw any) Value {
	switch value := raw.(type) {
	case nil:
		return Value{Kind: KindNil}
	case string:
		return Value{Kind: KindString, Text: value}
	case []byte:
		return Value{Kind: KindString, Text: string(value)}
	case bool:
		return Value{Kind: KindBool, Bool: value}
	case int:
		return Value{Kind: KindInt, Int: int64(value)}
	case int8:
		return Value{Kind: KindInt, Int: int64(value)}
	case int16:
		return Value{Kind: KindInt, Int: int64(value)}
	case int32:
		return Value{Kind: KindInt, Int: int64(value)}
	case int64:
		return Value{Kind: KindInt, Int: value}
	case uint:
		return Value{Kind: KindInt, Int: int64(value)}
	case uint8:
		return Value{Kind: KindInt, Int: int64(value)}
	case uint16:
		return Value{Kind: KindInt, Int: int64(value)}
	case uint32:
		return Value{Kind: KindInt, Int: int64(value)}
	case uint64:
		return Value{Kind: KindInt, Int: int64(value)}
	case float32:
		return Value{Kind: KindFloat, Float: float64(value)}
	case float64:
		return Value{Kind: KindFloat, Float: value}
	case error:
		return Value{Kind: KindString, Text: value.Error()}
	}

	rv := reflect.ValueOf(raw)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		items := make([]Value, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			items[i] = Normalize(rv.Index(i).Interface())
		}
		return Value{Kind: KindArray, Items: items}
	case reflect.Map:
		pairs := make([]Pair, 0, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			pairs = append(pairs, Pair{
				Key:   fmt.Sprint(iter.Key().Interface()),
				Value: Normalize(iter.Value().Interface()),
			})
		}
		return Value{Kind: KindMap, Pairs: pairs}
	default:
		return Value{Kind: KindString, Text: fmt.Sprint(raw)}
	}
}

func (v Value) Display() string {
	switch v.Kind {
	case KindNil:
		return "(nil)"
	case KindString:
		return v.Text
	case KindInt:
		return strconv.FormatInt(v.Int, 10)
	case KindFloat:
		return strconv.FormatFloat(v.Float, 'f', -1, 64)
	case KindBool:
		return strconv.FormatBool(v.Bool)
	case KindArray:
		return fmt.Sprintf("[%d items]", len(v.Items))
	case KindMap:
		return fmt.Sprintf("{%d fields}", len(v.Pairs))
	default:
		return ""
	}
}
