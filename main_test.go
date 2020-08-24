package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestBasic(t *testing.T) {

	Convey("Check what \"articles\" page is running correctly", t, func() {
		So(len(getArticles()), ShouldEqual, 1)
		Convey("It should return array with one item", nil)

	})

}
