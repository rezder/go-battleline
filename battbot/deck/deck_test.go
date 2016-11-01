package deck

import (
	"testing"
)

func TestMaxValues(t *testing.T) {
	troops := make(map[int]bool)
	troops[10] = true
	troops[19] = true
	troops[37] = true
	troops[28] = true
	troops[6] = true
	scoutReturnTroops := make([]int, 2)
	scoutReturnTroops[0] = 40
	scoutReturnTroops[1] = 17
	testMaxValueCheck(t, troops, scoutReturnTroops, 36)

}
func TestMaxValuesMax(t *testing.T) {
	troops := make(map[int]bool)
	troops[10] = true
	troops[20] = true
	troops[30] = true
	troops[28] = true
	troops[3] = true
	scoutReturnTroops := make([]int, 2)
	scoutReturnTroops[0] = 40
	scoutReturnTroops[1] = 17
	testMaxValueCheck(t, troops, scoutReturnTroops, 38)
}
func testMaxValueCheck(t *testing.T, troops map[int]bool, scr []int, exp int) {
	values := maxValues(troops, scr)
	t.Logf("Values: %v", values)
	sum := 0
	for _, v := range values {
		sum = sum + v
	}
	if sum != exp {
		t.Errorf("Expected: %v got %v", exp, sum)
	}
}
