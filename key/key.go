package key

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"reflect"
	"sort"
	"unicode"
)

// Key is an interface that allows to decode or encode specific data from LAMMPS
// data file. For instance, a Key can be used to decode the Masses part.
type Key interface {
	// Name returns the identifier or keyword of the Key.
	Name() Name
	// Keyword analyzes a line. For instance, it will analyze the line
	// containing "Masses". If the Key correponds to the Masses part, Keyword
	// will return true and the next lines can be safely decoded with the Decode
	// method. If Keyword returns false, meaning that the Key is not able to
	// decode the next lines of "Masses", Decode should not be called.
	Keyword([]byte) bool

	// SetKeys sets one or several Keys to a Key. For instance, the Masses Key
	// requires the Atoms Key in order to Decode and Encode safely the data. To
	// know what are the required Keys, please call the NameOfReqKeys method.
	SetKeys(...Key) error

	// SetKeysVal calls the Set method for the Keys that were set via the
	// SetKeys method. For instance, for Masses, it will assign to the
	// NameAtomTypes Key the number of atom types based on the length of the map
	// that was created via the Set or Decode methods.
	SetKeysVal() error

	// Encode writes the data, for instance a list of masses corresponding to
	// each atom type. It will also writes the "header" of this list
	// ("Masses\n\n"). Generally, this method will call the Get method.
	Encode(io.Writer) error
	// Decode analyzes the first line that went through the Keyword method and
	// the next lines. Generally, this method will call the Set method. Do not
	// call Decode before calling Keyword or errors will be produced.
	Decode([]byte, *bufio.Scanner) error

	// Set sets a specific data into Key. For instance, in Masses, interface{}
	// should be a map where the keys are the atom types and the values are the
	// masses.
	Set(interface{}) error

	// Get returns the data stored into Key. For instance, in Masses, it will
	// return a map where the keys are the atom types and the values are the
	// masses.
	Get() interface{}

	// Check verifies if the data is correct. For instance, for Masses, this
	// methods verifies if the masses are not lower than zero.
	Check() error
}

// Name is a unique identifier. It characterizes each Key and is case-sensitive.
type Name string

const (
	// NameAtomsNbr is the Name related to the number of atoms.
	NameAtomsNbr Name = "atoms"
	// NameBondsNbr is the Name related to the number of bonds.
	NameBondsNbr Name = "bonds"
	// NameAnglesNbr is the Name related to the number of angles.
	NameAnglesNbr Name = "angles"
	// NameDihedralsNbr is the Name related to the number of dihedrals.
	NameDihedralsNbr Name = "dihedrals"

	// NameAtomTypes is the Name related to the number of atom types.
	NameAtomTypes Name = "atom types"
	// NameBondTypes is the Name related to the number of bond types.
	NameBondTypes Name = "bond types"
	// NameAngleTypes is the Name related to the number of angle types.
	NameAngleTypes Name = "angle types"
	// NameDihedralTypes is the Name related to the number of dihedral types.
	NameDihedralTypes Name = "dihedral types"

	// NameBoxX is the Name related to the size of the box for the x coordinate.
	NameBoxX Name = "xlo xhi"
	// NameBoxY is the Name related to the size of the box for the y coordinate.
	NameBoxY Name = "ylo yhi"
	// NameBoxZ is the Name related to the size of the box for the z coordinate.
	NameBoxZ Name = "zlo zhi"

	// NameMasses is the Name related to the masses table (1st column: atom
	// type, 2nd column: mass).
	NameMasses Name = "Masses"

	// NamePairCoeffs is the Name related to the Pair Coeffs table (1st column: atom
	// type, other columns: depend on pair_style)
	NamePairCoeffs Name = "Pair Coeffs"
	// NameBondCoeffs is the Name related to the Bond Coeffs table (1st column:
	// bond type, other columns: related to bond_style).
	NameBondCoeffs Name = "Bond Coeffs"
	// NameAngleCoeffs is the Name related to the Angle Coeffs table (1st
	// column: angle type, other columns: depend on angle_style).
	NameAngleCoeffs Name = "Angle Coeffs"
	// NameDihedralCoeffs is the Name related to the Dihedral Coeffs table (1st
	// column: dihedral type, other columns: depend on dihedral_style).
	NameDihedralCoeffs Name = "Dihedral Coeffs"

	// NameAtoms is the Name related to the Atoms table. In order: atom number,
	// molecule number, atom type, charge, x, y, z, nx, ny, and nz. The
	// parameters nx, ny, and nz are optional.
	NameAtoms Name = "Atoms"
	// NameBonds is the Name related to the Bonds table. 1st column: bond
	// number, second column: bond type, third: atom 1, fourth: atom 2.
	NameBonds Name = "Bonds"
	// NameAngles is the Name related to the Angles table. 1st column: angle
	// number, second column: angle type, third: atom 1, fourth: atom 2, fifth:
	// atom 3.
	NameAngles Name = "Angles"
	// NameDihedrals is the Name related to the Dihedrals table. 1st column:
	// dihedral number, second column: dihedral type, third: atom 1, fourth:
	// atom 2, fifth: atom 3, sixth: atom 4.
	NameDihedrals Name = "Dihedrals"

	// NameTitle is the Name related to the title of the LAMMPS data file. It is
	// located at the first line of the file.
	NameTitle Name = "Title"
)

// ListNames is a list containing all the Names.
var ListNames []Name = []Name{
	NameAngleCoeffs,
	NameAngleTypes,
	NameAngles,
	NameAnglesNbr,
	NameAtomTypes,
	NameAtoms,
	NameAtomsNbr,
	NameBondCoeffs,
	NameBondTypes,
	NameBonds,
	NameBondsNbr,
	NameBoxX,
	NameBoxY,
	NameBoxZ,
	NameDihedralCoeffs,
	NameDihedralTypes,
	NameDihedrals,
	NameDihedralsNbr,
	NameMasses,
	NamePairCoeffs,
	NameTitle,
}

// ErrUnsupported is an error return if a feature is unsupported by a Key.
var ErrUnsupported error = errors.New("unsupported")

// delComments deletes everything that is after "#".
func delComments(s []byte) []byte {
	if idx := bytes.IndexRune(s, '#'); idx != -1 {
		s = s[:idx]
	}
	return s
}

// sortIntsMap returns the keys sorted in increasing order. If the keys are not
// int or m is not a map, this method will panic.
func sortIntsMap(m interface{}) (keys []int) {
	val := reflect.ValueOf(m)
	for _, k := range val.MapKeys() {
		keys = append(keys, k.Interface().(int))
	}
	sort.Ints(keys)
	return
}

// keywordHeader tests whether the byte slice s begins with prefix after
// trimming the spaces and after a number.
func keywordHeader(s []byte, prefix [][]byte) bool {
	if len(prefix) == 0 {
		return false
	}
	s = bytes.TrimLeftFunc(s, unicode.IsSpace)
	if !bytes.HasPrefix(s, prefix[0]) {
		return false
	}
	if len(prefix) > 1 {
		for _, p := range prefix[1:] {
			idx := bytes.IndexFunc(s, unicode.IsSpace)
			if idx == -1 {
				return false
			}
			s = bytes.TrimLeftFunc(s[idx:], unicode.IsSpace)
			if !bytes.HasPrefix(s, p) {
				return false
			}
		}
	}
	return true
}

// keyword tests whether the byte slice s begins with prefix after trimming the
// spaces.
func keyword(s, prefix []byte) bool {
	s = bytes.TrimLeftFunc(s, unicode.IsSpace)
	return bytes.HasPrefix(s, prefix)
}
