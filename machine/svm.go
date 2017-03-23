package machine

import (
	"fmt"
	"math"
	"strconv"
)

type Fld interface {
	String() string
	UpdRowFloats(v uint8, nextix int, mat map[int]float64) (map[int]float64, int)
	RowFloatsToValue(nextix int, mat map[int]float64) (uint8, int)
	RowValueToValue(uint8) uint8
	ValueToRowValue(uint8) uint8
	RowValueToTxt(uint8) string
}

func ExtractMPosFld(fld Fld, mpos MPos) (v uint8) {
	mposFld, ok := fld.(MPosFld)
	if ok {
		v = mpos[mposFld.MposIx]
	} else {
		panic(fmt.Sprintf("Field %v have not been implemented"))
	}
	return v
}

func ExtractRow(mpos MPos, mMove MMove, flds []Fld) (y float64, x map[int]float64) {
	x = make(map[int]float64)
	if len(mMove) == 0 {
		y = 1
	} else {
		copy(mpos[len(mpos)-4:], mMove)
	}
	nextix := 1
	for _, fld := range flds {
		v := ExtractMPosFld(fld, mpos)
		x, nextix = fld.UpdRowFloats(v, nextix, x)
	}
	return y, x
}

type DomFld struct {
	Name string
	Domain
}

func (df *DomFld) String() string {
	return df.Name
}
func (df *DomFld) RowValueToTxt(v uint8) string {
	ix := df.RowValueToValue(v)
	return df.ValueToTxt(ix)
}
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
func (df *DomFld) ValueToRowValue(domval uint8) uint8 {

	domix := findDomIndex(domval, df.DomValues())
	if domix == -1 {
		panic(fmt.Sprintf("Value %v do not exist in domain %+v", domval, df.Domain))
	}
	return uint8(domix)
}
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
func (df *DomFld) UpdRowFloats(value uint8, nextix int, mat map[int]float64) (map[int]float64, int) {
	ix := findDomIndex(value, df.DomValues())
	if ix == -1 {
		panic(fmt.Sprintf("Fld \"%v\" Value %v do not exist in domain: %+v", df.Name, value, df.Domain))
	}
	mat[nextix+ix] = float64(1)

	return mat, nextix + len(df.DomValues())
}

type ValueFld struct {
	Name  string
	Scale float64
}

func (df *ValueFld) RowValueToTxt(v uint8) string {
	return strconv.Itoa(int(v))
}
func (vf *ValueFld) RowValueToValue(v uint8) uint8 {
	return v
}
func (vf *ValueFld) ValueToRowValue(v uint8) uint8 {
	return v
}
func (vf *ValueFld) String() string {
	return vf.Name
}
func (vf *ValueFld) UpdRowFloats(v uint8, nextix int, mat map[int]float64) (map[int]float64, int) {
	if v != 0 {
		mat[nextix] = floor(float64(v)/vf.Scale, 3)
	}
	return mat, nextix + 1
}
func (vf *ValueFld) RowFloatsToValue(nextix int, mat map[int]float64) (uint8, int) {
	v := uint8(roundInt(mat[nextix] * vf.Scale))
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
