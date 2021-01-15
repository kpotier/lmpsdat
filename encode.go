package lmpsdat

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/kpotier/lmpsdat/key"
)

// Encoder writes LAMMPS data values to an input stream.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
	}
}

// Encode writes the LAMMPS data of v to the stream.
func (enc *Encoder) Encode(v interface{}) error {
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

	for n, f := range nFields {
		field := val.Field(f).Interface()
		k := keys[n]
		if err := k.Set(field); err != nil {
			return fmt.Errorf("k.Set for Key = %s: %w", n, err)
		}
		if err := k.SetKeysVal(); err != nil && !errors.Is(err, key.ErrUnsupported) {
			return fmt.Errorf("k.SetKeysVal for Key = %s: %w", n, err)
		}
	}

	for _, k := range keys {
		err := k.Check()
		if err != nil {
			return fmt.Errorf("k.Check for Key = %s: %w", k.Name(), err)
		}
	}

	var title string
	if k, ok := keys[key.NameTitle]; ok {
		title = k.Get().(string)
	}
	fmt.Fprintf(enc.w, "%s\n\n", title) // errors are omitted and will appear when using k.Encode

	set := false
	nbr := []key.Name{key.NameAtomsNbr, key.NameBondsNbr, key.NameAnglesNbr, key.NameDihedralsNbr}
	for _, n := range nbr {
		if k, ok := keys[n]; ok {
			if err := k.Encode(enc.w); err != nil {
				return fmt.Errorf("k.Encode for Key = %s: %w", n, err)
			}
			set = true
		}
	}
	if set {
		fmt.Fprint(enc.w, "\n")
	}

	set = false
	types := []key.Name{key.NameAtomTypes, key.NameBondTypes, key.NameAngleTypes, key.NameDihedralTypes}
	for _, n := range types {
		if k, ok := keys[n]; ok {
			if err := k.Encode(enc.w); err != nil {
				return fmt.Errorf("k.Encode for Key = %s: %w", n, err)
			}
			set = true
		}
	}
	if set {
		fmt.Fprint(enc.w, "\n")
	}

	set = false
	box := []key.Name{key.NameBoxX, key.NameBoxY, key.NameBoxZ}
	for _, n := range box {
		if k, ok := keys[n]; ok {
			if err := k.Encode(enc.w); err != nil {
				return fmt.Errorf("k.Encode for Key = %s: %w", n, err)
			}
			set = true
		}
	}
	if set {
		fmt.Fprint(enc.w, "\n")
	}

	tables := []key.Name{key.NameMasses, key.NamePairCoeffs, key.NameBondCoeffs, key.NameAngleCoeffs, key.NameDihedralCoeffs, key.NameAtoms, key.NameBonds, key.NameAngles, key.NameDihedrals}
	for _, n := range tables {
		if k, ok := keys[n]; ok {
			if err := k.Encode(enc.w); err != nil {
				return fmt.Errorf("k.Encode for Key = %s: %w", n, err)
			}
			fmt.Fprint(enc.w, "\n")
		}
	}

	return nil
}
