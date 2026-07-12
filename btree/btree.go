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


// На позиции последнего оффсета размер ноды в байтах
func (node BNode) nbytes() uint16 {
	return node.keyValuePosition(node.bnode_keys_count())
}




// Меньше или равный ключ ищем
func LookupKeyLE (node BNode, key []byte) uint16 {
	nkeys := node.bnode_keys_count()
	
	found_key := uint16(0)
	for i := uint16(1); i < nkeys; i++ {
		cmp := bytes.Compare(node.get_key(i), key)
		if (cmp <= 0) {
			found_key = i
		}
		if (cmp >= 0) {
			break
		}

	}
	return found_key
}


// Функция Insert в Node новый Key/value, предназначена как для листьев, так и для внутренних узлов
func nodeAppendKV(node BNode, idx uint16, ptr uint64, key []byte, value []byte) {
	node.set_pointer(idx, ptr)
	
	key_position := node.keyValuePosition(idx)
	node.set_header(2, node.bnode_keys_count() + uint16(1))
	binary.LittleEndian.PutUint16(node[key_position:], uint16(len(key)))
	binary.LittleEndian.PutUint16(node[key_position + 2:], uint16(len(value)))

	copy(node[key_position + 4 :], key)
	copy(node[key_position + 4 + uint16(len(key)):], value)

	offset,_ := node.get_offset(idx)
	node.set_offset(idx + 1, offset + 4 + uint16(len(key) + len(value)))
}

// Вставка нескольких элементов со второй ноды (старая) в первую (новая). dstNew - куда начать вставку в новой ноде, srcOld - откуда брать элементы в старой. n - число элементов 
func nodeAppendRange(new BNode, old BNode, dstNew uint16, srcOld uint16, n uint16) {
	insertIndex := dstNew 
	for i := srcOld; i < n + srcOld; i++ {
		key := old.get_key(i)
		val := old.get_value(i)
		nodeAppendKV(new, insertIndex, 0, key, val)
		insertIndex++
	}
}


// Copy on Write вставка
func leafInsert(new BNode, old BNode, idx uint16, key []byte, value []byte) {
	new.set_header(2, old.bnode_keys_count() + 1)
	nodeAppendRange(new, old, 0, 0, idx)
	nodeAppendKV(new, idx, 0, key, value)
	nodeAppendRange(new, old, idx+1, idx, old.bnode_keys_count() - idx)
}