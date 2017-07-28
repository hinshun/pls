package namegen

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

var (
	names = []string{
		"mario",
		"luigi",
		"toad",
		"bowser",
		"waluigi",
		"wario",
		"peach",
	}
)

func GetRandomName() string {
	return names[rand.Intn(len(names))]
}
