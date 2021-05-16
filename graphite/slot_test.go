package graphite

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestJSON(t *testing.T) {
	p := Position{10, 20}
	b, _ := json.MarshalIndent(p, "", "\t")
	fmt.Println(string(b))
}
