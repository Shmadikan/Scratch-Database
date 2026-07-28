package btree

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"strings"
)
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

	// Не работает короче нормально сплит, точнее новый рут, содержит неправильные ключи для ссылки на детей, так как там есть сторожевое значение, порядок ключей неправильный и опять же не те ключи ищет
	// root := structure.tree.get_node(structure.tree.root_number_page)
	// first_kid,_ := root.get_kidPointer(0)
	assert.Equal(t, structure.get("9"), flowed_value, "Not coorect")
	assert.Equal(t, structure.to_string(), " 9")
}