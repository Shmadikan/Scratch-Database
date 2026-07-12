package btree

import (
	"encoding/binary"
	"errors"
	"bytes"
)

type BNode []byte

const HEADER = 4
const IndexERROR = "Index out of range"

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
		return errors.New(IndexERROR)
	}
	child_ptr := HEADER + idx*8
	binary.LittleEndian.PutUint64(node[child_ptr:child_ptr+9], ptr)
	return nil
}



func offsetPos(node BNode, idx uint16) (uint16, error) {
	if idx < 1 || idx > node.bnode_keys_count() {
		return 0, errors.New(IndexERROR)
	}
	return HEADER + 8*node.bnode_keys_count() + 2*(idx-1), nil
}


// декодировать значение n-го смещение
func (node BNode) get_offset(idx uint16) (uint16, error) {
	if (idx > node.bnode_keys_count()){
		return 0, errors.New(IndexERROR)
	}
	if (idx == 0) {
		return 0, nil
	}

	offset,_ := offsetPos(node, idx)
	return binary.LittleEndian.Uint16(node[offset:]), nil
}

// Установить значение для n-го смещения
func (node BNode) set_offset(idx uint16, offset_value uint16) error {
	if (idx > node.bnode_keys_count()) {
		return errors.New(IndexERROR)
	}
	offset, _ := offsetPos(node, idx)
	binary.LittleEndian.PutUint16(node[offset:], offset_value)
	return nil
}


// Получить значение байт смещения для ключа по индексу смещения
func (node BNode) keyValuePosition(idx uint16) uint16 {
	offset, error := node.get_offset(idx)
	if (error != nil) {
		return 0
	}
	return HEADER + 8*node.bnode_keys_count() + 2*node.bnode_keys_count() + offset

}


// Получить ключ по смещению
func (node BNode) get_key(idx uint16) []byte {
	offset := node.keyValuePosition(idx)
	keylen := binary.LittleEndian.Uint16(node[offset:])
	return node[offset+4:][:keylen]
}

func (node BNode) get_value(idx uint16) []byte {
	offset := node.keyValuePosition(idx)
	value_len := binary.LittleEndian.Uint16(node[(offset + 2):])
	keylen := binary.LittleEndian.Uint16(node[offset:])
	return node[offset+4+keylen:][:value_len]
}