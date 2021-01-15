package key

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Masses is used to encode and/or decode a table containing the masses for each
// atom type from a LAMMPS data file. This table has a header where a blank line
// separate the values from it. Each value (= 1 line) has 2 columns (the first
// one is the type, the second is the mass). More information about the
// structure of this table can be found in the LAMMPS documentation.
//
// Masses can be instanced by using the built-in new function.
type Masses struct {
	types *Header
	v     map[int]float64
}

// Name returns NameMasses. It corresponds to the header of the table.
func (m *Masses) Name() Name {
	return NameMasses
}

// Keyword tests whether the byte slice s begins with Name after trimming the
// spaces. Keyword is useful to detect the header of the Masses table.
func (m *Masses) Keyword(s []byte) bool {
	return keyword(s, []byte(m.Name()))
}

// SetKeys assigns one or more Keys to Atoms. This method only accepts *Header
// with Name equal to NameAtomTypes.
func (m *Masses) SetKeys(k ...Key) error {
	if len(k) != 1 {
		return fmt.Errorf("too much keys: only one key is accepted")
	}
	var ok bool
	m.types, ok = k[0].(*Header) // may be dangerous to assign before checking for errors
	if !ok {
		return fmt.Errorf("type assertion error: Key provided is not *Header")
	}
	if m.types.name != NameAtomTypes {
		return fmt.Errorf("Key provided does not have a Name equal to NameAtomTypes")
	}
	return nil
}

// SetKeysVal assigns to the NameAtomTypes Key the number of types based on the
// length of the map that is created via the Set or Decode methods.
//
// This method needs a Key in order to work. This Key is an instance of Header
// with Name equal to NameAtomTypes. Use the Set method to assign this Key.
func (m *Masses) SetKeysVal() error {
	if m.types == nil {
		return fmt.Errorf("Key that is an instance of *Header with Name equal to NameAtomTypes is nil: use the Set method")
	}
	return m.types.Set(len(m.v))
}

// Encode writes a table containing the header, a blank line and each value (= 1
// line) (mass) into a writer.
//
// This method does not check the integrity and correctness of each value. To do
// so, use the Check method.
func (m *Masses) Encode(w io.Writer) error {
	if m.v == nil {
		return fmt.Errorf("map[int]float64 is nil: use the Decode or Set methods")
	}
	if len(m.v) == 0 {
		return nil
	}
	keys := sortIntsMap(m.v)
	fmt.Fprint(w, m.Name(), "\n\n")
	for _, k := range keys {
		v := m.v[k]
		_, err := fmt.Fprintf(w, "%d %g\n", k, v)
		if err != nil {
			return fmt.Errorf("fmt.Fprintf: %w", err)
		}
	}
	return nil
}

// Decode reads a reader where the offset is after the header of the table (at
// the beginning of the blank line). It reads each value (= 1 line) and decodes
// a float64 that is put into a map where the keys are the identifiers of each
// mass.
//
// This method needs a Key in order to work. This Key is an instance of Header
// with Name equal to NameAtomTypes. Use the Set method to assign this Key.
//
// Moreover, this method does not check the integrity and corectness of the
// values decoded. To do so, use the Check method.
//
// Decode method does not return io.EOF error. The use of the Check method after
// Decode is therefore highly recommended.
func (m *Masses) Decode(s []byte, r *bufio.Scanner) error {
	if m.types == nil {
		return fmt.Errorf("Key that is an instance of *Header with Name equal to NameAtomTypes is nil: use the Set method")
	}

	m.v = make(map[int]float64)

	if ok := r.Scan(); !ok {
		if r.Err() != nil {
			return fmt.Errorf("r.Scan first line: %w", r.Err())
		}
		return nil
	}

	types := m.types.Get().(int)
	for i := 0; i < types && r.Scan(); i++ {
		f := strings.Fields(r.Text())
		if len(f) < 2 {
			return fmt.Errorf("not enough fields = %d, expected > 2", len(f))
		}
		atomType, err := strconv.Atoi(f[0])
		if err != nil {
			return fmt.Errorf("strconv.Atoi: %w", err)
		}
		mass, err := strconv.ParseFloat(f[1], 64)
		if err != nil {
			return fmt.Errorf("strconv.ParseFloat: %w", err)
		}
		m.v[atomType] = mass
	}
	if r.Err() != nil {
		return fmt.Errorf("r.Scan: %w", r.Err())
	}
	return nil
}

// Set puts a custom map[int]float64.
//
// This method does not check the integrity or correctness of the passed data.
// The use of the Check method after Set is therefore highly recommended.
func (m *Masses) Set(v interface{}) error {
	var ok bool
	m.v, ok = v.(map[int]float64)
	if !ok {
		return fmt.Errorf("type assertion error: value is not map[int]float64")
	}
	return nil
}

// Get returns a map[int]float64 where the keys are the atom types. As this
// method returns an interface, it must be useful to perform a type assertion
// after calling this method.
func (m *Masses) Get() interface{} {
	return m.v
}

// Check verifies the integrity and correctness of the data decoded with the
// Decode method or set with the Set method.
//
// This method needs a Key in order to work. This Key is an instance of Header
// with Name equal to NameAtomTypes. Use the Set method to assign this Key.
func (m *Masses) Check() error {
	if m.types == nil {
		return fmt.Errorf("Key that is an instance of *Header with Name equal to NameAtomTypes is nil: use the Set method")
	}
	types := m.types.Get().(int)
	if len(m.v) != types {
		return fmt.Errorf("number of masses (= 1 line = 1 type) = %d is not equal to the number of atom types = %d", len(m.v), types)
	}
	for typ, mass := range m.v {
		if mass < 0. {
			return fmt.Errorf("mass of type = %d is lower than zero = %g", typ, mass)
		}
		if typ < 1 || typ > types {
			return fmt.Errorf("type = %d is invalid: it must be greater than zero and lower or equal than the number of types = %d", typ, types)
		}
	}
	return nil
}
