package key

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"unicode"
)

// Box is used to encode and/or decode the size of the box for a specific
// coordinate (x, y, or z) from a LAMMPS data file. It is represented as
// "%float64% %float64% xlo xhi" where xlo is the point in space where the box
// begins in the x coordinate (given by the first float64), and xhi is the point
// in space where the box ends in the x coordinate (given by the second
// float64). For instance, "0.0 1.0 ylo yhi" means that the box in the y
// coordinate starts at 0.0 and ends at 1.0.
//
// Box must be instanced by using the NewBox function.
type Box struct {
	name    Name
	nameSep [][]byte

	vBytes [2][]byte
	vlo    float64
	vhi    float64
}

// NewBox returns an instance of Box. The recommended Names are NameBoxX,
// NameBoxY, and NameBoxZ.
func NewBox(name Name) *Box {
	nameSep := bytes.Fields([]byte(name))
	return &Box{name: name, nameSep: nameSep}
}

// Name returns the Name passed in NewBox. It corresponds to the keyword that
// can whether be "xlo xhi" (NameBoxX) or "ylo yhi" (NameBoxY) or "zlo zhi"
// (NameBoxX).
func (b *Box) Name() Name {
	return b.name
}

// Keyword tests whether the byte slice s ends with the Name after two float64s.
// Keyword is useful to detect if Box can correctly decode the two float64s.
func (b *Box) Keyword(s []byte) bool {
	for i := 0; i < 2; i++ {
		s = bytes.TrimLeftFunc(s, unicode.IsSpace)
		idx := bytes.IndexFunc(s, unicode.IsSpace)
		if idx < 1 {
			return false
		}
		b.vBytes[i] = s[:idx] // store the two float64s as []byte to allow faster decoding.
		s = s[idx:]
	}
	return keywordHeader(s, b.nameSep)
}

// SetKeys assigns one or more Keys to Box. This method always return
// ErrUnsupported as it is unsupported by Box.
func (b *Box) SetKeys(k ...Key) error {
	return ErrUnsupported
}

// SetKeysVal returns ErrUnsupported as it is unsupported by Box.
func (b *Box) SetKeysVal() error {
	return ErrUnsupported
}

// Encode writes the box size followed by the Name into a writer. For instance,
// it writes, "%float64% %float64% xlo xhi" if Name is NameBoxX.
//
// This method does not check the integrity and correctness of each value. To do
// so, use the Check method.
func (b *Box) Encode(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%g %g %s\n", b.vlo, b.vhi, b.Name())
	return err
}

// Decode converts the box size for a specific coordinate into two float64s.
// This method will return errors if Keyword was not called before.
//
// This method does not check the integrity or correctness of the passed data.
// The use of the Check method after Decode is therefore highly recommended.
func (b *Box) Decode(s []byte, r *bufio.Scanner) error {
	var err error
	b.vlo, err = strconv.ParseFloat(string(b.vBytes[0]), 64)
	if err != nil {
		return fmt.Errorf("strconv.ParseFloat lo: %w", err)
	}
	b.vhi, err = strconv.ParseFloat(string(b.vBytes[1]), 64)
	return err
}

// Set puts a custom [2]float64.
//
// This method does not check the integrity or correctness of the passed data.
// The use of the Check method after Set is therefore highly recommended.
func (b *Box) Set(v interface{}) error {
	val, ok := v.([2]float64)
	if !ok {
		return fmt.Errorf("type assertion error: value is not [2]float64")
	}
	b.vlo = val[0]
	b.vhi = val[1]
	return nil
}

// Get returns [2]float64 where the first value is the point where the box
// begins and the second value is the point in space where the box ends. As this
// method returns an interface, it must be useful to perform a type assertion
// after calling this method.
func (b *Box) Get() interface{} {
	return [2]float64{b.vlo, b.vhi}
}

// Check verifies the integrity and correctness of the data decoded with the
// Decode method or set with the Set method.
func (b *Box) Check() error {
	if b.vlo > b.vhi {
		return fmt.Errorf("lo = %g is greater than hi = %g", b.vlo, b.vhi)
	}
	return nil
}
