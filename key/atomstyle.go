package key

import (
	"fmt"
	"io"
	"strconv"
)

// AtomStyle is the style of atoms. This determines what attributes are
// associated with the atoms. For more information, please check the atom_style
// command in the LAMMPS documentation.
type AtomStyle interface {
	Name() string
	Encode(atom *Atom, w io.Writer) error
	Decode(f []string) (int, *Atom, error)
}

// The atom_style below are supported by this program. By default, the atom
// style is full (AtomStyleFull).
var (
	AtomStyleFull   AtomStyle = atomStyleFull("full")
	AtomStyleAtomic AtomStyle = atomStyleAtomic("atomic")
)

// ListAtomStyles is a list containing all the atom styles.
var ListAtomStyles []AtomStyle = []AtomStyle{
	AtomStyleFull,
	AtomStyleAtomic,
}

type atomStyleFull string

func (a atomStyleFull) Name() string {
	return string(a)
}

// Encode encodes the data for AtomStyleFull. It doesn't encode the N image
// sets.
func (a atomStyleFull) Encode(atom *Atom, w io.Writer) error {
	_, err := fmt.Fprintf(w, "%d %d %g %g %g %g", atom.MolTag, atom.AtomType, atom.Q, atom.X, atom.Y, atom.Z)
	return err
}

// Decode converts each column into a number (float64 or int) for the AtomStyleFull.
func (a atomStyleFull) Decode(f []string) (id int, atom *Atom, err error) {
	if len(f) < 7 {
		err = fmt.Errorf("not enough fields = %d, want >= 7", len(f))
		return
	}

	if id, err = strconv.Atoi(f[0]); err != nil {
		err = fmt.Errorf("strconv.Atoi id: %w", err)
		return
	}

	if atom.MolTag, err = strconv.Atoi(f[1]); err != nil {
		err = fmt.Errorf("strconv.Atoi MolTag: %w", err)
		return
	}
	if atom.AtomType, err = strconv.Atoi(f[2]); err != nil {
		err = fmt.Errorf("strconv.Atoi AtomType: %w", err)
		return
	}
	if atom.Q, err = strconv.ParseFloat(f[3], 64); err != nil {
		err = fmt.Errorf("strconv.ParseFloat Q: %w", err)
		return
	}
	if atom.X, err = strconv.ParseFloat(f[4], 64); err != nil {
		err = fmt.Errorf("strconv.ParseFloat X: %w", err)
		return
	}
	if atom.Y, err = strconv.ParseFloat(f[5], 64); err != nil {
		err = fmt.Errorf("strconv.ParseFloat Y: %w", err)
		return
	}
	if atom.Z, err = strconv.ParseFloat(f[6], 64); err != nil {
		err = fmt.Errorf("strconv.ParseFloat Z: %w", err)
		return
	}

	atom.N = false
	if len(f) == 10 {
		atom.N = true
		if atom.NX, err = strconv.Atoi(f[7]); err != nil {
			err = fmt.Errorf("strconv.Atoi NX: %w", err)
			return
		}
		if atom.NY, err = strconv.Atoi(f[8]); err != nil {
			err = fmt.Errorf("strconv.Atoi NY: %w", err)
			return
		}
		if atom.NZ, err = strconv.Atoi(f[9]); err != nil {
			err = fmt.Errorf("strconv.Atoi NZ: %w", err)
			return
		}
	}

	return
}

type atomStyleAtomic string

func (a atomStyleAtomic) Name() string {
	return string(a)
}

// Encode encodes the data for AtomStyleAtomic. It doesn't encode the N image
// sets.
func (a atomStyleAtomic) Encode(atom *Atom, w io.Writer) error {
	_, err := fmt.Fprintf(w, "%d %g %g %g", atom.AtomType, atom.X, atom.Y, atom.Z)
	return err
}

// Decode converts each column into a number (float64 or int) for the atomStyleAtomic.
func (a atomStyleAtomic) Decode(f []string) (id int, atom *Atom, err error) {
	if len(f) < 5 {
		err = fmt.Errorf("not enough fields = %d, want >= 5", len(f))
		return
	}

	if id, err = strconv.Atoi(f[0]); err != nil {
		err = fmt.Errorf("strconv.Atoi id: %w", err)
		return
	}

	if atom.AtomType, err = strconv.Atoi(f[1]); err != nil {
		err = fmt.Errorf("strconv.Atoi AtomType: %w", err)
		return
	}

	if atom.X, err = strconv.ParseFloat(f[2], 64); err != nil {
		err = fmt.Errorf("strconv.ParseFloat X: %w", err)
		return
	}
	if atom.Y, err = strconv.ParseFloat(f[3], 64); err != nil {
		err = fmt.Errorf("strconv.ParseFloat Y: %w", err)
		return
	}
	if atom.Z, err = strconv.ParseFloat(f[4], 64); err != nil {
		err = fmt.Errorf("strconv.ParseFloat Z: %w", err)
		return
	}

	atom.N = false
	if len(f) == 8 {
		atom.N = true
		if atom.NX, err = strconv.Atoi(f[5]); err != nil {
			err = fmt.Errorf("strconv.Atoi NX: %w", err)
			return
		}
		if atom.NY, err = strconv.Atoi(f[6]); err != nil {
			err = fmt.Errorf("strconv.Atoi NY: %w", err)
			return
		}
		if atom.NZ, err = strconv.Atoi(f[7]); err != nil {
			err = fmt.Errorf("strconv.Atoi NZ: %w", err)
			return
		}
	}

	return
}
