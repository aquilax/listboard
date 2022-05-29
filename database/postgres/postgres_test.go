package postgres

import (
	"reflect"
	"testing"

	"github.com/aquilax/listboard/database"
)

func TestImplementsDatabase(t *testing.T) {
	inter := reflect.TypeOf((*database.Database)(nil)).Elem()

	if !reflect.TypeOf(New()).Implements(inter) {
		t.Errorf("Postgres does not implement the database interface")
	}
}