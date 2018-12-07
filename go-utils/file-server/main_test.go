package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_main(t *testing.T) {
	// table driven tests are recommended
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		// use go convey for better readability
		Convey(tt.name, t, func() {
			main()
		})
	}
}
