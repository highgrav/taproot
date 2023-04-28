package forms

import (
	"fmt"
	"testing"
)

func TestFromMap(t *testing.T) {
	type MyTestForm struct {
		TestString string `json:"test_string"`
		TestInt    int32  `json:"test_int"`
	}

	testMap := map[string][]string{
		"test_string": []string{"my_test_string"},
		"test_int":    []string{"32"},
	}
	elem, err := FromMap[MyTestForm](testMap)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%d\n", elem.TestInt)

	if elem.TestInt != 32 || elem.TestString != "my_test_string" {
		t.Error("failed to properly set struct values")
	}
}
