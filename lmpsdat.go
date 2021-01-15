package lmpsdat

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/kpotier/lmpsdat/key"
)

// createNames returns a map that links the Names to the field identifiers of a
// structure and a map that links the Names to the corresponding Keys.
// lmpsdat:"Atoms" must include the Atom Style. For instance, it should be
// lmpsdat:"Atoms, full". If the Atom Style is not specified or does not exist,
// the Atom Style "full" will be used.
func createNames(typ reflect.Type) (map[key.Name]int, map[key.Name]key.Key) {
	atomStyle := key.AtomStyleFull
	names := make([]key.Name, 0)
	namesFields := make(map[key.Name]int, 0)
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		v, ok := f.Tag.Lookup("lmpsdat")
		if !ok {
			continue
		}
		if strings.HasPrefix(v, string(key.NameAtoms)) { // case where lmpsdat:"Atoms, ..."
			idx := strings.IndexRune(v, ',')
			if idx >= 0 && idx <= len(v) {
				as := strings.TrimSpace(v[idx+1:])
				if key.IsAtomStyle(as) {
					atomStyle = key.NewAtomStyle(as)
				} else {
					fmt.Fprintf(os.Stderr, "WARNING: atom style = %s is not supported", as)
				}
				v = strings.TrimSpace(v[:idx])
			}
		}
		n := key.Name(v)
		if key.IsName(n) {
			namesFields[n] = i
			names = append(names, n)
		} else {
			fmt.Fprintf(os.Stderr, "WARNING: name = %s is not supported", v)
		}
	}
	return namesFields, key.MakeKeys(names, atomStyle)
}

// headBody separate the keys. It reproduces what the LAMMPS data parser does.
func headBody(keys map[key.Name]key.Key) (headers, bodies map[key.Name]key.Key) {
	headers = make(map[key.Name]key.Key)
	bodies = make(map[key.Name]key.Key)
	for n, k := range keys {
		if key.IsHeader(k) {
			headers[n] = k
		} else {
			bodies[n] = k
		}
	}
	return
}

// keyDecode calls the Keyword method for several Keys. If a Keyword returns
// true, the Decode method will be called and this function will return true.
func keyDecode(s []byte, keys map[key.Name]key.Key, r *bufio.Scanner) (bool, error) {
	for n, k := range keys {
		if k.Keyword(s) {
			err := k.Decode(s, r)
			if err != nil {
				return true, fmt.Errorf("k.Decode for Key = %s: %w", k.Name(), err)
			}
			delete(keys, n)
			return true, nil
		}
	}
	return false, nil
}
