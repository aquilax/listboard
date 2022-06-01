package cached

import (
	"reflect"
	"testing"

	"github.com/aquilax/listboard/database"
	"github.com/aquilax/listboard/database/memory"
)

func TestImplementsDatabase(t *testing.T) {
	inter := reflect.TypeOf((*database.Database)(nil)).Elem()

	if !reflect.TypeOf(New(memory.New())).Implements(inter) {
		t.Errorf("Cached does not implement the database interface")
	}
}
