package lmpsdat

import (
	"bufio"
	"fmt"
	"io"
	"reflect"

	"github.com/kpotier/lmpsdat/key"
)

// Decoder reads and decodes LAMMPS data values from an input stream.
type Decoder struct {
	r io.Reader
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: r,
	}
}

// Decode reads the next LAMMPS data-encoded value from its input and stores it
// in the value pointed to by v.
func (dec *Decoder) Decode(v interface{}) error {
	ptr := reflect.TypeOf(v)
	if ptr.Kind() != reflect.Ptr {
		return fmt.Errorf("interface passed is not a pointer")
	}

	val := reflect.ValueOf(v).Elem()
	typ := ptr.Elem()
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("interface passed is not a pointer of a struct")
	}

	nFields, keys := createNames(typ)
	kHead, kBody := headBody(keys)

	inHeader := true
	r := bufio.NewScanner(dec.r)

	if ok := r.Scan(); !ok {
		if r.Err() != nil {
			return fmt.Errorf("r.Scan title: %w", r.Err())
		}
		return nil
	}
	if k, ok := keys[key.NameTitle]; ok {
		if err := k.Set(r.Text()); err != nil {
			return fmt.Errorf("k.Set for Key = %s: %w", key.NameTitle, err)
		}
	}

	for r.Scan() {
		s := r.Bytes()
		if inHeader {
			ok, err := keyDecode(s, kHead, r)
			if err != nil {
				return err
			} else if ok {
				continue
			}
		}
		ok, err := keyDecode(s, kBody, r)
		if err != nil {
			return err
		} else if ok {
			inHeader = false
		}
	}
	if r.Err() != nil {
		return fmt.Errorf("r.Scan: %w", r.Err())
	}

	for _, k := range keys {
		err := k.Check()
		if err != nil {
			return fmt.Errorf("k.Check for Key = %s: %w", k.Name(), err)
		}
	}

	for n, f := range nFields {
		v := reflect.ValueOf(keys[n].Get())
		field := val.Field(f)
		if !field.Type().AssignableTo(v.Type()) {
			return fmt.Errorf("Key = %s has type = %s that is not assignable to type = %s", n, v.Type(), field.Type())
		}
		field.Set(v)
	}

	return nil
}
