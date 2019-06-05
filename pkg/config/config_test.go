package config

import (
	"reflect"
	"testing"
)

type Config struct {
	FieldBool    bool
	FieldInt     int
	FieldInt8    int8
	FieldInt16   int16
	FieldInt32   int32
	FieldInt64   int64
	FieldUint    uint
	FieldUint8   uint8
	FieldUint16  uint16
	FieldUint32  uint32
	FieldUint64  uint64
	FieldString  string
	FieldFloat32 float32
	FieldFloat64 float64
	fieldBool    bool
	fieldInt     int
	fieldInt8    int8
	fieldInt16   int16
	fieldInt32   int32
	fieldInt64   int64
	fieldUint    uint
	fieldUint8   uint8
	fieldUint16  uint16
	fieldUint32  uint32
	fieldUint64  uint64
	fieldFloat32 float32
	fieldFloat64 float64
	fieldString  string
}

func TestConfigItems(t *testing.T) {
	cfg := &Config{}
	fields, err := configItems(cfg)
	if err != nil {
		t.Errorf("No error expeted, got %s", err)
	}
	if len(fields) != 14 {
		t.Errorf("Number of exported fields should be 14, got %d", len(fields))
	}
	for _, f := range fields {
		switch f.v.Kind() {
		case reflect.Bool:
			f.v.SetBool(true)
		case reflect.Int:
			f.v.SetInt(-1)
		case reflect.Int8:
			f.v.SetInt(-8)
		case reflect.Int16:
			f.v.SetInt(-16)
		case reflect.Int32:
			f.v.SetInt(-32)
		case reflect.Int64:
			f.v.SetInt(-64)
		case reflect.Uint:
			f.v.SetUint(1)
		case reflect.Uint8:
			f.v.SetUint(8)
		case reflect.Uint16:
			f.v.SetUint(16)
		case reflect.Uint32:
			f.v.SetUint(32)
		case reflect.Uint64:
			f.v.SetUint(64)
		case reflect.Float32:
			f.v.SetFloat(.32)
		case reflect.Float64:
			f.v.SetFloat(.64)
		case reflect.String:
			f.v.SetString("hello")
		default:
			t.Errorf("Unexpected kind, %s", f.v.Kind())
		}
	}
	if cfg.FieldBool != true {
		t.Errorf("Expecting cfg.FieldBool set to true, got %t", cfg.FieldBool)
	}
	if cfg.FieldInt != -1 {
		t.Errorf("Expecting cfg.FieldInt set to -1, got %d", cfg.FieldInt)
	}
	if cfg.FieldInt8 != -8 {
		t.Errorf("Expecting cfg.FieldInt8 set to -8, got %d", cfg.FieldInt8)
	}
	if cfg.FieldInt16 != -16 {
		t.Errorf("Expecting cfg.FieldInt16 set to -16, got %d", cfg.FieldInt16)
	}
	if cfg.FieldInt32 != -32 {
		t.Errorf("Expecting cfg.FieldInt32 set to -32, got %d", cfg.FieldInt32)
	}
	if cfg.FieldInt64 != -64 {
		t.Errorf("Expecting cfg.FieldInt64 set to -64, got %d", cfg.FieldInt64)
	}
	if cfg.FieldUint != 1 {
		t.Errorf("Expecting cfg.FieldUint set to 1, got %d", cfg.FieldUint)
	}
	if cfg.FieldUint8 != 8 {
		t.Errorf("Expecting cfg.FieldUint8 set to 8, got %d", cfg.FieldUint8)
	}
	if cfg.FieldUint16 != 16 {
		t.Errorf("Expecting cfg.FieldUint16 set to 16, got %d", cfg.FieldUint16)
	}
	if cfg.FieldUint32 != 32 {
		t.Errorf("Expecting cfg.FieldUint32 set to 1, got %d", cfg.FieldUint32)
	}
	if cfg.FieldUint64 != 64 {
		t.Errorf("Expecting cfg.FieldUint64 set to 64, got %d", cfg.FieldUint64)
	}
	if cfg.FieldFloat32 != .32 {
		t.Errorf("Expecting cfg.FieldUint64 set to 0.32, got %f", cfg.FieldFloat32)
	}
	if cfg.FieldFloat64 != .64 {
		t.Errorf("Expecting cfg.FieldUint64 set to 0.64, got %f", cfg.FieldFloat64)
	}
	if cfg.FieldString != "hello" {
		t.Errorf("Expecting cfg.FieldString set to \"hello\", got %q", cfg.FieldString)
	}
}

func TestNameToEnv(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"x", "X"},
		{"X", "X"},
		{"xYz", "X_YZ"},
		{"xYZ", "X_YZ"},
		{"x_yz", "X_YZ"},
		{"x_Yz", "X_YZ"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := nameToEnv(tt.in)
			if got != tt.out {
				t.Errorf("got %q, want %q", got, tt.out)
			}
		})
	}
}
