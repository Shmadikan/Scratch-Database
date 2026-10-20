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

	store := NewInMemoryStore()
	store.Insert("1", "2")

	assert.Equal(t, store.Get("1"), "2", "Not correct key.")
	assert.Equal(t, store.Get("6"), "", "Error")

	store.Insert("4", "9439429423")
	store.Insert("6", "123143232")
	flowedValue := strings.Repeat("0", 2500)

	store.Insert("8", flowedValue)

	assert.Equal(t, store.Get("4"), "9439429423", "Not correct key.")
	assert.Equal(t, store.Get("6"), "123143232", "Not correct key.")
	assert.Equal(t, store.Get("8"), flowedValue, "Not correct key.")
	store.Insert("9", flowedValue)
	assert.Equal(t, store.Get("9"), flowedValue, "Not correct")

}

// Тест с удалением и вставкой
func TestDelete(t *testing.T) {
	store := NewInMemoryStore()
	bigValue := strings.Repeat("0", 2500)
	store.Insert("1", "2")
	store.Insert("2", "4")
	store.Insert("3", "5")
	store.Insert("4", "6")
	store.Insert("5", "7")
	store.Insert("6", bigValue)
	assert.Equal(t, store.Get("6"), bigValue, "Not correct")

	store.Insert("7", bigValue)

	assert.Equal(t, store.Get("2"), "4", "Not correct")
	store.Insert("9", "test")
	store.Insert("10", "huina")

	assert.Equal(t, store.Get("10"), "huina", "Not correct")
	store.Insert("11", bigValue)
	store.Insert("12", bigValue)
	store.Insert("13", bigValue)

	assert.Equal(t, store.Delete("11"), true, "Not correct")
	assert.Equal(t, store.Delete("13"), true, "Not correct")
	assert.Equal(t, store.Delete("6"), true, "Not correct")
	assert.Equal(t, store.Delete("7"), true, "Not correct")
	assert.Equal(t, store.Delete("5"), true, "Not correct")
	assert.Equal(t, store.Delete("4"), true, "Not correct")
	assert.Equal(t, store.Delete("3"), true, "Not correct")
	assert.Equal(t, store.Delete("2"), true, "Not correct")
	assert.Equal(t, store.Delete("1"), true, "Not correct")
	assert.Equal(t, store.Delete("12"), true, "Not correct")
	assert.Equal(t, store.Delete("9"), true, "Not correct")
	assert.Equal(t, store.Delete("a"), false, "Not correct")

	assert.Equal(t, store.Get("10"), "huina", "Not correct")
	assert.Equal(t, store.Delete("10"), true, "Not correct")
	assert.Equal(t, store.root.nodeType(), leafNode, "Not correct")

	store.Insert("1", "2")
	store.Insert("2", "4")
	store.Insert("3", "5")
	store.Insert("4", "6")
	store.Insert("5", "7")
	store.Insert("6", bigValue)
	assert.Equal(t, store.Get("6"), bigValue, "Not correct")

	store.Insert("7", bigValue)

	assert.Equal(t, store.Get("2"), "4", "Not correct")
	store.Insert("9", "test")
	store.Insert("10", "huina")

	assert.Equal(t, store.Get("10"), "huina", "Not correct")
	store.Insert("11", bigValue)
	store.Insert("12", bigValue)
	store.Insert("13", bigValue)

	assert.Equal(t, store.Delete("11"), true, "Not correct")
	assert.Equal(t, store.Delete("13"), true, "Not correct")
	assert.Equal(t, store.Delete("6"), true, "Not correct")
	assert.Equal(t, store.Delete("7"), true, "Not correct")
	assert.Equal(t, store.Delete("5"), true, "Not correct")
	assert.Equal(t, store.Delete("4"), true, "Not correct")
	assert.Equal(t, store.Delete("3"), true, "Not correct")
	assert.Equal(t, store.Delete("2"), true, "Not correct")
	assert.Equal(t, store.Delete("1"), true, "Not correct")
	assert.Equal(t, store.Delete("12"), true, "Not correct")
	assert.Equal(t, store.Delete("9"), true, "Not correct")
	assert.Equal(t, store.Delete("a"), false, "Not correct")

	assert.Equal(t, store.Get("10"), "huina", "Not correct")
	assert.Equal(t, store.Delete("10"), true, "Not correct")
	assert.Equal(t, store.root.nodeType(), leafNode, "Not correct")
}

// Тесты операций get, insert и delete.

func TestGet(t *testing.T) {
	store := NewInMemoryStore()
	store.Insert("apple", "1")
	store.Insert("banana", "2")
	store.Insert("cherry", "3")

	assert.Equal(t, "1", store.Get("apple"))
	assert.Equal(t, "2", store.Get("banana"))
	assert.Equal(t, "3", store.Get("cherry"))
	assert.Equal(t, "", store.Get("pear"))
	assert.Equal(t, "", store.Get("0"))
}

// Вставка в неотсортированном порядке
func TestInsertUnsortedOrder(t *testing.T) {
	store := NewInMemoryStore()
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
		store.Insert(it.key, it.val)
	}
	for _, it := range items {
		assert.Equal(t, it.val, store.Get(it.key))
	}
}

func TestDeleteNonExistent(t *testing.T) {
	store := NewInMemoryStore()
	store.Insert("1", "one")
	store.Insert("2", "two")

	assert.False(t, store.Delete("999"), "Not correct")
	assert.Equal(t, "one", store.Get("1"))
	assert.Equal(t, "two", store.Get("2"))
}

func TestDeleteExistingKey(t *testing.T) {
	store := NewInMemoryStore()
	store.Insert("1", "one")
	store.Insert("2", "two")
	store.Insert("3", "three")

	assert.True(t, store.Delete("2"))
	assert.Equal(t, "", store.Get("2"))
	assert.Equal(t, "one", store.Get("1"))
	assert.Equal(t, "three", store.Get("3"))
}

// Вставка ключей в неотсортированном порядке, затем чтение их в случайном порядке.
func TestInsertUnsortedRandomGet(t *testing.T) {
	store := NewInMemoryStore()
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
		store.Insert(it.key, it.val)
	}

	r := rand.New(rand.NewSource(42))
	for _, idx := range r.Perm(len(items)) {
		it := items[idx]
		assert.Equal(t, it.val, store.Get(it.key))
	}
}

func TestTesting(t *testing.T) {
	f := "Test Harm"
	k := strings.Split(f, " ")
	first := k[0]

	r := first[0]
	fmt.Println(string(r))
}
