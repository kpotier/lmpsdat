package key

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Links is used to encode and/or decode a table containing the links (e.g.
// Bonds, Angles) from a LAMMPS data file. This table has a header where a blank
// line separate the values from it. Each value (= 1 line) has 4 or more columns
// (the first one is the identifier, the second is the type, the third is atom1
// linked to atom2 (the fourth), and so on...). More information about the
// structure of this table can be found in the LAMMPS documentation.
//
// Links can be instanced by using the NewLinks method.
type Links struct {
	name  Name
	links int

	nbr      *Header
	types    *Header
	atomsNbr *Header
	v        map[int]*Link
}

// Link contains the type (e.g. bond type number 1) and the links (e.g. atom1
// linked to atom2). The identifier is not included in it.
type Link struct {
	typ   int
	links []int
}

// NewLinks returns an instance of Links. If links is equal to 2, then the
// number of colums must be equal to 4 (1 identifier, 1 type, and 2 atoms).
func NewLinks(name Name, links int) *Links {
	return &Links{name: name, links: links + 2}
}

// Name returns the Name passed in NewLinks. It corresponds to the header of the
// table.
func (l *Links) Name() Name {
	return l.name
}

// Keyword tests whether the byte slice s begins with Name after trimming the
// spaces. Keyword is useful to detect the header of the Coeffs table.
func (l *Links) Keyword(s []byte) bool {
	return keyword(s, []byte(l.Name()))
}

// SetKeys assigns one or more Keys to Links. This method only accepts *Header.
// One key must have a Name equal to NameAtomsNbr, another must have a suffix
// equal to "types" (e.g. bond types). Other Headers are considered as the
// number of values (e.g. BondsNbr).
func (l *Links) SetKeys(k ...Key) error {
	for _, key := range k {
		header, ok := key.(*Header)
		if !ok {
			return fmt.Errorf("type assertion error: key provided is not *Header")
		}
		if header.Name() == NameAtomsNbr {
			l.atomsNbr = header
		} else if strings.HasSuffix(string(header.Name()), "types") {
			l.types = header
		} else {
			l.nbr = header
		}
	}
	return nil
}

// Encode writes a table containing the header, a blank line and each value (= 1
// line) into a writer.
//
// This method does not check the integrity and correctness of each value. To do
// so, use the Check method.
func (l *Links) Encode(w io.Writer) error {
	if l.v == nil {
		return fmt.Errorf("map[int]*Link is nil: use the Decode or Set methods")
	}
	if len(l.v) == 0 {
		return nil
	}

	keys := sortIntsMap(l.v)
	fmt.Fprint(w, l.Name(), "\n\n")
	for _, k := range keys {
		link := l.v[k]
		if _, err := fmt.Fprintf(w, "%d %d", k, link.typ); err != nil {
			return fmt.Errorf("fmt.Fprintf: %w", err)
		}
		for _, v := range link.links {
			if _, err := fmt.Fprintf(w, " %d", v); err != nil {
				return fmt.Errorf("fmt.Fprintf link: %w", err)
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
// an instance of Link that is put into a map where the keys are the identifiers
// of each value.
//
// This method needs a Key in order to work. This Key is an instance of Header
// that represent the number of values (e.g. NameBondsNbr). Use the Set method
// to assign this Key.
//
// Moreover, this method does not check the integrity and corectness of the
// values decoded. To do so, use the Check method.
//
// Decode method does not return io.EOF error. The use of the Check method after
// Decode is therefore highly recommended.
func (l *Links) Decode(s []byte, r *bufio.Scanner) error {
	if l.nbr == nil {
		return fmt.Errorf("Key that is an instance of *Header that represent the number of values is nil: use the Set method")
	}

	types := l.nbr.Get().(int)
	l.v = make(map[int]*Link)

	if ok := r.Scan(); !ok {
		if r.Err() != nil {
			return fmt.Errorf("r.Scan first line: %w", r.Err())
		}
		return nil
	}

	for i := 0; i < types && r.Scan(); i++ {
		f := strings.Fields(r.Text())
		if len(f) < l.links {
			return fmt.Errorf("not enough fields = %d, want >= %d", len(f), l.links)
		}

		id, err := strconv.Atoi(f[0])
		if err != nil {
			return fmt.Errorf("strconv.Atoi id: %w", err)
		}

		typ, err := strconv.Atoi(f[1])
		if err != nil {
			return fmt.Errorf("strconv.Atoi type: %w", err)
		}

		var links []int
		for _, v := range f[2:l.links] {
			atom, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("strconv.Atoi link: %w", err)
			}
			links = append(links, atom)
		}
		l.v[id] = &Link{typ: typ, links: links}
	}
	if r.Err() != nil {
		return fmt.Errorf("r.Scan: %w", r.Err())
	}
	return nil
}

// Set puts a custom map[int]*Link.
//
// This method does not check the integrity or correctness of the passed data.
// The use of the Check method after Set is therefore highly recommended.
func (l *Links) Set(v interface{}) error {
	var ok bool
	l.v, ok = v.(map[int]*Link)
	if !ok {
		return fmt.Errorf("type assertion error: value is not map[int]*Link")
	}
	return nil
}

// Get returns a map[int]*Link where the keys are the identifiers of the links.
// As this method returns an interface, it must be useful to perform a type
// assertion after calling this method.
func (l *Links) Get() interface{} {
	return l.v
}

// Check verifies the integrity and correctness of the data decoded with the
// Decode method or set with the Set method.
//
// This method needs three Keys in order to work. The first Key is the number of
// types, the second is the number of atoms, and the third is the number of
// values (identifiers).
func (l *Links) Check() error {
	if l.types == nil || l.atomsNbr == nil || l.nbr == nil {
		return fmt.Errorf("one or more Keys are nil: use the Set method")
	}

	nbr := l.nbr.Get().(int)
	types := l.types.Get().(int)
	atomsNbr := l.atomsNbr.Get().(int)

	if len(l.v) != nbr {
		return fmt.Errorf("number of assigned values (ids) = %d is not equal to the number of expected values = %d", len(l.v), nbr)
	}

	for id, link := range l.v {
		if id < 1 || id > nbr {
			return fmt.Errorf("id = %d is invalid: it must be greater than zero and lower or equal than the number of id = %d", id, nbr)
		}
		if link.typ < 1 || link.typ > types {
			return fmt.Errorf("type = %d is invalid: it must be greater than zero and lower or equal than the number of types = %d", id, nbr)
		}
		for _, atom := range link.links {
			if atom < 1 || atom > atomsNbr {
				return fmt.Errorf("atom = %d is invalid: it must be greater than zero and lower or equal than the number of atoms = %d", id, nbr)
			}
		}
	}
	return nil
}

// SetKeysVal assigns to the NamexxxNbr Key the number of bonds, dihedrals,
// angle, etc. based on the length of the map that is created via the Set or
// Decode methods.
//
// This method needs a Key in order to work. This Key is an instance of Header
// that represent the number of values (e.g. NameBondsNbr). Use the Set method
// to assign this Key.
func (l *Links) SetKeysVal() error {
	if l.nbr == nil {
		return fmt.Errorf("Key that is an instance of *Header that represent the number of values is nil: use the Set method")
	}
	return l.nbr.Set(len(l.v))
}
