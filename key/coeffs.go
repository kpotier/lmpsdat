package key

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Coeffs is used to encode and/or decode a table containing the coefficients
// (e.g. Bond Coeffs, Pair Coeffs) from a LAMMPS data file. This table has a
// header where a blank line separate the values from it. Each value (= 1 line =
// 1 type) has 2 or more columns. More information about the structure of this
// table can be found in the LAMMPS documentation.
//
// Coeffs can be instanced by using the NewCoeffs method.
type Coeffs struct {
	name  Name
	types *Header
	v     map[int][]float64
}

// NewCoeffs returns an instance of Coeffs. The recommended Names are
// NameBondCoeffs, NamePairCoeffs, NameAngleCoeffs, and NameDihedralCoeffs.
func NewCoeffs(name Name) *Coeffs {
	return &Coeffs{name: name}
}

// Name returns the Name passed in NewCoeffs. It corresponds to the header of
// the table.
func (c *Coeffs) Name() Name {
	return c.name
}

// Keyword tests whether the byte slice s begins with Name after trimming the
// spaces. Keyword is useful to detect the header of the Coeffs table.
func (c *Coeffs) Keyword(s []byte) bool {
	return keyword(s, []byte(c.Name()))
}

// SetKeys assigns one or more Keys to Atoms. This method only accepts *Header
// with Name equal to NamexxxTypes where xxx can be Atom, Angle, Bond, etc. Only
// one Key must be passed.
func (c *Coeffs) SetKeys(k ...Key) error {
	if len(k) != 1 {
		return fmt.Errorf("only one Key is accepted")
	}
	var ok bool
	c.types, ok = k[0].(*Header)
	if !ok {
		return fmt.Errorf("type assertion error: Key provided is not *Header")
	}
	return nil
}

// SetKeysVal assigns to the NamexxxTypes (where xxx can be Atom, Angle, Bond,
// etc.) Key the number of types based on the length of the map that is created
// via the Set or Decode methods.
//
// This method needs a Key in order to work. This Key is an instance of Header
// with Name equal to NamesxxTypes where xxx can be Atom, Angle, Bond, etc. Use
// the Set method to assign this Key.
func (c *Coeffs) SetKeysVal() error {
	if c.types == nil {
		return fmt.Errorf("Key that is an instance of *Header with Name equal to NamexxxTypes is nil: use the Set method")
	}
	return c.types.Set(len(c.v))
}

// Encode writes a table containing the header, a blank line and each value (= 1
// line = 1 type) into a writer.
//
// This method does not check the integrity and correctness of each value. To do
// so, use the Check method.
func (c *Coeffs) Encode(w io.Writer) error {
	if c.v == nil {
		return fmt.Errorf("map[int][]float64 is nil: use the Decode or Set methods")
	}
	if len(c.v) == 0 {
		return nil
	}

	keys := sortIntsMap(c.v)
	fmt.Fprint(w, c.Name(), "\n\n")
	for _, k := range keys {
		if _, err := fmt.Fprintf(w, "%d", k); err != nil {
			return fmt.Errorf("fmt.Fprintf: %w", err)
		}
		for _, v := range c.v[k] {
			if _, err := fmt.Fprintf(w, " %g", v); err != nil {
				return fmt.Errorf("fmt.Fprintf coeff: %w", err)
			}
		}
		if _, err := fmt.Fprint(w, "\n"); err != nil {
			return fmt.Errorf("fmt.Fprintf newline: %w", err)
		}
	}
	return nil
}

// Decode reads a reader where the offset is after the header of the table (at
// the beginning of the blank line). It reads each value (= 1 line) and creates
// a slice of float64s that is put into a map where the keys are the identifiers
// of each set of coefficient.
//
// This method needs a Key in order to work. This Key is an instance of Header
// with Name equal to NamexxxTypes where xxx can be Atom, Angle, Bond, etc. Use
// the Set method to assign this Key.
//
// Moreover, this method does not check the integrity and corectness of the
// values decoded. To do so, use the Check method.
//
// Decode method does not return io.EOF error. The use of the Check method after
// Decode is therefore highly recommended.
func (c *Coeffs) Decode(s []byte, r *bufio.Scanner) error {
	if c.types == nil {
		return fmt.Errorf("Key that is an instance of *Header with Name equal to NamexxxTypes is nil: use the Set method")
	}

	types := c.types.Get().(int)
	c.v = make(map[int][]float64)

	if ok := r.Scan(); !ok {
		if r.Err() != nil {
			return fmt.Errorf("r.Scan first line: %w", r.Err())
		}
		return nil
	}

	for i := 0; i < types && r.Scan(); i++ {
		s := delComments(r.Bytes())
		f := strings.Fields(string(s))
		if len(f) < 2 {
			return fmt.Errorf("not enough fields = %d, want >= 2", len(f))
		}
		typ, err := strconv.Atoi(f[0])
		if err != nil {
			return fmt.Errorf("strconv.Atoi type: %w", err)
		}
		var coeffs []float64
		for _, v := range f[1:] {
			coeff, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return fmt.Errorf("strconv.ParseFloat: %w", err)
			}
			coeffs = append(coeffs, coeff)
		}
		c.v[typ] = coeffs
	}
	if r.Err() != nil {
		return fmt.Errorf("r.Scan: %w", r.Err())
	}
	return nil
}

// Set puts a custom map[int][]float64.
//
// This method does not check the integrity or correctness of the passed data.
// The use of the Check method after Set is therefore highly recommended.
func (c *Coeffs) Set(v interface{}) error {
	var ok bool
	c.v, ok = v.(map[int][]float64)
	if !ok {
		return fmt.Errorf("type assertion error: value is not map[int][]float64")
	}
	return nil
}

// Get returns a map[int][]float64 where the keys are the identifiers of the atoms.
// As this method returns an interface, it must be useful to perform a type
// assertion after calling this method.
func (c *Coeffs) Get() interface{} {
	return c.v
}

// Check verifies the integrity and correctness of the data decoded with the
// Decode method or set with the Set method.
//
// This method needs a Keys in order to work. This Key is an instance of Header
// with Name equal to NamexxxTypes where xxx can be Atom, Angle, Bond, etc. Use
// the Set method to assign this Key.
func (c *Coeffs) Check() error {
	if c.types == nil {
		return fmt.Errorf("Key that is an instance of *Header with Name equal to NamexxxTypes is nil: use the Set method")
	}
	types := c.types.Get().(int)
	if len(c.v) != types {
		return fmt.Errorf("number of sets of coefficients (= 1 line = 1 type) = %d is not equal to the number of types = %d", len(c.v), types)
	}
	for typ := range c.v {
		if typ < 1 || typ > types {
			return fmt.Errorf("type = %d is invalid: it must be greater than zero and lower or equal than the number of types = %d", typ, types)
		}
	}
	return nil
}
