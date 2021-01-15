package key

import (
	"bufio"
	"fmt"
	"io"
)

// Title is used to encode and/or decode the title from a LAMMPS data file. It
// is the first line of the LAMMPS data file.
//
// Title must be instanced by using the built-in new function.
type Title struct {
	v string
}

// Name returns NameTitle. It corresponds to the title of the LAMMPS data file
// that is the first line.
func (t *Title) Name() Name {
	return NameTitle
}

// Keyword always return false as it is unsupported by Title.
func (t *Title) Keyword(s []byte) bool {
	return false
}

// SetKeys assigns one or more Keys to Title. This method always return
// ErrUnsupported as it is unsupported by Title.
func (t *Title) SetKeys(k ...Key) error {
	return ErrUnsupported
}

// SetKeysVal returns ErrUnsupported as it is unsupported by Title.
func (t *Title) SetKeysVal() error {
	return ErrUnsupported
}

// Encode writes the title of the LAMMPS data file.
func (t *Title) Encode(w io.Writer) error {
	_, err := fmt.Fprintln(w, t.v)
	return err
}

// Decode assigns the title of the LAMMPS data file into Title.
func (t *Title) Decode(s []byte, r *bufio.Scanner) error {
	t.v = string(s)
	return nil
}

// Set puts a custom string.
func (t *Title) Set(v interface{}) error {
	val, ok := v.(string)
	if !ok {
		return fmt.Errorf("type assertion error: value is not string")
	}
	t.v = val
	return nil
}

// Get returns a string that is the title of LAMMPS data file.
func (t *Title) Get() interface{} {
	return t.v
}

// Check verifies the integrity and correctness of the data decoded with the
// Decode method or set with the Set method. This method always return nil.
func (t *Title) Check() error {
	return nil
}
