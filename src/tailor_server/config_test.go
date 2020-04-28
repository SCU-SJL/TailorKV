package tailor_server

import (
	"fmt"
	"testing"
)

func TestGetConfig(t *testing.T) {
	fmt.Println(GetConfig("config.xml"))
}
