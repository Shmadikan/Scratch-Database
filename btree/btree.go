package btree

import (
	"encoding/binary"
	"errors"
)

type BNode []byte

const HEADER = 4


// Node представлена как последовательность байтов
type Btree struct {
	root_number_page uint64
	get_node func(uint64) BNode
	new func(BNode) uint64
	delete_node func(uint64)
}


func (node BNode) bnode_type() uint16 {
	return binary.LittleEndian.Uint16(node[0:2])
}

func (node BNode) bnode_keys_count() uint16 {
	return binary.LittleEndian.Uint16(node[2:4])
}

func (node BNode) set_header(type_node uint16, keys_len uint16) {
	binary.LittleEndian.PutUint16(node[0:2], type_node)
	binary.LittleEndian.PutUint16(node[2:4], keys_len)
}

func (node BNode) child_pointer(idx uint16) (uint64, error) {
	if (idx > node.bnode_keys_count()) {
		return 0, errors.New("Not match keys")
	}

	child_ptr := HEADER + idx*8
	return binary.LittleEndian.Uint64(node[child_ptr:child_ptr+9]), nil
} 

func (node BNode) set_pointer(idx uint16, ptr uint64) error {
	if (idx > node.bnode_keys_count()) {
		return errors.New("Index ouf of range")
	}
	child_ptr := HEADER + idx*8
	binary.LittleEndian.PutUint64(node[child_ptr:child_ptr+9], ptr)
	return nil
}

func (node BNode) get_offset(idx uint16) (uint16, error) {
	if (idx > node.bnode_keys_count()){
		return 0, errors.New("Index out of range")
	}

	offset := HEADER + node.bnode_keys_count()*8 + 2*idx
	return binary.LittleEndian.Uint16(node[offset:]), nil
}
