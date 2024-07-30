package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/eyebluecn/tank/code/enums"
)

func TestEnums(t *testing.T) {
	fmt.Println(strings.Join(enums.DefaultVideo("").GetAllString(), "_"))
	for _, e := range enums.DefaultVideo("").GetAllString() {
		fmt.Println(e)
	}
}
