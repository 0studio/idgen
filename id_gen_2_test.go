package idgen

import (
	"fmt"
	"testing"
)

func TestGetId(t *testing.T) {
	idGenerator := NewIdGenerator(1, 1)
	id := (idGenerator).GetId()
	fmt.Println(id)
	id = (idGenerator).GetId()
	fmt.Println(id)
}
