package main

import "testing"

func TestData_GetWhenMatchWithRawData(t *testing.T) {
	result, err := data.GetWhenMatchWithRawData("plaz")
	if err != nil {
		t.Log(err)
	} else {
		t.Log(len(result))

	}
}
