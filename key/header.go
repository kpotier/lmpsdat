package key

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"unicode"
)

// Header is used to encode and/or decode an integer followed by a keyword
// (Name) from a LAMMPS data file. Header can be used to describe the number of
// types, number of atoms, and so on.
//
// Coeffs can be instanced by using the NewHeader method.
type Header struct {
	name    Name
	nameSep [][]byte

	vBytes []byte
	v      int
}

// NewHeader returns an instance of Header.
func NewHeader(name Name) *Header {
	nameSep := bytes.Fields([]byte(name))
	return &Header{name: name, nameSep: nameSep}
}

// Name returns the Name passed in NewHeader. It corresponds to the keyword that
// is preceded by an integer. It can be NameAtomsNbr, NameAtomTypes, etc.
func (h *Header) Name() Name {
	return h.name
}

// Keyword tests whether the byte slice s ends with the Name after an integer.
// Keyword is useful to detect if Header can correctly decode the integer.
func (h *Header) Keyword(s []byte) bool {
	s = bytes.TrimLeftFunc(s, unicode.IsSpace)
	idx := bytes.IndexFunc(s, unicode.IsSpace) // always a space after the number. After this space there is Name.
	if idx != -1 {
		h.vBytes = s[:idx] // store the number as []byte to allow faster decoding in Decode.
		return keywordHeader(s[idx:], h.nameSep)
	}
	return false
}

// SetKeys assigns one or more Keys to Header. This method always return
// ErrUnsupported as it is unsupported by Header.
func (h *Header) SetKeys(k ...Key) error {
	return ErrUnsupported
}

// SetKeysVal returns ErrUnsupported as it is unsupported by Header.
func (h *Header) SetKeysVal() error {
	return ErrUnsupported
}

// Encode writes an integer followed by the Name into a writer. For instance, it
// writes, "%int% atoms" if Name is NameAtomsNbr.
//
// This method does not check the integrity and correctness of each value. To do
// so, use the Check method.
func (h *Header) Encode(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%d %s\n", h.Get().(int), h.Name())
	return err
}

// Decode converts the integer for a specific keyword. This method will return
// errors if Keyword was not called before.
//
// This method does not check the integrity or correctness of the passed data.
// The use of the Check method after Decode is therefore highly recommended.
func (h *Header) Decode(s []byte, r *bufio.Scanner) error {
	var err error
	h.v, err = strconv.Atoi(string(h.vBytes))
	if err != nil {
		return fmt.Errorf("strconv.Atoi: %w", err)
	}
	return nil
}

// Set puts a custom int.
//
// This method does not check the integrity or correctness of the passed data.
// The use of the Check method after Set is therefore highly recommended.
func (h *Header) Set(v interface{}) error {
	var ok bool
	h.v, ok = v.(int)
	if !ok {
		return fmt.Errorf("type assertion (int) error")
	}
	return nil
}

// Get returns int. As this method returns an interface, it must be useful to
// perform a type assertion after calling this method.
func (h *Header) Get() interface{} {
	return h.v
}

// Check verifies the integrity and correctness of the data decoded with the
// Decode method or set with the Set method.
func (h *Header) Check() error {
	if h.v < 0 {
		return fmt.Errorf("integer = %d is lower than zero", h.v)
	}
	return nil
}
