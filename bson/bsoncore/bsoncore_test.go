package bsoncore

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mongodb/mongo-go-driver/bson/bsontype"
	"github.com/mongodb/mongo-go-driver/bson/decimal"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
)

func TestAppend(t *testing.T) {
	bits := math.Float64bits(3.14159)
	pi := make([]byte, 8)
	binary.LittleEndian.PutUint64(pi, bits)

	testCases := []struct {
		name     string
		fn       interface{}
		params   []interface{}
		expected []byte
	}{
		{
			"AppendType",
			AppendType,
			[]interface{}{make([]byte, 0), bsontype.Null},
			[]byte{byte(bsontype.Null)},
		},
		{
			"AppendKey",
			AppendKey,
			[]interface{}{make([]byte, 0), "foobar"},
			[]byte{'f', 'o', 'o', 'b', 'a', 'r', 0x00},
		},
		{
			"AppendHeader",
			AppendHeader,
			[]interface{}{make([]byte, 0), bsontype.Null, "foobar"},
			[]byte{byte(bsontype.Null), 'f', 'o', 'o', 'b', 'a', 'r', 0x00},
		},
		{
			"AppendDouble",
			AppendDouble,
			[]interface{}{make([]byte, 0), float64(3.14159)},
			pi,
		},
		{
			"AppendDoubleElement",
			AppendDoubleElement,
			[]interface{}{make([]byte, 0), "foobar", float64(3.14159)},
			append([]byte{byte(bsontype.Double), 'f', 'o', 'o', 'b', 'a', 'r', 0x00}, pi...),
		},
		{
			"AppendString",
			AppendString,
			[]interface{}{make([]byte, 0), "barbaz"},
			[]byte{0x07, 0x00, 0x00, 0x00, 'b', 'a', 'r', 'b', 'a', 'z', 0x00},
		},
		{
			"AppendStringElement",
			AppendStringElement,
			[]interface{}{make([]byte, 0), "foobar", "barbaz"},
			[]byte{byte(bsontype.String),
				'f', 'o', 'o', 'b', 'a', 'r', 0x00,
				0x07, 0x00, 0x00, 0x00, 'b', 'a', 'r', 'b', 'a', 'z', 0x00,
			},
		},
		{
			"AppendDocument",
			AppendDocument,
			[]interface{}{[]byte{0x05, 0x00, 0x00, 0x00, 0x00}, []byte{0x05, 0x00, 0x00, 0x00, 0x00}},
			[]byte{0x05, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00},
		},
		{
			"AppendDocumentElement",
			AppendDocumentElement,
			[]interface{}{make([]byte, 0), "foobar", []byte{0x05, 0x00, 0x00, 0x00, 0x00}},
			[]byte{byte(bsontype.EmbeddedDocument),
				'f', 'o', 'o', 'b', 'a', 'r', 0x00,
				0x05, 0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			"AppendArray",
			AppendArray,
			[]interface{}{[]byte{0x05, 0x00, 0x00, 0x00, 0x00}, []byte{0x05, 0x00, 0x00, 0x00, 0x00}},
			[]byte{0x05, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00},
		},
		{
			"AppendArrayElement",
			AppendArrayElement,
			[]interface{}{make([]byte, 0), "foobar", []byte{0x05, 0x00, 0x00, 0x00, 0x00}},
			[]byte{byte(bsontype.Array),
				'f', 'o', 'o', 'b', 'a', 'r', 0x00,
				0x05, 0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			"AppendBinary Subtype2",
			AppendBinary,
			[]interface{}{make([]byte, 0), byte(0x02), []byte{0x01, 0x02, 0x03}},
			[]byte{0x07, 0x00, 0x00, 0x00, 0x02, 0x03, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03},
		},
		{
			"AppendBinaryElement Subtype 2",
			AppendBinaryElement,
			[]interface{}{make([]byte, 0), "foobar", byte(0x02), []byte{0x01, 0x02, 0x03}},
			[]byte{byte(bsontype.Binary),
				'f', 'o', 'o', 'b', 'a', 'r', 0x00,
				0x07, 0x00, 0x00, 0x00,
				0x02,
				0x03, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03,
			},
		},
		{
			"AppendBinary",
			AppendBinary,
			[]interface{}{make([]byte, 0), byte(0xFF), []byte{0x01, 0x02, 0x03}},
			[]byte{0x03, 0x00, 0x00, 0x00, 0xFF, 0x01, 0x02, 0x03},
		},
		{
			"AppendBinaryElement",
			AppendBinaryElement,
			[]interface{}{make([]byte, 0), "foobar", byte(0xFF), []byte{0x01, 0x02, 0x03}},
			[]byte{byte(bsontype.Binary),
				'f', 'o', 'o', 'b', 'a', 'r', 0x00,
				0x03, 0x00, 0x00, 0x00,
				0xFF,
				0x01, 0x02, 0x03,
			},
		},
		{
			"AppendUndefinedElement",
			AppendUndefinedElement,
			[]interface{}{make([]byte, 0), "foobar"},
			[]byte{byte(bsontype.Undefined), 'f', 'o', 'o', 'b', 'a', 'r', 0x00},
		},
		{
			"AppendObjectID",
			AppendObjectID,
			[]interface{}{
				make([]byte, 0),
				objectid.ObjectID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C},
			},
			[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C},
		},
		{
			"AppendObjectIDElement",
			AppendObjectIDElement,
			[]interface{}{
				make([]byte, 0), "foobar",
				objectid.ObjectID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C},
			},
			[]byte{byte(bsontype.ObjectID),
				'f', 'o', 'o', 'b', 'a', 'r', 0x00,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C,
			},
		},
		{
			"AppendBoolean (true)",
			AppendBoolean,
			[]interface{}{make([]byte, 0), true},
			[]byte{0x01},
		},
		{
			"AppendBoolean (false)",
			AppendBoolean,
			[]interface{}{make([]byte, 0), false},
			[]byte{0x00},
		},
		{
			"AppendBooleanElement",
			AppendBooleanElement,
			[]interface{}{make([]byte, 0), "foobar", true},
			[]byte{byte(bsontype.Boolean), 'f', 'o', 'o', 'b', 'a', 'r', 0x00, 0x01},
		},
		{
			"AppendDateTime",
			AppendDateTime,
			[]interface{}{make([]byte, 0), int64(256)},
			[]byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			"AppendDateTimeElement",
			AppendDateTimeElement,
			[]interface{}{make([]byte, 0), "foobar", int64(256)},
			[]byte{byte(bsontype.DateTime), 'f', 'o', 'o', 'b', 'a', 'r', 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			"AppendNullElement",
			AppendNullElement,
			[]interface{}{make([]byte, 0), "foobar"},
			[]byte{byte(bsontype.Null), 'f', 'o', 'o', 'b', 'a', 'r', 0x00},
		},
		{
			"AppendRegex",
			AppendRegex,
			[]interface{}{make([]byte, 0), "bar", "baz"},
			[]byte{'b', 'a', 'r', 0x00, 'b', 'a', 'z', 0x00},
		},
		{
			"AppendRegexElement",
			AppendRegexElement,
			[]interface{}{make([]byte, 0), "foobar", "bar", "baz"},
			[]byte{byte(bsontype.Regex),
				'f', 'o', 'o', 'b', 'a', 'r', 0x00,
				'b', 'a', 'r', 0x00, 'b', 'a', 'z', 0x00,
			},
		},
		{
			"AppendDBPointer",
			AppendDBPointer,
			[]interface{}{
				make([]byte, 0),
				"foobar",
				objectid.ObjectID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C},
			},
			[]byte{
				0x07, 0x00, 0x00, 0x00, 'f', 'o', 'o', 'b', 'a', 'r', 0x00,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C,
			},
		},
		{
			"AppendDBPointerElement",
			AppendDBPointerElement,
			[]interface{}{
				make([]byte, 0), "foobar",
				"barbaz",
				objectid.ObjectID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C},
			},
			[]byte{byte(bsontype.DBPointer),
				'f', 'o', 'o', 'b', 'a', 'r', 0x00,
				0x07, 0x00, 0x00, 0x00, 'b', 'a', 'r', 'b', 'a', 'z', 0x00,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C,
			},
		},
		{
			"AppendJavaScript",
			AppendJavaScript,
			[]interface{}{make([]byte, 0), "barbaz"},
			[]byte{0x07, 0x00, 0x00, 0x00, 'b', 'a', 'r', 'b', 'a', 'z', 0x00},
		},
		{
			"AppendJavaScriptElement",
			AppendJavaScriptElement,
			[]interface{}{make([]byte, 0), "foobar", "barbaz"},
			[]byte{byte(bsontype.JavaScript),
				'f', 'o', 'o', 'b', 'a', 'r', 0x00,
				0x07, 0x00, 0x00, 0x00, 'b', 'a', 'r', 'b', 'a', 'z', 0x00,
			},
		},
		{
			"AppendSymbol",
			AppendSymbol,
			[]interface{}{make([]byte, 0), "barbaz"},
			[]byte{0x07, 0x00, 0x00, 0x00, 'b', 'a', 'r', 'b', 'a', 'z', 0x00},
		},
		{
			"AppendSymbolElement",
			AppendSymbolElement,
			[]interface{}{make([]byte, 0), "foobar", "barbaz"},
			[]byte{byte(bsontype.Symbol),
				'f', 'o', 'o', 'b', 'a', 'r', 0x00,
				0x07, 0x00, 0x00, 0x00, 'b', 'a', 'r', 'b', 'a', 'z', 0x00,
			},
		},
		{
			"AppendCodeWithScope",
			AppendCodeWithScope,
			[]interface{}{[]byte{0x05, 0x00, 0x00, 0x00, 0x00}, "foobar", []byte{0x05, 0x00, 0x00, 0x00, 0x00}},
			[]byte{0x05, 0x00, 0x00, 0x00, 0x00,
				0x14, 0x00, 0x00, 0x00,
				0x07, 0x00, 0x00, 0x00, 'f', 'o', 'o', 'b', 'a', 'r', 0x00,
				0x05, 0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			"AppendCodeWithScopeElement",
			AppendCodeWithScopeElement,
			[]interface{}{make([]byte, 0), "foobar", "barbaz", []byte{0x05, 0x00, 0x00, 0x00, 0x00}},
			[]byte{byte(bsontype.CodeWithScope),
				'f', 'o', 'o', 'b', 'a', 'r', 0x00,
				0x14, 0x00, 0x00, 0x00,
				0x07, 0x00, 0x00, 0x00, 'b', 'a', 'r', 'b', 'a', 'z', 0x00,
				0x05, 0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			"AppendInt32",
			AppendInt32,
			[]interface{}{make([]byte, 0), int32(256)},
			[]byte{0x00, 0x01, 0x00, 0x00},
		},
		{
			"AppendInt32Element",
			AppendInt32Element,
			[]interface{}{make([]byte, 0), "foobar", int32(256)},
			[]byte{byte(bsontype.Int32), 'f', 'o', 'o', 'b', 'a', 'r', 0x00, 0x00, 0x01, 0x00, 0x00},
		},
		{
			"AppendTimestamp",
			AppendTimestamp,
			[]interface{}{make([]byte, 0), uint32(65536), uint32(256)},
			[]byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00},
		},
		{
			"AppendTimestampElement",
			AppendTimestampElement,
			[]interface{}{make([]byte, 0), "foobar", uint32(65536), uint32(256)},
			[]byte{byte(bsontype.Timestamp), 'f', 'o', 'o', 'b', 'a', 'r', 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00},
		},
		{
			"AppendInt64",
			AppendInt64,
			[]interface{}{make([]byte, 0), int64(4294967296)},
			[]byte{0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
		{
			"AppendInt64Element",
			AppendInt64Element,
			[]interface{}{make([]byte, 0), "foobar", int64(4294967296)},
			[]byte{byte(bsontype.Int64), 'f', 'o', 'o', 'b', 'a', 'r', 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
		{
			"AppendDecimal128",
			AppendDecimal128,
			[]interface{}{make([]byte, 0), decimal.NewDecimal128(4294967296, 65536)},
			[]byte{
				0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
			},
		},
		{
			"AppendDecimal128Element",
			AppendDecimal128Element,
			[]interface{}{make([]byte, 0), "foobar", decimal.NewDecimal128(4294967296, 65536)},
			[]byte{
				byte(bsontype.Decimal128), 'f', 'o', 'o', 'b', 'a', 'r', 0x00,
				0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
			},
		},
		{
			"AppendMaxKeyElement",
			AppendMaxKeyElement,
			[]interface{}{make([]byte, 0), "foobar"},
			[]byte{byte(bsontype.MaxKey), 'f', 'o', 'o', 'b', 'a', 'r', 0x00},
		},
		{
			"AppendMinKeyElement",
			AppendMinKeyElement,
			[]interface{}{make([]byte, 0), "foobar"},
			[]byte{byte(bsontype.MinKey), 'f', 'o', 'o', 'b', 'a', 'r', 0x00},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := reflect.ValueOf(tc.fn)
			if fn.Kind() != reflect.Func {
				t.Fatalf("fn must be of kind Func but is a %v", fn.Kind())
			}
			if fn.Type().NumIn() != len(tc.params) {
				t.Fatalf("tc.params must match the number of params in tc.fn. params %d; fn %d", fn.Type().NumIn(), len(tc.params))
			}
			if fn.Type().NumOut() != 1 || fn.Type().Out(0) != reflect.TypeOf([]byte{}) {
				t.Fatalf("fn must have one return parameter and it must be a []byte.")
			}
			params := make([]reflect.Value, 0, len(tc.params))
			for _, param := range tc.params {
				params = append(params, reflect.ValueOf(param))
			}
			results := fn.Call(params)
			got := results[0].Interface().([]byte)
			want := tc.expected
			if !bytes.Equal(got, want) {
				t.Errorf("Did not receive expected bytes. got %v; want %v", got, want)
			}
		})
	}
}

func TestRead(t *testing.T) {
	bits := math.Float64bits(3.14159)
	pi := make([]byte, 8)
	binary.LittleEndian.PutUint64(pi, bits)

	testCases := []struct {
		name     string
		fn       interface{}
		param    []byte
		expected []interface{}
	}{
		{
			"ReadType/not enough bytes",
			ReadType,
			[]byte{},
			[]interface{}{bsontype.Type(0), []byte{}, false},
		},
		{
			"ReadType/success",
			ReadType,
			[]byte{0x0A},
			[]interface{}{bsontype.Null, []byte{}, true},
		},
		{
			"ReadKey/not enough bytes",
			ReadKey,
			[]byte{},
			[]interface{}{"", []byte{}, false},
		},
		{
			"ReadKey/success",
			ReadKey,
			[]byte{'f', 'o', 'o', 'b', 'a', 'r', 0x00},
			[]interface{}{"foobar", []byte{}, true},
		},
		{
			"ReadHeader/not enough bytes (type)",
			ReadHeader,
			[]byte{},
			[]interface{}{bsontype.Type(0), "", []byte{}, false},
		},
		{
			"ReadHeader/not enough bytes (key)",
			ReadHeader,
			[]byte{0x0A, 'f', 'o', 'o'},
			[]interface{}{bsontype.Type(0), "", []byte{0x0A, 'f', 'o', 'o'}, false},
		},
		{
			"ReadHeader/success",
			ReadHeader,
			[]byte{0x0A, 'f', 'o', 'o', 'b', 'a', 'r', 0x00},
			[]interface{}{bsontype.Null, "foobar", []byte{}, true},
		},
		{
			"ReadDouble/not enough bytes",
			ReadDouble,
			[]byte{0x01, 0x02, 0x03, 0x04},
			[]interface{}{float64(0.00), []byte{0x01, 0x02, 0x03, 0x04}, false},
		},
		{
			"ReadDouble/success",
			ReadDouble,
			pi,
			[]interface{}{float64(3.14159), []byte{}, true},
		},
		{
			"ReadString/not enough bytes (length)",
			ReadString,
			[]byte{},
			[]interface{}{"", []byte{}, false},
		},
		{
			"ReadString/not enough bytes (value)",
			ReadString,
			[]byte{0x0F, 0x00, 0x00, 0x00},
			[]interface{}{"", []byte{0x0F, 0x00, 0x00, 0x00}, false},
		},
		{
			"ReadString/success",
			ReadString,
			[]byte{0x07, 0x00, 0x00, 0x00, 'f', 'o', 'o', 'b', 'a', 'r', 0x00},
			[]interface{}{"foobar", []byte{}, true},
		},
		{
			"ReadDocument/not enough bytes (length)",
			ReadDocument,
			[]byte{},
			[]interface{}{[]byte(nil), []byte{}, false},
		},
		{
			"ReadDocument/not enough bytes (value)",
			ReadDocument,
			[]byte{0x0F, 0x00, 0x00, 0x00},
			[]interface{}{[]byte(nil), []byte{0x0F, 0x00, 0x00, 0x00}, false},
		},
		{
			"ReadDocument/success",
			ReadDocument,
			[]byte{0x0A, 0x00, 0x00, 0x00, 0x0A, 'f', 'o', 'o', 0x00, 0x00},
			[]interface{}{[]byte{0x0A, 0x00, 0x00, 0x00, 0x0A, 'f', 'o', 'o', 0x00, 0x00}, []byte{}, true},
		},
		{
			"ReadArray/not enough bytes (length)",
			ReadArray,
			[]byte{},
			[]interface{}{[]byte(nil), []byte{}, false},
		},
		{
			"ReadArray/not enough bytes (value)",
			ReadArray,
			[]byte{0x0F, 0x00, 0x00, 0x00},
			[]interface{}{[]byte(nil), []byte{0x0F, 0x00, 0x00, 0x00}, false},
		},
		{
			"ReadArray/success",
			ReadArray,
			[]byte{0x08, 0x00, 0x00, 0x00, 0x0A, '0', 0x00, 0x00},
			[]interface{}{[]byte{0x08, 0x00, 0x00, 0x00, 0x0A, '0', 0x00, 0x00}, []byte{}, true},
		},
		{
			"ReadBinary/not enough bytes (length)",
			ReadBinary,
			[]byte{},
			[]interface{}{byte(0), []byte(nil), []byte{}, false},
		},
		{
			"ReadBinary/not enough bytes (subtype)",
			ReadBinary,
			[]byte{0x0F, 0x00, 0x00, 0x00},
			[]interface{}{byte(0), []byte(nil), []byte{0x0F, 0x00, 0x00, 0x00}, false},
		},
		{
			"ReadBinary/not enough bytes (value)",
			ReadBinary,
			[]byte{0x0F, 0x00, 0x00, 0x00, 0x00},
			[]interface{}{byte(0), []byte(nil), []byte{0x0F, 0x00, 0x00, 0x00, 0x00}, false},
		},
		{
			"ReadBinary/not enough bytes (subtype 2 length)",
			ReadBinary,
			[]byte{0x03, 0x00, 0x00, 0x00, 0x02, 0x0F, 0x00, 0x00},
			[]interface{}{byte(0), []byte(nil), []byte{0x03, 0x00, 0x00, 0x00, 0x02, 0x0F, 0x00, 0x00}, false},
		},
		{
			"ReadBinary/not enough bytes (subtype 2 value)",
			ReadBinary,
			[]byte{0x0F, 0x00, 0x00, 0x00, 0x02, 0x0F, 0x00, 0x00, 0x00, 0x01, 0x02},
			[]interface{}{
				byte(0), []byte(nil),
				[]byte{0x0F, 0x00, 0x00, 0x00, 0x02, 0x0F, 0x00, 0x00, 0x00, 0x01, 0x02}, false,
			},
		},
		{
			"ReadBinary/success (subtype 2)",
			ReadBinary,
			[]byte{0x06, 0x00, 0x00, 0x00, 0x02, 0x02, 0x00, 0x00, 0x00, 0x01, 0x02},
			[]interface{}{byte(0x02), []byte{0x01, 0x02}, []byte{}, true},
		},
		{
			"ReadBinary/success",
			ReadBinary,
			[]byte{0x03, 0x00, 0x00, 0x00, 0xFF, 0x01, 0x02, 0x03},
			[]interface{}{byte(0xFF), []byte{0x01, 0x02, 0x03}, []byte{}, true},
		},
		{
			"ReadObjectID/not enough bytes",
			ReadObjectID,
			[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			[]interface{}{objectid.ObjectID{}, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, false},
		},
		{
			"ReadObjectID/success",
			ReadObjectID,
			[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C},
			[]interface{}{
				objectid.ObjectID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C},
				[]byte{}, true,
			},
		},
		{
			"ReadBoolean/not enough bytes",
			ReadBoolean,
			[]byte{},
			[]interface{}{false, []byte{}, false},
		},
		{
			"ReadBoolean/success",
			ReadBoolean,
			[]byte{0x01},
			[]interface{}{true, []byte{}, true},
		},
		{
			"ReadDateTime/not enough bytes",
			ReadDateTime,
			[]byte{0x01, 0x02, 0x03, 0x04},
			[]interface{}{int64(0), []byte{0x01, 0x02, 0x03, 0x04}, false},
		},
		{
			"ReadDateTime/success",
			ReadDateTime,
			[]byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
			[]interface{}{int64(65536), []byte{}, true},
		},
		{
			"ReadRegex/not enough bytes (pattern)",
			ReadRegex,
			[]byte{},
			[]interface{}{"", "", []byte{}, false},
		},
		{
			"ReadRegex/not enough bytes (options)",
			ReadRegex,
			[]byte{'f', 'o', 'o', 0x00},
			[]interface{}{"", "", []byte{'f', 'o', 'o', 0x00}, false},
		},
		{
			"ReadRegex/success",
			ReadRegex,
			[]byte{'f', 'o', 'o', 0x00, 'b', 'a', 'r', 0x00},
			[]interface{}{"foo", "bar", []byte{}, true},
		},
		{
			"ReadDBPointer/not enough bytes (ns)",
			ReadDBPointer,
			[]byte{},
			[]interface{}{"", objectid.ObjectID{}, []byte{}, false},
		},
		{
			"ReadDBPointer/not enough bytes (objectID)",
			ReadDBPointer,
			[]byte{0x04, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00},
			[]interface{}{"", objectid.ObjectID{}, []byte{0x04, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00}, false},
		},
		{
			"ReadDBPointer/success",
			ReadDBPointer,
			[]byte{
				0x04, 0x00, 0x00, 0x00, 'f', 'o', 'o', 0x00,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C,
			},
			[]interface{}{
				"foo", objectid.ObjectID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C},
				[]byte{}, true,
			},
		},
		{
			"ReadJavaScript/not enough bytes (length)",
			ReadJavaScript,
			[]byte{},
			[]interface{}{"", []byte{}, false},
		},
		{
			"ReadJavaScript/not enough bytes (value)",
			ReadJavaScript,
			[]byte{0x0F, 0x00, 0x00, 0x00},
			[]interface{}{"", []byte{0x0F, 0x00, 0x00, 0x00}, false},
		},
		{
			"ReadJavaScript/success",
			ReadJavaScript,
			[]byte{0x07, 0x00, 0x00, 0x00, 'f', 'o', 'o', 'b', 'a', 'r', 0x00},
			[]interface{}{"foobar", []byte{}, true},
		},
		{
			"ReadSymbol/not enough bytes (length)",
			ReadSymbol,
			[]byte{},
			[]interface{}{"", []byte{}, false},
		},
		{
			"ReadSymbol/not enough bytes (value)",
			ReadSymbol,
			[]byte{0x0F, 0x00, 0x00, 0x00},
			[]interface{}{"", []byte{0x0F, 0x00, 0x00, 0x00}, false},
		},
		{
			"ReadSymbol/success",
			ReadSymbol,
			[]byte{0x07, 0x00, 0x00, 0x00, 'f', 'o', 'o', 'b', 'a', 'r', 0x00},
			[]interface{}{"foobar", []byte{}, true},
		},
		{
			"ReadCodeWithScope/ not enough bytes (length)",
			ReadCodeWithScope,
			[]byte{},
			[]interface{}{"", []byte(nil), []byte{}, false},
		},
		{
			"ReadCodeWithScope/ not enough bytes (value)",
			ReadCodeWithScope,
			[]byte{0x0F, 0x00, 0x00, 0x00},
			[]interface{}{"", []byte(nil), []byte{0x0F, 0x00, 0x00, 0x00}, false},
		},
		{
			"ReadCodeWithScope/not enough bytes (code value)",
			ReadCodeWithScope,
			[]byte{
				0x0C, 0x00, 0x00, 0x00,
				0x0F, 0x00, 0x00, 0x00,
				'f', 'o', 'o', 0x00,
			},
			[]interface{}{
				"", []byte(nil),
				[]byte{
					0x0C, 0x00, 0x00, 0x00,
					0x0F, 0x00, 0x00, 0x00,
					'f', 'o', 'o', 0x00,
				},
				false,
			},
		},
		{
			"ReadCodeWithScope/success",
			ReadCodeWithScope,
			[]byte{
				0x19, 0x00, 0x00, 0x00,
				0x07, 0x00, 0x00, 0x00, 'f', 'o', 'o', 'b', 'a', 'r', 0x00,
				0x0A, 0x00, 0x00, 0x00, 0x0A, 'f', 'o', 'o', 0x00, 0x00,
			},
			[]interface{}{
				"foobar", []byte{0x0A, 0x00, 0x00, 0x00, 0x0A, 'f', 'o', 'o', 0x00, 0x00},
				[]byte{}, true,
			},
		},
		{
			"ReadInt32/not enough bytes",
			ReadInt32,
			[]byte{0x01},
			[]interface{}{int32(0), []byte{0x01}, false},
		},
		{
			"ReadInt32/success",
			ReadInt32,
			[]byte{0x00, 0x01, 0x00, 0x00},
			[]interface{}{int32(256), []byte{}, true},
		},
		{
			"ReadTimestamp/not enough bytes (increment)",
			ReadTimestamp,
			[]byte{},
			[]interface{}{uint32(0), uint32(0), []byte{}, false},
		},
		{
			"ReadTimestamp/not enough bytes (timestamp)",
			ReadTimestamp,
			[]byte{0x00, 0x01, 0x00, 0x00},
			[]interface{}{uint32(0), uint32(0), []byte{0x00, 0x01, 0x00, 0x00}, false},
		},
		{
			"ReadTimestamp/success",
			ReadTimestamp,
			[]byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00},
			[]interface{}{uint32(65536), uint32(256), []byte{}, true},
		},
		{
			"ReadInt64/not enough bytes",
			ReadInt64,
			[]byte{0x01},
			[]interface{}{int64(0), []byte{0x01}, false},
		},
		{
			"ReadInt64/success",
			ReadInt64,
			[]byte{0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
			[]interface{}{int64(4294967296), []byte{}, true},
		},
		{
			"ReadDecimal128/not enough bytes (low)",
			ReadDecimal128,
			[]byte{},
			[]interface{}{decimal.Decimal128{}, []byte{}, false},
		},
		{
			"ReadDecimal128/not enough bytes (high)",
			ReadDecimal128,
			[]byte{0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
			[]interface{}{decimal.Decimal128{}, []byte{0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}, false},
		},
		{
			"ReadDecimal128/success",
			ReadDecimal128,
			[]byte{
				0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
			},
			[]interface{}{decimal.NewDecimal128(4294967296, 16777216), []byte{}, true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := reflect.ValueOf(tc.fn)
			if fn.Kind() != reflect.Func {
				t.Fatalf("fn must be of kind Func but it is a %v", fn.Kind())
			}
			if fn.Type().NumIn() != 1 || fn.Type().In(0) != reflect.TypeOf([]byte{}) {
				t.Fatalf("fn must have one parameter and it must be a []byte.")
			}
			results := fn.Call([]reflect.Value{reflect.ValueOf(tc.param)})
			if len(results) != len(tc.expected) {
				t.Fatalf("Length of results does not match. got %d; want %d", len(results), len(tc.expected))
			}
			for idx := range results {
				got := results[idx].Interface()
				want := tc.expected[idx]
				if !cmp.Equal(got, want, cmp.Comparer(compareDecimal128)) {
					t.Errorf("Result %d does not match. got %v; want %v", idx, got, want)
				}
			}
		})
	}
}

func compareDecimal128(d1, d2 decimal.Decimal128) bool {
	d1H, d1L := d1.GetBytes()
	d2H, d2L := d2.GetBytes()

	if d1H != d2H {
		return false
	}

	if d1L != d2L {
		return false
	}

	return true
}
