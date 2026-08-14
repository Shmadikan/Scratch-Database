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
	
	assert.Equal(t, structure.get("10"), "huina", "Not correct")
	structure.insert("11", big_value)
	structure.insert("12", big_value)
	structure.insert("13", big_value)
	assert.Equal(t, structure.delete("11"), true, "Not correct")
	assert.Equal(t, structure.delete("13"), true, "Not correct")
	assert.Equal(t, structure.delete("6"), true, "Not correct")
	assert.Equal(t, structure.delete("7"), true, "Not correct")
	assert.Equal(t, structure.delete("5"), true, "Not correct")
	assert.Equal(t, structure.delete("4"), true, "Not correct")
	assert.Equal(t, structure.delete("3"), true, "Not correct")
	assert.Equal(t, structure.delete("2"), true, "Not correct")
	assert.Equal(t, structure.delete("1"), true, "Not correct")
	assert.Equal(t, structure.delete("12"), true, "Not correct")
	assert.Equal(t, structure.delete("9"), true, "Not correct")
	assert.Equal(t, structure.delete("a"), false, "Not correct")




	assert.Equal(t, structure.get("10"), "huina", "Not correct")
	assert.Equal(t, structure.delete("10"), true, "Not correct")
	assert.Equal(t, structure.root.bnode_keys_count(), uint16(1), "Not correct")
}

// Тесты операций get, insert и delete.

func TestGet(t *testing.T) {
	c := newC()
	c.insert("apple", "1")
	c.insert("banana", "2")
	c.insert("cherry", "3")

	assert.Equal(t, "1", c.get("apple"))
	assert.Equal(t, "2", c.get("banana"))
	assert.Equal(t, "3", c.get("cherry"))
	assert.Equal(t, "", c.get("pear"))
	assert.Equal(t, "", c.get("0"))
}

// Вставка в неотсортированном порядке
func TestInsertUnsortedOrder(t *testing.T) {
	c := newC()
	items := []struct {
		key string
		val string
	}{
		{"5", "five"},
		{"3", "three"},
		{"8", "eight"},
		{"1", "one"},
		{"9", "nine"},
		{"2", "two"},
		{"7", "seven"},
		{"4", "four"},
		{"6", "six"},
	}

	for _, it := range items {
		c.insert(it.key, it.val)
	}
	for _, it := range items {
		assert.Equal(t, it.val, c.get(it.key))
	}
}

func TestDeleteNonExistent(t *testing.T) {
	c := newC()
	c.insert("1", "one")
	c.insert("2", "two")

	assert.False(t, c.delete("999"), "Not correct")
	assert.Equal(t, "one", c.get("1"))
	assert.Equal(t, "two", c.get("2"))
}

func TestDeleteExistingKey(t *testing.T) {
	c := newC()
	c.insert("1", "one")
	c.insert("2", "two")
	c.insert("3", "three")

	assert.True(t, c.delete("2"))
	assert.Equal(t, "", c.get("2"))
	assert.Equal(t, "one", c.get("1"))
	assert.Equal(t, "three", c.get("3"))
}

// Вставка ключей в неотсортированном порядке, затем чтение их в случайном порядке.
func TestInsertUnsortedRandomGet(t *testing.T) {
	c := newC()
	items := []struct {
		key string
		val string
	}{
		{"5", "five"},
		{"3", "three"},
		{"8", "eight"},
		{"1", "one"},
		{"9", "nine"},
		{"2", "two"},
		{"7", "seven"},
		{"4", "four"},
		{"6", "six"},
		{"0", "zero"},
	}

	for _, it := range items {
		c.insert(it.key, it.val)
	}

	r := rand.New(rand.NewSource(42))
	for _, idx := range r.Perm(len(items)) {
		it := items[idx]
		assert.Equal(t, it.val, c.get(it.key))
	}
}
