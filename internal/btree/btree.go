package btree

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type BNode []byte

const (
	headerSize = 4
	ptrSize    = 8
	offsetSize = 2
	pageSize   = 4096
)

const (
	internalNode uint16 = 1
	leafNode     uint16 = 2
)

var errIndexOutOfRange = errors.New("index out of range")

// Node представлена как последовательность байтов
type Btree struct {
	rootPageNumber uint64
	getNode        func(uint64) BNode
	newNode        func(BNode) uint64
	deleteNode     func(uint64)
}

func (node BNode) nodeType() uint16 {
	return binary.LittleEndian.Uint16(node[0:2])
}

func (node BNode) keyCount() uint16 {
	return binary.LittleEndian.Uint16(node[2:4])
}

func (node BNode) setHeader(nodeType uint16, keyCount uint16) {
	binary.LittleEndian.PutUint16(node[0:2], nodeType)
	binary.LittleEndian.PutUint16(node[2:4], keyCount)
}

// Получить ребёнка ноды
func (node BNode) getChildPointer(idx uint16) (uint64, error) {
	if idx > node.keyCount() {
		return 0, errors.New("not match keys")
	}

	childPtr := headerSize + idx*8
	return binary.LittleEndian.Uint64(node[childPtr : childPtr+9]), nil
}

func (node BNode) setPointer(idx uint16, ptr uint64) error {
	if idx > node.keyCount() {
		return errIndexOutOfRange
	}
	childPtr := headerSize + idx*8
	binary.LittleEndian.PutUint64(node[childPtr:childPtr+9], ptr)
	return nil
}

func offsetPos(node BNode, idx uint16) (uint16, error) {
	if idx < 1 || idx > node.keyCount() {
		return 0, errIndexOutOfRange
	}
	return headerSize + 8*node.keyCount() + 2*(idx-1), nil
}

// декодировать значение n-го смещение
func (node BNode) getOffset(idx uint16) (uint16, error) {
	if idx == 0 {
		return 0, nil
	}
	if idx > node.keyCount() {
		return 0, errIndexOutOfRange
	}

	offset, _ := offsetPos(node, idx)
	return binary.LittleEndian.Uint16(node[offset:]), nil
}

// Установить значение для n-го смещения
func (node BNode) setOffset(idx uint16, offsetValue uint16) error {
	if idx > node.keyCount() {
		return errIndexOutOfRange
	}
	offset, _ := offsetPos(node, idx)
	binary.LittleEndian.PutUint16(node[offset:], offsetValue)
	return nil
}

// Получить абсолютную позицию KV
func (node BNode) keyValuePosition(idx uint16) uint16 {
	offset, err := node.getOffset(idx)
	if err != nil {
		return 0
	}
	return headerSize + 8*node.keyCount() + 2*node.keyCount() + offset

}

// Получить ключ по смещению
func (node BNode) getKey(idx uint16) []byte {
	offset := node.keyValuePosition(idx)
	keyLen := binary.LittleEndian.Uint16(node[offset:])
	return node[offset+4:][:keyLen]
}

func (node BNode) getValue(idx uint16) []byte {
	offset := node.keyValuePosition(idx)
	valueLen := binary.LittleEndian.Uint16(node[(offset + 2):])
	keyLen := binary.LittleEndian.Uint16(node[offset:])
	return node[offset+4+keyLen:][:valueLen]
}

// На позиции последнего оффсета размер ноды в байтах
func (node BNode) nbytes() uint16 {
	return node.keyValuePosition(node.keyCount())
}

// Меньше или равный ключ ищем
func lookupKeyLE(node BNode, key []byte) uint16 {
	nkeys := node.keyCount()

	keyIndex := uint16(0)
	for i := uint16(1); i < nkeys; i++ {
		cmp := bytes.Compare(node.getKey(i), key)
		if cmp <= 0 {
			keyIndex = i
		}
		if cmp >= 0 {
			break
		}

	}
	return keyIndex
}

// Функция Insert в Node новый Key/value, предназначена как для листьев, так и для внутренних узлов
func appendKV(node BNode, idx uint16, ptr uint64, key []byte, value []byte) {
	node.setPointer(idx, ptr)

	keyPosition := node.keyValuePosition(idx)
	binary.LittleEndian.PutUint16(node[keyPosition:], uint16(len(key)))
	binary.LittleEndian.PutUint16(node[keyPosition+2:], uint16(len(value)))

	copy(node[keyPosition+4:], key)
	copy(node[keyPosition+4+uint16(len(key)):], value)

	offset, _ := node.getOffset(idx)
	node.setOffset(idx+1, offset+4+uint16(len(key)+len(value)))
}

// Вставка нескольких элементов со второй ноды (старая) в первую (новую). dstNew - куда начать вставку в новой ноде, srcOld - откуда брать элементы в старой. n - число элементов
func appendRange(dst BNode, src BNode, dstNew uint16, srcOld uint16, n uint16) {
	insertIndex := dstNew
	for i := srcOld; i < n+srcOld; i++ {
		key := src.getKey(i)
		val := src.getValue(i)
		ptr, _ := src.getChildPointer(i)
		appendKV(dst, insertIndex, ptr, key, val)
		insertIndex++
	}
}

// Деление ноды на 2, вторая нода всегда вмещается в размер страницы
func splitNode2(old BNode, left BNode, right BNode) {

	headerType := old.nodeType()
	headerKeySize := old.keyCount()

	right.setHeader(headerType, 0)
	index := uint16(0)
	nodeOversize := uint16(0)
	for ; index < old.keyCount() && old.nbytes()-nodeOversize > pageSize; index++ {
		kvPos := old.keyValuePosition(index)
		keyLen := binary.LittleEndian.Uint16(old[kvPos:])
		valLen := binary.LittleEndian.Uint16(old[kvPos+2:])
		nodeOversize += keyLen + valLen + uint16(4)
	}
	right.setHeader(headerType, headerKeySize-index)
	appendRange(right, old, 0, index, headerKeySize-index)
	left.setHeader(headerType, index)
	appendRange(left, old, 0, 0, index)
}

// Сплитим на 3 ноды
func splitNode3(old BNode) (uint16, [3]BNode) {
	if old.nbytes() <= pageSize {
		old = old[:pageSize]
		return 1, [3]BNode{old}
	}
	left := make(BNode, 3*pageSize)
	right := make(BNode, pageSize)
	splitNode2(old, left, right)
	if left.nbytes() <= pageSize {
		left = left[:pageSize]
		return 2, [3]BNode{left, right}
	}
	middle := make(BNode, pageSize)
	leftEnd := make(BNode, pageSize)
	splitNode2(left, leftEnd, middle)
	return 3, [3]BNode{right, middle, leftEnd}

}

// Copy on Write вставка
func leafInsert(dst BNode, old BNode, idx uint16, key []byte, value []byte) {
	dst.setHeader(leafNode, old.keyCount()+1)
	appendRange(dst, old, 0, 0, idx)
	appendKV(dst, idx, 0, key, value)
	appendRange(dst, old, idx+1, idx, old.keyCount()-idx)
}

// Вставка во внутренний узел
func insertChildren(tree *Btree, dst BNode, old BNode, idx uint16, kids ...BNode) {
	kidCount := uint16(len(kids))
	dst.setHeader(internalNode, old.keyCount()+kidCount-1)
	appendRange(dst, old, 0, 0, idx)
	for i, kid := range kids {
		appendKV(dst, uint16(i)+idx, tree.newNode(kid), kid.getKey(0), nil)
	}
	appendRange(dst, old, idx+kidCount, idx+1, old.keyCount()-(idx+1))
}

func nodeInsert(tree *Btree, dst BNode, old BNode, idx uint16, key []byte, val []byte) {
	ptr, _ := old.getChildPointer(idx)
	node := treeInsert(tree, tree.getNode(ptr), key, val)
	splitCount, split := splitNode3(node)
	tree.deleteNode(ptr)
	insertChildren(tree, dst, old, idx, split[:splitCount]...)
}

// Метод для добавления нового элемента в B-tree
func treeInsert(tree *Btree, node BNode, key []byte, val []byte) BNode {
	dst := BNode(make([]byte, 2*pageSize))

	idxInsert := lookupKeyLE(node, key)
	if node.nodeType() == leafNode {
		leafInsert(dst, node, idxInsert+1, key, val)
		// Сделать при равенстве значений
	} else {
		nodeInsert(tree, dst, node, idxInsert, key, val)
	}
	return dst
}

// Высокуровнеый insert
func (tree *Btree) Insert(key []byte, val []byte) {
	if tree.rootPageNumber == 0 {
		root := BNode(make([]byte, pageSize))
		root.setHeader(leafNode, 2)
		appendKV(root, 0, 0, nil, nil) // Сторожевое значение, чтобы всегда что то находилось, акутально только в начале
		appendKV(root, 1, 0, key, val)
		tree.rootPageNumber = tree.newNode(root)
		return
	}

	node := treeInsert(tree, tree.getNode(tree.rootPageNumber), key, val)

	splitCount, split := splitNode3(node)
	tree.deleteNode(tree.rootPageNumber)
	if splitCount > 1 {
		root := BNode(make([]byte, pageSize))
		root.setHeader(internalNode, splitCount)
		for i, childNode := range split[:splitCount] {
			ptr, nodeKey := tree.newNode(childNode), childNode.getKey(0)
			appendKV(root, uint16(i), ptr, nodeKey, nil)
		}
		tree.rootPageNumber = tree.newNode(root)
		// [1,1000]
		// [1,...,1000],[1000, 10000]

	} else {
		tree.rootPageNumber = tree.newNode(split[0])
	}
}

func leafDelete(dst BNode, old BNode, idx uint16) {
	dst.setHeader(leafNode, old.keyCount()-1)
	appendRange(dst, old, 0, 0, idx)
	appendRange(dst, old, idx, idx+1, old.keyCount()-(idx+1))
}

// Метод принимает родительскую ноду, индекс на ребёнка и самого ребёнка по этому же индексу
func shouldMerge(tree *Btree, parentNode BNode, updatedKidIdx uint16, updatedKid BNode) (int, BNode) {
	if updatedKid.nbytes() >= pageSize/4 {
		return 0, BNode{}
	}

	if updatedKidIdx > 0 {
		kidPtr, _ := parentNode.getChildPointer(updatedKidIdx - 1)
		left := tree.getNode(kidPtr)
		if left.nbytes()+updatedKid.nbytes()-headerSize <= pageSize {
			return -1, left
		}
	}

	if updatedKidIdx+1 < parentNode.keyCount() {
		kidPtr, _ := parentNode.getChildPointer(updatedKidIdx + 1)
		right := tree.getNode(kidPtr)
		if right.nbytes()+updatedKid.nbytes()-headerSize <= pageSize {
			return 1, right
		}
	}
	return 0, BNode{}
}

func mergeNode(dst BNode, left BNode, right BNode) {
	dst.setHeader(left.nodeType(), left.keyCount()+right.keyCount())
	appendRange(dst, left, 0, 0, left.keyCount())
	appendRange(dst, right, left.keyCount(), 0, right.keyCount())
}

// Метод
func replaceChild(dst BNode, old BNode, idx uint16, ptr uint64, key []byte) {
	dst.setHeader(old.nodeType(), old.keyCount()-1)
	appendRange(dst, old, 0, 0, idx)
	appendKV(dst, idx, ptr, key, nil)
	appendRange(dst, old, idx+1, idx+2, old.keyCount()-(idx+2))
}

func treeDelete(tree *Btree, node BNode, key []byte) BNode {
	deleteIndex := lookupKeyLE(node, key)

	if node.nodeType() == leafNode {
		kvByte := node.getKey(deleteIndex)
		if bytes.Equal(kvByte, key) {
			dst := BNode(make([]byte, pageSize))
			leafDelete(dst, node, deleteIndex)
			return dst
		}
	} else {
		return nodeDelete(tree, node, deleteIndex, key)
	}
	return BNode{}

}

func nodeDelete(tree *Btree, node BNode, kidDeleteIndex uint16, key []byte) BNode {
	newParentNode := BNode(make([]byte, pageSize))
	kidPtr, _ := node.getChildPointer(kidDeleteIndex)
	kid := tree.getNode(kidPtr)
	updatedNode := treeDelete(tree, kid, key)
	if len(updatedNode) == 0 {
		return BNode{}
	}
	tree.deleteNode(kidPtr)

	nmerge, sibling := shouldMerge(tree, node, kidDeleteIndex, updatedNode)
	if nmerge < 0 {
		tree.deleteNode(uint64(kidDeleteIndex) - 1)
		merge := BNode(make([]byte, pageSize))
		mergeNode(merge, sibling, updatedNode)
		replaceChild(newParentNode, node, kidDeleteIndex-1, tree.newNode(merge), merge.getKey(0))
		return newParentNode
	}
	if nmerge > 0 {
		tree.deleteNode(uint64(kidDeleteIndex) + 1)
		merge := BNode(make([]byte, pageSize))
		mergeNode(merge, updatedNode, sibling)
		replaceChild(newParentNode, node, kidDeleteIndex, tree.newNode(merge), merge.getKey(0))
		return newParentNode
	}
	if nmerge == 0 {
		if updatedNode.keyCount() > 0 {
			insertChildren(tree, newParentNode, node, kidDeleteIndex, updatedNode)
		} else {
			newParentNode.setHeader(internalNode, 0)
		}
	}
	return newParentNode

}

func (tree *Btree) Delete(key []byte) bool {
	if tree.rootPageNumber == 0 {
		return false
	}
	newRoot := treeDelete(tree, tree.getNode(tree.rootPageNumber), key)
	if len(newRoot) == 0 {
		return false
	}

	tree.deleteNode(tree.rootPageNumber)
	page := tree.newNode(newRoot)
	tree.rootPageNumber = page

	if newRoot.keyCount() == 1 {
		kid, _ := newRoot.getChildPointer(0)
		if tree.getNode(kid).keyCount() == 1 {
			newRoot.setHeader(leafNode, 1)
			appendKV(newRoot, 0, 0, nil, nil)
			tree.deleteNode(kid)
		}
	}
	return true
}

func (tree *Btree) Get(key []byte) (value []byte, found bool) {
	root := tree.rootPageNumber
	node := tree.getNode(root)
	index := lookupKeyLE(node, []byte(key))

	for node.nodeType() != leafNode {
		ptr, _ := node.getChildPointer(index)
		kidNode := tree.getNode(ptr)
		node = kidNode
		index = lookupKeyLE(node, []byte(key))
	}
	if node.nodeType() == leafNode {
		if bytes.Compare([]byte(key), node.getKey(index)) == 0 {
			return node.getValue(index), true
		}
		return BNode{}, false
	}
	return BNode{}, false
}
