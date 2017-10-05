package login

import (
	"fmt"
)

var (
	//StatusAll is the log in status domain.
	StatusAll StatusAllST
)

func init() {
	StatusAll = newStatusAllST()
}

//StatusAllST is the log in status singleton.
type StatusAllST struct {
	None     Status
	Ok       Status
	Down     Status
	InValid  Status
	Exist    Status
	Disabled Status
	Err      Status
}

func newStatusAllST() (l StatusAllST) {
	l.None = 0
	l.Ok = 1
	l.Down = 2
	l.InValid = 3
	l.Disabled = 4
	l.Exist = 5
	l.Err = 6
	return l
}

//Status the log in status domain value.
type Status int

func (l Status) String() (txt string) {
	switch l {
	case StatusAll.None:
		txt = "None"
	case StatusAll.Ok:
		txt = "OK"
	case StatusAll.Down:
		txt = "Games server is down."
	case StatusAll.InValid:
		txt = "Crediential is invalid"
	case StatusAll.Disabled:
		txt = "Account disabled"
	case StatusAll.Exist:
		txt = "Double access"
	default:
		panic(fmt.Sprintf("Login status: %v does not exist ", int(l)))
	}
	return txt
}
func (l Status) IsOk() bool {
	return l == StatusAll.Ok
}
func (l Status) IsDown() bool {
	return l == StatusAll.Down
}
func (l Status) IsInValid() bool {
	return l == StatusAll.InValid
}
func (l Status) IsExist() bool {
	return l == StatusAll.Exist
}
func (l Status) IsDisable() bool {
	return l == StatusAll.Disabled
}
