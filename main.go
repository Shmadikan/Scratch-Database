package main

import (
	"db/btree"
	"encoding/binary"
	"fmt"
)

type person struct {
	name string
	age  int
}

func (p person) greet() string {

	return "Hello, my name is " + p.name
}

func (p *person) haveBirthday() int {
	p.age++
	return p.age
}

func main() {
	mmap := make(map[string]int)
	mmap["3"] = 1
	mmap["3"]--
	
	fmt.Println(binary.LittleEndian)
}
