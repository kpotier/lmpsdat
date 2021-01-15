package key

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Atom contains information about a particular atom. For instance, it has the
// position and the charge. The identifier of the atom is not included in this
// structure.
//
// This structure is used by Atoms. A map where the keys are the identifiers of
// the atoms and the values are a pointer of Atom can be obtained or set with
// the Set or Get methods. Decode and Encode methods of Atoms make use of this
// map to encode/decode a table to/from a LAMMPS data file.
//
// The meaning of each term can be found in the LAMMPS documentation.
type Atom struct {
	MolTag   int
	AtomType int
	Q        float64
	X        float64
	Y        float64
	Z        float64

	// if N is set to true, NX, NY, and NZ must be specified.
	N  bool
	NX int
	NY int
	NZ int
}

// Atoms is used to encode and/or decode a table containing the atoms from a
// LAMMPS data file. This table has a header where a blank line separate the
// values from it. Each value (atom) (= 1 line) has 7 or 10 columns (if NX, NY,
// and NZ are set). More information about the structure of this table can be
// found in the LAMMPS documentation.
//
// Atoms can be instanced by using the NewAtoms function.
type Atoms struct {
	atomStyle AtomStyle
	atomsNbr  *Header
	atomTypes *Header
	v         map[int]*Atom
}

// NewAtoms returns an instance of Atoms with a specific atom style. It panics
// if the Atom Style does not exist.
func NewAtoms(as AtomStyle) *Atoms {
	return &Atoms{atomStyle: as}
}

// Name returns NameAtoms. It corresponds to the header of the table.
func (a *Atoms) Name() Name {
	return NameAtoms
}

// Keyword tests whether the byte slice s begins with Name after trimming the
// spaces. Keyword is useful to detect the header of the Atoms table.
func (a *Atoms) Keyword(s []byte) bool {
	return keyword(s, []byte(a.Name()))
}

// SetKeys assigns one or more Keys to Atoms. This method only accepts *Header
// with Name equal to NameAtomsNbr or NameAtomTypes.
func (a *Atoms) SetKeys(k ...Key) error {
	for _, key := range k {
		header, ok := key.(*Header)
		if !ok {
			return fmt.Errorf("type assertion error: Key provided is not *Header")
		}
		switch header.Name() {
		case NameAtomsNbr:
			a.atomsNbr = header
		case NameAtomTypes:
			a.atomTypes = header
		default:
			return fmt.Errorf("Key provided does not have a Name equal to NameAtomsNbr or NameAtomTypes")
		}
	}
	return nil
}

// Encode writes a table containing the header, a blank line and each value (= 1
// line) (atom) into a writer.
//
// This method does not check the integrity and correctness of each value. To do
// so, use the Check method.
func (a *Atoms) Encode(w io.Writer) error {
	if a.v == nil {
		return fmt.Errorf("map[int]*Atom is nil: use the Decode or Set methods")
	}
	if len(a.v) == 0 {
		return nil
	}

	keys := sortIntsMap(a.v)
	fmt.Fprint(w, a.Name(), "\n\n")
	for _, k := range keys {
		var err error
		var v = a.v[k]

		_, err = fmt.Fprintf(w, "%d ", k)
		if err != nil {
			return fmt.Errorf("fmt.Fprintf id: %w", err)
		}

		err = a.atomStyle.Encode(v, w)
		if err != nil {
			return fmt.Errorf("a.atomStyle.Encode named %s: %w", a.atomStyle.Name(), err)
		}

		if v.N {
			_, err = fmt.Fprintf(w, " %d %d %d\n", v.NX, v.NY, v.NZ)
		} else {
			_, err = fmt.Fprint(w, "\n")
		}
		if err != nil {
			return fmt.Errorf("fmt.Fprintf newline/optional params: %w", err)
		}
	}
	return nil
}

// Decode reads a reader where the offset is after the header of the Atoms table
// (at the beginning of the blank line). It reads each value (= 1 line) (atom)
// and creates an instance of Atom that is put into a map where the keys are the
// identifiers of each atom.
//
// This method needs a Key in order to work. This Key is an instance of Header
// with Name equal to NameAtomsNbr. Use the Set method to assign this Key.
//
// Moreover, this method does not check the integrity and corectness of the
// values decoded. To do so, use the Check method.
//
// Decode method does not return io.EOF error. The use of the Check method after
// Decode is therefore highly recommended.
func (a *Atoms) Decode(s []byte, r *bufio.Scanner) error {
	if a.atomsNbr == nil {
		return fmt.Errorf("Key that is an instance of *Header with Name equal to NameAtomsNbr is nil: use the Set method")
	}

	if ok := r.Scan(); !ok {
		if r.Err() != nil {
			return fmt.Errorf("r.Scan first line: %w", r.Err())
		}
		return nil
	}

	a.v = make(map[int]*Atom)
	atomsNbr := a.atomsNbr.Get().(int)
	for i := 0; i < atomsNbr && r.Scan(); i++ {
		s := delComments(r.Bytes())
		f := strings.Fields(string(s))
		id, atom, err := a.atomStyle.Decode(f)
		if err != nil {
			return err
		}
		a.v[id] = atom
	}
	if r.Err() != nil {
		return fmt.Errorf("r.Scan: %w", r.Err())
	}
	return nil
}

// Set puts a custom map[int]*Atom.
//
// This method does not check the integrity or correctness of the passed data.
// The use of the Check method after Set is therefore highly recommended.
func (a *Atoms) Set(v interface{}) error {
	var ok bool
	a.v, ok = v.(map[int]*Atom)
	if !ok {
		return fmt.Errorf("type assertion error: value is not map[int]*Atom)")
	}
	return nil
}

// Get returns a map[int]*Atom where the keys are the identifiers of the atoms.
// As this method returns an interface, it must be useful to perform a type
// assertion after calling this method.
func (a *Atoms) Get() interface{} {
	return a.v
}

// Check verifies the integrity and correctness of the data decoded with the
// Decode method or set with the Set method.
//
// This method needs two Keys in order to work. These Key are instances of
// Header with Name equal to NameAtomsNbr and NameAtomTypes. Use the Set
// method to assign these Keys.
func (a *Atoms) Check() error {
	if a.atomTypes == nil || a.atomsNbr == nil {
		return fmt.Errorf("one or more Keys are nil: use the Set method")
	}

	atomsNbr := a.atomsNbr.Get().(int)
	atomsTypes := a.atomTypes.Get().(int)

	if len(a.v) != atomsNbr {
		return fmt.Errorf("number of assigned atoms = %d is not equal to the number of expected atoms = %d", len(a.v), atomsNbr)
	}
	if len(a.v) == 0 {
		return nil
	}

	first := true
	n := false
	for typ, atom := range a.v {
		if first {
			n = atom.N // the first value is the reference
			first = false
		}
		if typ < 1 || typ > atomsNbr {
			return fmt.Errorf("identifier = %d is invalid: it must be greater than zero and lower or equal than the number of atoms = %d", typ, atomsNbr)
		}
		//if atom.MolTag < 1 {
		//	return fmt.Errorf("molecule tag is lower than one for atom %d", typ)
		//}
		if atom.AtomType < 1 || atom.AtomType > atomsTypes {
			return fmt.Errorf("type = %d is invalid: it must be greater than zero and lower or equal than the number of types = %d", atom.AtomType, atomsTypes)
		}
		if atom.N != n {
			return fmt.Errorf("n defined to %v but atom %d has n set to %v", n, typ, atom.N)
		}
	}
	return nil
}

// SetKeysVal assigns to the NameAtomsNbr Key the number of atoms based on the
// length of the map that is created via the Set or Decode methods.
//
// This method needs a Key in order to work. This Key is an instance of Header
// with Name equal to NameAtomsNbr. Use the Set method to assign this Key.
func (a *Atoms) SetKeysVal() error {
	if a.atomsNbr == nil {
		return fmt.Errorf("Key that is an instance of *Header with Name equal to NameAtomsNbr is nil: use the Set method")
	}
	return a.atomsNbr.Set(len(a.v))
}
