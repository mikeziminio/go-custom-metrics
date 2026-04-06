package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
)

func FillFlags[T any](dst, def *T) error {
	if dst == nil {
		return fmt.Errorf("dst must not be nil")
	}
	if def == nil {
		return fmt.Errorf("def must not be nil")
	}

	dstElem := reflect.ValueOf(dst).Elem()
	defElem := reflect.ValueOf(def).Elem()

	if dstElem.Kind() != reflect.Struct || defElem.Kind() != reflect.Struct {
		return fmt.Errorf("src and def must be pointers to structs")
	}

	t := dstElem.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		f, ok := field.Tag.Lookup("flag")
		if !ok {
			continue
		}

		srcField := dstElem.Field(i)
		defField := defElem.Field(i)

		switch field.Type.Kind() {
		case reflect.String:
			doc := field.Tag.Get("doc")
			flag.StringVar(srcField.Addr().Interface().(*string), f, defField.String(), doc)
		case reflect.Float64:
			doc := field.Tag.Get("doc")
			flag.Float64Var(srcField.Addr().Interface().(*float64), f, defField.Float(), doc)
		case reflect.Int:
			doc := field.Tag.Get("doc")
			flag.IntVar(srcField.Addr().Interface().(*int), f, int(defField.Int()), doc)
		case reflect.Bool:
			doc := field.Tag.Get("doc")
			flag.BoolVar(srcField.Addr().Interface().(*bool), f, defField.Bool(), doc)
		}
	}
	return nil
}

func MergeOnlyFlags[T any](dst *T, src *T, flags map[string]struct{}) error {
	if dst == nil {
		return fmt.Errorf("dst must not be nil")
	}
	if src == nil {
		return fmt.Errorf("src must not be nil")
	}

	srcElem := reflect.ValueOf(src).Elem()
	dstElem := reflect.ValueOf(dst).Elem()

	if srcElem.Kind() != reflect.Struct || dstElem.Kind() != reflect.Struct {
		return fmt.Errorf("src and dst must be pointers to structs")
	}

	t := srcElem.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		f, ok := field.Tag.Lookup("flag")
		if !ok {
			continue
		}

		if _, set := flags[f]; set {
			srcField := srcElem.Field(i)
			dstField := dstElem.Field(i)
			dstField.Set(srcField)
		}
	}
	return nil
}

func FillConfigFromFile[T any](dst, def *T, path string) error {
	if dst == nil {
		return fmt.Errorf("dst must not be nil")
	}
	if def == nil {
		return fmt.Errorf("def must not be nil")
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	c := *def
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}
	*dst = c

	return nil
}
