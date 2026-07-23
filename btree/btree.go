package btree

import (
	"encoding/binary"
	"errors"
	"bytes"
)

type BNode []byte

const HEADER = 4
const PTR_SIZE = 8
const OFFSET_SIZE = 2
const PAGE_SIZE = 4096
const IndexERROR = "Index out of range"
const LEAF = 2
const INTERNAL_NODE = 1

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


// Получить ребёнка ноды
func (node BNode) get_kidPointer(idx uint16) (uint64, error) {
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
	if idx == 0 {
		return 0, nil
	}
	if idx > node.bnode_keys_count() {
		return 0, errors.New(IndexERROR)
	}

	offset, _ := offsetPos(node, idx)
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


// Получить абсолютную позицию KV 
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
	
	key_index := uint16(0)
	for i := uint16(1); i < nkeys; i++ {
		cmp := bytes.Compare(node.get_key(i), key)
		if (cmp <= 0) {
			key_index = i
		}
		if (cmp >= 0) {
			break
		}

	}
	return key_index
}


// Функция Insert в Node новый Key/value, предназначена как для листьев, так и для внутренних узлов
func nodeAppendKV(node BNode, idx uint16, ptr uint64, key []byte, value []byte) {
	node.set_pointer(idx, ptr)
	
	key_position := node.keyValuePosition(idx)
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


// Вставка во внутренний узел
func NkidInsert(tree *Btree, new BNode, old BNode, idx uint16, kids ...BNode) {
	kid_count := uint16(len(kids))
	new.set_header(1, old.bnode_keys_count() + kid_count)
	nodeAppendRange(new, old, 0, 0, idx)
	for i, kid := range kids {
		nodeAppendKV(new, uint16(i)+idx, tree.new(kid), tree.get_node(0), tree.get_node(0))
	}
	nodeAppendRange(new, old, idx+kid_count, idx+kid_count-1, old.bnode_keys_count() - (idx + kid_count) - uint16(1))
}

// Деление ноды на 2, вторая нода всегда вмещается в размер страницы
func SplitNode2(old BNode, left BNode, right BNode){
	
	header_type := old.bnode_type()
	header_key_size := old.bnode_keys_count()
	
	right_node_size := HEADER
	index := uint16(0)
	for ; index < header_key_size; index++ {
		right_node_size += PTR_SIZE
		child_node,_ := old.get_kidPointer(index)
		
		
		offset, _ := old.get_offset(index)
		right.set_offset(index, offset)
		right_node_size += OFFSET_SIZE

		key := old.get_key(index)
		value := old.get_value(index)
		key_size := len(key)
		value_size := len(value)
		right_node_size += key_size + value_size
		if right_node_size > PAGE_SIZE {
			break
		}
		
		nodeAppendKV(right, index, child_node, key, value)
	}
	right.set_header(header_type, index + 1)

	for ; index < header_key_size; index ++ {
		key := old.get_key(index)
		value := old.get_value(index)
		child_node, _ := old.get_kidPointer(index)
		nodeAppendKV(left, index, child_node, key, value)
	}
}

// Сплитим на 3 ноды
func SpliteNode3(old BNode) (uint16, [3]BNode) {
	if old.nbytes() <= PAGE_SIZE {
		old = old[:PAGE_SIZE]
		return 1, [3]BNode{old}
	}
	left := make(BNode, 3*PAGE_SIZE)
	right := make(BNode, PAGE_SIZE)
	SplitNode2(old, left, right)
	if left.nbytes() <= PAGE_SIZE {
		left = left[:PAGE_SIZE]
		return 2, [3]BNode{left, right}
	}
	middle := make(BNode, PAGE_SIZE)
	left_end := make(BNode, PAGE_SIZE)
	SplitNode2(left, left_end, middle)
	return 3, [3]BNode{right, middle, left_end} 

}


func NodeInsert(tree *Btree, new BNode, old BNode, idx uint16, key []byte, val []byte) {
	ptr,_ := old.get_kidPointer(idx)
	node := TreeInsert(tree, tree.get_node(ptr), key, val)
	number_node, split := SpliteNode3(node)
	NkidInsert(tree, new, old, idx, split[:number_node]...)
}


// Метод для добавления нового элемента в B-tree
func TreeInsert(tree *Btree, node BNode, key []byte, val[]byte) BNode {
	new := BNode(make([]byte, 2*PAGE_SIZE))
	
	idx_insert := LookupKeyLE(node, key)
	if node.bnode_type() == LEAF {
		leafInsert(new, node, idx_insert + 1, key, val)
		// Сделать при равенстве значений
	} else {
		NodeInsert(tree, new, node, idx_insert, key, val)
	}
	return new
}



// Высокуровнеый insert
func (tree *Btree) Insert(key []byte, val []byte){
	if tree.root_number_page == 0 {
		root := BNode(make([]byte, PAGE_SIZE))
		root.set_header(LEAF, 2)
		nodeAppendKV(root, 0, 0, nil, nil) // Сторожевое значение, чтобы всегда что то находилось, акутально только в начале
		nodeAppendKV(root, 1, 0, key, val)
		tree.root_number_page = tree.new(root)
		return
	}

	node := TreeInsert(tree, tree.get_node(tree.root_number_page), key, val)
	
	nsplit, split := SpliteNode3(node)
	tree.delete_node(tree.root_number_page)
	if nsplit > 1 {
		root := BNode(make([]byte, PAGE_SIZE))
		root.set_header(INTERNAL_NODE, nsplit)
		for i, node := range(split[:nsplit]) {
			ptr, key := tree.new(node), node.get_key(0)
			nodeAppendKV(root, uint16(i), ptr, key, nil)
		}
		tree.root_number_page = tree.new(root)
		// [1,1000]
		// [1,...,1000],[1000, 10000]

	} else {
		tree.root_number_page = tree.new(split[0])
	}
}


func leafDelete(new BNode, old BNode, idx uint16) {
	nodeAppendRange(new, old, 0, 0, idx - 1)
	nodeAppendRange(new, old, idx, idx+1, old.bnode_keys_count() - idx + 1)
}


// Метод принимает родительскую ноду, индекс на ребёнка и самого ребёнка по этому же индексу
func shouldMerge(tree *Btree, parentNode BNode, current_updatedNode_idx uint16, updated_kidNode BNode) (int, BNode) {
	if updated_kidNode.nbytes() >= PAGE_SIZE / 4 {
		return 0, BNode{}
	}
	kid_ptr, _ := parentNode.get_kidPointer(current_updatedNode_idx-1)
	left := tree.get_node(kid_ptr)
	if current_updatedNode_idx > 0 {
		if left.nbytes() + updated_kidNode.nbytes() - HEADER <= PAGE_SIZE{
			return -1, left
		}
	}
	kid_ptr, _ = parentNode.get_kidPointer(current_updatedNode_idx+1)
	right := tree.get_node(kid_ptr)
	if current_updatedNode_idx + 1 < parentNode.bnode_keys_count() {
		if right.nbytes() + updated_kidNode.nbytes() - HEADER <= PAGE_SIZE {
			return 1, right
		}
	}
	return 0, BNode{}
}


func mergeNode(new BNode, left BNode, right BNode) {
	nodeAppendRange(new, left, 0, 0, left.bnode_keys_count())
	nodeAppendRange(new, right, left.bnode_keys_count() - 1, 0, right.bnode_keys_count())
}


// Метод
func nodeReplace2Kid(new BNode, old BNode, idx uint16, ptr uint64, key []byte) {
	nodeAppendRange(new, old, 0, 0, idx)
	nodeAppendKV(new, idx, ptr, key, nil)
	nodeAppendRange(new, old, idx + 1, idx+1, old.bnode_keys_count() - idx+1)
}


func treeDelete(tree *Btree, node BNode, key BNode) BNode {
	delete_index := LookupKeyLE(node, key)
	
	
	new := BNode(make([]byte, PAGE_SIZE))
	if node.bnode_type() == LEAF {
		kvByte := node.get_key(delete_index)
		if bytes.Compare(kvByte, key) == 0 {
			leafDelete(new, node, delete_index)
			return new
		}
	} else {
		return nodeDelete(tree, node, delete_index, key)
	}
	

}


func nodeDelete(tree *Btree, node BNode, kid_delete_index uint16, key BNode) BNode {
	newParentNode := BNode(make([]byte, PAGE_SIZE))
	kid_ptr,_ := node.get_kidPointer(kid_delete_index)
	kid := tree.get_node(kid_ptr)
	updated_node := treeDelete(tree, kid, key)
	if len(updated_node) == 0 {
		return BNode{}
	}
	tree.delete_node(kid_ptr)
	
	nmerge, sibling := shouldMerge(tree, node, kid_delete_index, updated_node)
	if nmerge < 0 {
		tree.delete_node(uint64(kid_delete_index) - 1)
		merge := BNode(make([]byte, PAGE_SIZE))
		mergeNode(merge, sibling, updated_node)
		nodeReplace2Kid(newParentNode, node, kid_delete_index, tree.new(merge), merge.get_key(0))
		return newParentNode
	}
	if nmerge > 0 {
		tree.delete_node(uint64(kid_delete_index) + 1)
		merge := BNode(make([]byte, PAGE_SIZE))
		mergeNode(merge, sibling, updated_node)
		nodeReplace2Kid(newParentNode, node, kid_delete_index, tree.new(merge), merge.get_key(0))
		return newParentNode
	}
	if nmerge == 0 {
		if updated_node.bnode_keys_count() > 0 {
			NkidInsert(tree, newParentNode, node, kid_delete_index, updated_node)
		} else {
			newParentNode.set_header(INTERNAL_NODE, 0)
		}
	}
	return newParentNode

}


func (tree *Btree) delete(key []byte) bool {
	if tree.root_number_page == 0 {
		return false
	}
	tmp := tree.root_number_page
	treeDelete(tree, tree.get_node(tree.root_number_page), key)
	if tmp != tree.root_number_page {
		return true
	}
	return false
}

