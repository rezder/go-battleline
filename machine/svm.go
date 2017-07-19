package machine

import (
	"fmt"
	"math"
	"strconv"
)

//Fld a feature or a label in machine learning.
type Fld interface {
	String() string
	UpdRowFloats(v uint8, nextix int, mat map[int]float64) (map[int]float64, int)
	RowFloatsToValue(nextix int, mat map[int]float64) (uint8, int)
	RowValueToValue(uint8) uint8
	ValueToRowValue(uint8) uint8
	RowValueToTxt(uint8) string
}

//ExtractMPosFld extracts field value from a machine position.
func ExtractMPosFld(fld Fld, mpos MPos) (v uint8) {
	mposFld, ok := fld.(MPosFld)
	if ok {
		v = mpos[mposFld.MposIx]
	} else {
		panic(fmt.Sprintf("Field %v have not been implemented", fld))
	}
	return v
}

// ExtractSparseRow creates a sparse one-hot row.
func ExtractSparseRow(mpos MPos, flds []Fld) (x map[int]float64) {
	x = make(map[int]float64)
	nextix := 1
	for _, fld := range flds {
		v := ExtractMPosFld(fld, mpos)
		x, nextix = fld.UpdRowFloats(v, nextix, x)
	}
	return x
}

// ExtractRow create a row.
func ExtractRow(mpos MPos, flds []Fld) (x []uint8) {
	x = make([]uint8, len(flds))
	for i, fld := range flds {
		v := ExtractMPosFld(fld, mpos)
		x[i] = v
	}
	return x
}

// DomFld a sparse field.
type DomFld struct {
	Name string
	Domain
}

func (df *DomFld) String() string {
	return df.Name
}

//RowValueToTxt translates from the domain index to to text.
func (df *DomFld) RowValueToTxt(v uint8) string {
	ix := df.RowValueToValue(v)
	return df.ValueToTxt(ix)
}

//RowValueToValue translates from the domain index to domain value.
func (df *DomFld) RowValueToValue(ix uint8) (domval uint8) {

	if ix < 0 && int(ix) >= len(df.DomValues()) {
		panic(fmt.Sprintf("Index %v do not exist in domain %+v", ix, df.Domain))
	}
	domval = df.DomValues()[int(ix)]
	return domval
}
func findDomIndex(value uint8, values []uint8) (ix int) {
	ix = -1
	for i, v := range values {
		if v == value {
			ix = i
			break
		}
	}
	return ix
}

// ValueToRowValue translate from domain value to domain index.
func (df *DomFld) ValueToRowValue(domval uint8) uint8 {
	domix := findDomIndex(domval, df.DomValues())
	if domix == -1 {
		panic(fmt.Sprintf("Value %v do not exist in domain %+v", domval, df.Domain))
	}
	return uint8(domix)
}

// RowFloatsToValue reads domain value from on hot sparse matrix.
func (df *DomFld) RowFloatsToValue(nextix int, mat map[int]float64) (uint8, int) {
	ix := -1
	for i := nextix; i < nextix+len(df.DomValues()); i++ {
		if mat[i] != 0 {
			ix = i
			break
		}
	}
	if ix == -1 {
		panic("No value sat")
	}
	domvalue := df.DomValues()[ix-nextix]
	return domvalue, nextix + len(df.DomValues())
}

// UpdRowFloats update a on-hot sparse row  with domain value.
func (df *DomFld) UpdRowFloats(value uint8, nextix int, mat map[int]float64) (map[int]float64, int) {
	ix := findDomIndex(value, df.DomValues())
	if ix == -1 {
		panic(fmt.Sprintf("Fld \"%v\" Value %v do not exist in domain: %+v", df.Name, value, df.Domain))
	}
	mat[nextix+ix] = float64(1)

	return mat, nextix + len(df.DomValues())
}

// ValueFld a field with values.
type ValueFld struct {
	Name  string
	Scale float64
}

// RowValueToTxt translate row value to text.
func (vf *ValueFld) RowValueToTxt(v uint8) string {
	return strconv.Itoa(int(v))
}

// RowValueToValue translate row value to value.
// It does nothing.
func (vf *ValueFld) RowValueToValue(v uint8) uint8 {
	return v
}

// ValueToRowValue translate value to row value.
// It does nothing.
func (vf *ValueFld) ValueToRowValue(v uint8) uint8 {
	return v
}
func (vf *ValueFld) String() string {
	return vf.Name
}

// UpdRowFloats update a one-hot sparse row with a value.
func (vf *ValueFld) UpdRowFloats(v uint8, nextix int, mat map[int]float64) (map[int]float64, int) {
	if v != 0 {
		mat[nextix] = floor(float64(v)/vf.Scale, 3)
	}
	return mat, nextix + 1
}

// RowFloatsToValue reads value from one-hot sparse matrix
func (vf *ValueFld) RowFloatsToValue(nextix int, mat map[int]float64) (uint8, int) {
	v := roundInt(mat[nextix] * vf.Scale)
	return v, nextix + 1
}
func roundInt(v float64) uint8 {
	if v < 0 {
		return uint8(math.Ceil(v - 0.5))
	}
	return uint8(math.Floor(v + 0.5))
}
func floor(v float64, prec int) float64 {
	pow := math.Pow10(prec)
	v = v * pow
	if v < 0 {
		v = math.Ceil(v)
	} else {
		v = math.Floor(v)
	}
	return v / pow
}
