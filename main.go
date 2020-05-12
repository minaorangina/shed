package main

import (
	"github.com/minaorangina/shed/gameengine"
)

func main() {
	ge := gameengine.New()
	ge.Init([]string{"Harry", "Sally"})
}
