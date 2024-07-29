package main

import (
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"testing"
)

func Test(t *testing.T) {
	fooKey = attribute.Key("ex.com/foo")
	fmt.Println(rune('/'))
	fmt.Println(string(fooKey)[6])
}
