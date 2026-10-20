package btree

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type BNode []byte

const headerSize = 4
const ptrSize = 8
const offsetSize = 2
const pageSize = 4096
const leaf = 2
const internalNode = 1

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
func (node BNode) childPointer(idx uint16) (uint64, error) {
	if idx > node.keyCount() {
		return 0, errors.New("not match keys")
	}

	childPtr := headerSize + idx*8
	return binary.LittleEndian.Uint64(node[childPtr : childPtr+9]), nil
}

func (node BNode) setChildPointer(idx uint16, ptr uint64) error {
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
func nodeAppendKV(node BNode, idx uint16, ptr uint64, key []byte, value []byte) {
	node.setChildPointer(idx, ptr)

	keyPosition := node.keyValuePosition(idx)
	binary.LittleEndian.PutUint16(node[keyPosition:], uint16(len(key)))
	binary.LittleEndian.PutUint16(node[keyPosition+2:], uint16(len(value)))

	copy(node[keyPosition+4:], key)
	copy(node[keyPosition+4+uint16(len(key)):], value)

	offset, _ := node.getOffset(idx)
	node.setOffset(idx+1, offset+4+uint16(len(key)+len(value)))
}

// Вставка нескольких элементов со второй ноды (старая) в первую (новая). dstIdx - куда начать вставку в новой ноде, srcIdx - откуда брать элементы в старой. n - число элементов
func nodeAppendRange(dst BNode, old BNode, dstIdx uint16, srcIdx uint16, n uint16) {
	insertIndex := dstIdx
	for i := srcIdx; i < n+srcIdx; i++ {
		key := old.getKey(i)
		val := old.getValue(i)
		ptr, _ := old.childPointer(i)
		nodeAppendKV(dst, insertIndex, ptr, key, val)
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
	nodeAppendRange(right, old, 0, index, headerKeySize-index)
	left.setHeader(headerType, index)
	nodeAppendRange(left, old, 0, 0, index)
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
func leafInsert(newNode BNode, old BNode, idx uint16, key []byte, value []byte) {
	newNode.setHeader(leaf, old.keyCount()+1)
	nodeAppendRange(newNode, old, 0, 0, idx)
	nodeAppendKV(newNode, idx, 0, key, value)
	nodeAppendRange(newNode, old, idx+1, idx, old.keyCount()-idx)
}

// Вставка во внутренний узел
func insertChildren(tree *Btree, newNode BNode, old BNode, idx uint16, children ...BNode) {
	childCount := uint16(len(children))
	newNode.setHeader(internalNode, old.keyCount()+childCount-1)
	nodeAppendRange(newNode, old, 0, 0, idx)
	for i, child := range children {
		nodeAppendKV(newNode, uint16(i)+idx, tree.newNode(child), child.getKey(0), nil)
	}
	nodeAppendRange(newNode, old, idx+childCount, idx+1, old.keyCount()-(idx+1))
}

func nodeInsert(tree *Btree, newNode BNode, old BNode, idx uint16, key []byte, val []byte) {
	ptr, _ := old.childPointer(idx)
	node := treeInsert(tree, tree.getNode(ptr), key, val)
	nodeCount, split := splitNode3(node)
	tree.deleteNode(ptr)
	insertChildren(tree, newNode, old, idx, split[:nodeCount]...)
}

// Метод для добавления нового элемента в B-tree
func treeInsert(tree *Btree, node BNode, key []byte, val []byte) BNode {
	newNode := BNode(make([]byte, 2*pageSize))

	insertIdx := lookupKeyLE(node, key)
	if node.nodeType() == leaf {
		leafInsert(newNode, node, insertIdx+1, key, val)
		// Сделать при равенстве значений
	} else {
		nodeInsert(tree, newNode, node, insertIdx, key, val)
	}
	return newNode
}

// Высокуровнеый insert
func (tree *Btree) Insert(key []byte, val []byte) {
	if tree.rootPageNumber == 0 {
		root := BNode(make([]byte, pageSize))
		root.setHeader(leaf, 2)
		nodeAppendKV(root, 0, 0, nil, nil) // Сторожевое значение, чтобы всегда что то находилось, акутально только в начале
		nodeAppendKV(root, 1, 0, key, val)
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
			ptr, key := tree.newNode(childNode), childNode.getKey(0)
			nodeAppendKV(root, uint16(i), ptr, key, nil)
		}
		tree.rootPageNumber = tree.newNode(root)
		// [1,1000]
		// [1,...,1000],[1000, 10000]
	} else {
		tree.rootPageNumber = tree.newNode(split[0])
	}
}

func leafDelete(newNode BNode, old BNode, idx uint16) {
	newNode.setHeader(leaf, old.keyCount()-1)
	nodeAppendRange(newNode, old, 0, 0, idx)
	nodeAppendRange(newNode, old, idx, idx+1, old.keyCount()-(idx+1))
}

// Метод принимает родительскую ноду, индекс на ребёнка и самого ребёнка по этому же индексу
func shouldMerge(tree *Btree, parentNode BNode, updatedIdx uint16, updatedChild BNode) (int, BNode) {
	if updatedChild.nbytes() >= pageSize/4 {
		return 0, BNode{}
	}

	if updatedIdx > 0 {
		childPtr, _ := parentNode.childPointer(updatedIdx - 1)
		left := tree.getNode(childPtr)
		if left.nbytes()+updatedChild.nbytes()-headerSize <= pageSize {
			return -1, left
		}
	}

	if updatedIdx+1 < parentNode.keyCount() {
		childPtr, _ := parentNode.childPointer(updatedIdx + 1)
		right := tree.getNode(childPtr)
		if right.nbytes()+updatedChild.nbytes()-headerSize <= pageSize {
			return 1, right
		}
	}
	return 0, BNode{}
}

func mergeNode(newNode BNode, left BNode, right BNode) {
	newNode.setHeader(left.nodeType(), left.keyCount()+right.keyCount())
	nodeAppendRange(newNode, left, 0, 0, left.keyCount())
	nodeAppendRange(newNode, right, left.keyCount(), 0, right.keyCount())
}

// Метод
func replaceChild(newNode BNode, old BNode, idx uint16, ptr uint64, key []byte) {
	newNode.setHeader(old.nodeType(), old.keyCount()-1)
	nodeAppendRange(newNode, old, 0, 0, idx)
	nodeAppendKV(newNode, idx, ptr, key, nil)
	nodeAppendRange(newNode, old, idx+1, idx+2, old.keyCount()-(idx+2))
}

func treeDelete(tree *Btree, node BNode, key []byte) BNode {
	deleteIndex := lookupKeyLE(node, key)

	if node.nodeType() == leaf {
		keyBytes := node.getKey(deleteIndex)
		if bytes.Equal(keyBytes, key) {
			newNode := BNode(make([]byte, pageSize))
			leafDelete(newNode, node, deleteIndex)
			return newNode
		}
	} else {
		return nodeDelete(tree, node, deleteIndex, key)
	}
	return BNode{}
}

func nodeDelete(tree *Btree, node BNode, childIndex uint16, key []byte) BNode {
	newParentNode := BNode(make([]byte, pageSize))
	childPtr, _ := node.childPointer(childIndex)
	child := tree.getNode(childPtr)
	updatedNode := treeDelete(tree, child, key)
	if len(updatedNode) == 0 {
		return BNode{}
	}
	tree.deleteNode(childPtr)

	mergeDir, sibling := shouldMerge(tree, node, childIndex, updatedNode)
	if mergeDir < 0 {
		tree.deleteNode(uint64(childIndex) - 1)
		merge := BNode(make([]byte, pageSize))
		mergeNode(merge, sibling, updatedNode)
		replaceChild(newParentNode, node, childIndex-1, tree.newNode(merge), merge.getKey(0))
		return newParentNode
	}
	if mergeDir > 0 {
		tree.deleteNode(uint64(childIndex) + 1)
		merge := BNode(make([]byte, pageSize))
		mergeNode(merge, updatedNode, sibling)
		replaceChild(newParentNode, node, childIndex, tree.newNode(merge), merge.getKey(0))
		return newParentNode
	}
	if mergeDir == 0 {
		if updatedNode.keyCount() > 0 {
			insertChildren(tree, newParentNode, node, childIndex, updatedNode)
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
		child, _ := newRoot.childPointer(0)
		if tree.getNode(child).keyCount() == 1 {
			newRoot.setHeader(leaf, 1)
			nodeAppendKV(newRoot, 0, 0, nil, nil)
			tree.deleteNode(child)
		}
	}
	return true
}

func (tree *Btree) Get(key []byte) (value []byte, found bool) {
	root := tree.rootPageNumber
	node := tree.getNode(root)
	index := lookupKeyLE(node, []byte(key))

	for node.nodeType() != leaf {
		ptr, _ := node.childPointer(index)
		childNode := tree.getNode(ptr)
		node = childNode
		index = lookupKeyLE(node, []byte(key))
	}
	if node.nodeType() == leaf {
		if bytes.Compare([]byte(key), node.getKey(index)) == 0 {
			return node.getValue(index), true
		}
		return BNode{}, false
	}
	return BNode{}, false
}
