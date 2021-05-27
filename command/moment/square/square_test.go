package square

import (
	"testing"
)

func TestSquare(t *testing.T) {
	square := &AcFunSquare{}
	moments, err := square.GetMoment()
	if err != nil {
		t.Fatal(err)
	}
	if len(moments) == 0 {
		t.Fatal("failed to get acfun moment square")
	}
	t.Logf("%s", moments[0].ToString())
}
