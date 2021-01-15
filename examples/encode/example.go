package main

import (
	"os"

	"github.com/kpotier/lmpsdat"
	"github.com/kpotier/lmpsdat/key"
)

type example struct {
	Title string     `lmpsdat:"Title"`
	VolX  [2]float64 `lmpsdat:"xlo xhi"`
	VolY  [2]float64 `lmpsdat:"ylo yhi"`
	VolZ  [2]float64 `lmpsdat:"zlo zhi"`

	Atoms  map[int]*key.Atom `lmpsdat:"Atoms, atomic"` // atom style is atomic
	Masses map[int]float64   `lmpsdat:"Masses"`
}

func main() {
	var example example
	example.Title = "My title"
	example.VolX = [2]float64{0.0, 1.0}
	example.VolY = [2]float64{0.0, 1.0}
	example.VolZ = [2]float64{0.0, 1.0}

	example.Masses = map[int]float64{
		1: 1.0,
	}

	example.Atoms = map[int]*key.Atom{
		1: {
			MolTag:   1,
			AtomType: 1,
			X:        0.5,
			Y:        0.5,
			Z:        0.5,
		},
	}

	enc := lmpsdat.NewEncoder(os.Stdout)
	err := enc.Encode(&example)
	if err != nil {
		panic(err)
	}
}
