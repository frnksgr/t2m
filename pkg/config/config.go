package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type configItem struct {
	t reflect.StructField // to get name and tags
	v reflect.Value       // to set values
}

func configItems(cfg interface{}) ([]configItem, error) {
	t := reflect.TypeOf(cfg)
	v := reflect.ValueOf(cfg)
	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("Expected pointer, got %s", v.Kind())
	}
	vs := v.Elem()
	ts := t.Elem()
	if ts.Kind() != reflect.Struct {
		return nil, fmt.Errorf("Expected struct, got %s", ts.Kind())
	}
	fc := ts.NumField()
	fields := make([]configItem, 0, fc)
	for i := 0; i < fc; i++ {
		v := vs.Field(i)
		t := ts.Field(i)
		n := t.Name
		if n[0] <= 'Z' && n[0] >= 'A' { // exported field
			fields = append(fields, configItem{t, v})
		}
	}
	return fields, nil
}

func nameToEnv(n string) string {
	re := regexp.MustCompile("([a-z])([A-Z])")
	s := re.ReplaceAllString(n, "${1}_${2}")
	return strings.ToUpper(s)
}

// FromEnv bla bla
func FromEnv(cfg interface{}) error {
	fs, err := configItems(cfg)
	if err != nil {
		return err
	}
	for _, f := range fs {
		n := f.t.Tag.Get("env")
		if n == "" {
			n = nameToEnv(f.t.Name)
		}
		v := os.Getenv(n)
		if v != "" {
			switch f.v.Kind() {
			case reflect.Bool:
				v, err := strconv.ParseBool(v)
				if err != nil {
					return err
				}
				f.v.SetBool(v)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				v, err := strconv.Atoi(v)
				if err != nil {
					return err
				}
				f.v.SetInt(int64(v))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				v, err := strconv.Atoi(v)
				if err != nil {
					return err
				}
				f.v.SetUint(uint64(v))
			case reflect.Float32:
				v, err := strconv.ParseFloat(v, 32)
				if err != nil {
					return err
				}
				f.v.SetFloat(v)
			case reflect.Float64:
				v, err := strconv.ParseFloat(v, 34)
				if err != nil {
					return err
				}
				f.v.SetFloat(v)
			case reflect.String:
				f.v.SetString(v)
			default:
				return errors.New("unknown config type")
			}
		}
	}
	return nil
}

// TODO:

/*
// FromFlags bla bla
func FromFlags(cfg interface{}) error {
	return errors.New("not implemented")
}
*/

/*
// FromJSON bla bla
func FromJSON(cfg interface{}, r *io.Reader) error {
	return errors.New("not implemented")
}
*/
