package btree

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Случайный тест
func TestInsert(t *testing.T) {

	var structure *inMemoryStore = newInMemoryStore()
	structure.Insert("1", "2")

	assert.Equal(t, structure.Get("1"), "2", "Not correct key.")
	assert.Equal(t, structure.Get("6"), "", "Error")

	structure.Insert("4", "9439429423")
	structure.Insert("6", "123143232")
	flowedValue := strings.Repeat("0", 2500)

	structure.Insert("8", flowedValue)

	assert.Equal(t, structure.Get("4"), "9439429423", "Not correct key.")
	assert.Equal(t, structure.Get("6"), "123143232", "Not correct key.")
	assert.Equal(t, structure.Get("8"), flowedValue, "Not correct key.")
	structure.Insert("9", flowedValue)
	assert.Equal(t, structure.Get("9"), flowedValue, "Not correct")

}

// Тест с удалением и вставкой
func TestDelete(t *testing.T) {
	var structure *inMemoryStore = newInMemoryStore()
	bigValue := strings.Repeat("0", 2500)
	structure.Insert("1", "2")
	structure.Insert("2", "4")
	structure.Insert("3", "5")
	structure.Insert("4", "6")
	structure.Insert("5", "7")
	structure.Insert("6", bigValue)
	assert.Equal(t, structure.Get("6"), bigValue, "Not correct")

	structure.Insert("7", bigValue)

	assert.Equal(t, structure.Get("2"), "4", "Not correct")
	structure.Insert("9", "test")
	structure.Insert("10", "huina")

	assert.Equal(t, structure.Get("10"), "huina", "Not correct")
	structure.Insert("11", bigValue)
	structure.Insert("12", bigValue)
	structure.Insert("13", bigValue)

	assert.Equal(t, structure.Delete("11"), true, "Not correct")
	assert.Equal(t, structure.Delete("13"), true, "Not correct")
	assert.Equal(t, structure.Delete("6"), true, "Not correct")
	assert.Equal(t, structure.Delete("7"), true, "Not correct")
	assert.Equal(t, structure.Delete("5"), true, "Not correct")
	assert.Equal(t, structure.Delete("4"), true, "Not correct")
	assert.Equal(t, structure.Delete("3"), true, "Not correct")
	assert.Equal(t, structure.Delete("2"), true, "Not correct")
	assert.Equal(t, structure.Delete("1"), true, "Not correct")
	assert.Equal(t, structure.Delete("12"), true, "Not correct")
	assert.Equal(t, structure.Delete("9"), true, "Not correct")
	assert.Equal(t, structure.Delete("a"), false, "Not correct")

	assert.Equal(t, structure.Get("10"), "huina", "Not correct")
	assert.Equal(t, structure.Delete("10"), true, "Not correct")
	assert.Equal(t, structure.root.nodeType(), uint16(leaf), "Not correct")

	structure.Insert("1", "2")
	structure.Insert("2", "4")
	structure.Insert("3", "5")
	structure.Insert("4", "6")
	structure.Insert("5", "7")
	structure.Insert("6", bigValue)
	assert.Equal(t, structure.Get("6"), bigValue, "Not correct")

	structure.Insert("7", bigValue)

	assert.Equal(t, structure.Get("2"), "4", "Not correct")
	structure.Insert("9", "test")
	structure.Insert("10", "huina")

	assert.Equal(t, structure.Get("10"), "huina", "Not correct")
	structure.Insert("11", bigValue)
	structure.Insert("12", bigValue)
	structure.Insert("13", bigValue)

	assert.Equal(t, structure.Delete("11"), true, "Not correct")
	assert.Equal(t, structure.Delete("13"), true, "Not correct")
	assert.Equal(t, structure.Delete("6"), true, "Not correct")
	assert.Equal(t, structure.Delete("7"), true, "Not correct")
	assert.Equal(t, structure.Delete("5"), true, "Not correct")
	assert.Equal(t, structure.Delete("4"), true, "Not correct")
	assert.Equal(t, structure.Delete("3"), true, "Not correct")
	assert.Equal(t, structure.Delete("2"), true, "Not correct")
	assert.Equal(t, structure.Delete("1"), true, "Not correct")
	assert.Equal(t, structure.Delete("12"), true, "Not correct")
	assert.Equal(t, structure.Delete("9"), true, "Not correct")
	assert.Equal(t, structure.Delete("a"), false, "Not correct")

	assert.Equal(t, structure.Get("10"), "huina", "Not correct")
	assert.Equal(t, structure.Delete("10"), true, "Not correct")
	assert.Equal(t, structure.root.nodeType(), uint16(leaf), "Not correct")
}

// Тесты операций get, insert и delete.

func TestGet(t *testing.T) {
	c := newInMemoryStore()
	c.Insert("apple", "1")
	c.Insert("banana", "2")
	c.Insert("cherry", "3")

	assert.Equal(t, "1", c.Get("apple"))
	assert.Equal(t, "2", c.Get("banana"))
	assert.Equal(t, "3", c.Get("cherry"))
	assert.Equal(t, "", c.Get("pear"))
	assert.Equal(t, "", c.Get("0"))
}

// Вставка в неотсортированном порядке
func TestInsertUnsortedOrder(t *testing.T) {
	c := newInMemoryStore()
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
		c.Insert(it.key, it.val)
	}
	for _, it := range items {
		assert.Equal(t, it.val, c.Get(it.key))
	}
}

func TestDeleteNonExistent(t *testing.T) {
	c := newInMemoryStore()
	c.Insert("1", "one")
	c.Insert("2", "two")

	assert.False(t, c.Delete("999"), "Not correct")
	assert.Equal(t, "one", c.Get("1"))
	assert.Equal(t, "two", c.Get("2"))
}

func TestDeleteExistingKey(t *testing.T) {
	c := newInMemoryStore()
	c.Insert("1", "one")
	c.Insert("2", "two")
	c.Insert("3", "three")

	assert.True(t, c.Delete("2"))
	assert.Equal(t, "", c.Get("2"))
	assert.Equal(t, "one", c.Get("1"))
	assert.Equal(t, "three", c.Get("3"))
}

// Вставка ключей в неотсортированном порядке, затем чтение их в случайном порядке.
func TestInsertUnsortedRandomGet(t *testing.T) {
	c := newInMemoryStore()
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
		c.Insert(it.key, it.val)
	}

	r := rand.New(rand.NewSource(42))
	for _, idx := range r.Perm(len(items)) {
		it := items[idx]
		assert.Equal(t, it.val, c.Get(it.key))
	}
}

func TestTesting(t *testing.T) {
	f := "Test Harm"
	k := strings.Split(f, " ")
	first := k[0]

	r := first[0]
	fmt.Println(string(r))
}
