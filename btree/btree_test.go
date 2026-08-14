package btree

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Случайный тест
func TestInsert(t *testing.T) {

	var structure *C = newC()
	structure.insert("1", "2")

	assert.Equal(t, structure.get("1"), "2", "Not correct key.")
	assert.Equal(t, structure.get("6"), "", "Error")

	structure.insert("4", "9439429423")
	structure.insert("6", "123143232")
	flowed_value := strings.Repeat("0", 2500)

	structure.insert("8", flowed_value)

	assert.Equal(t, structure.get("4"), "9439429423", "Not correct key.")
	assert.Equal(t, structure.get("6"), "123143232", "Not correct key.")
	assert.Equal(t, structure.get("8"), flowed_value, "Not correct key.")
	structure.insert("9", flowed_value)
	assert.Equal(t, structure.get("9"), flowed_value, "Not correct")

}


// Тест с удалением и вставкой
func TestDelete(t *testing.T) {
	var structure *C = newC()
	big_value := strings.Repeat("0", 2500)
	structure.insert("1", "2")
	structure.insert("2", "4")
	structure.insert("3", "5")
	structure.insert("4", "6")
	structure.insert("5", "7")
	structure.insert("6", big_value)
	assert.Equal(t, structure.get("6"), big_value, "Not correct")
	
	structure.insert("7", big_value)

	root := structure.tree.get_node(structure.tree.root_number_page)
	assert.Equal(t, root.bnode_keys_count(), uint16(2), "Not correct")
	assert.Equal(t, structure.get("2"), "4", "Not correct")
	structure.insert("9", "test")
	structure.insert("10", "huina")
	
	assert.Equal(t, structure.get("a"), "huina", "Not correct")

}
