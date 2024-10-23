package pg

import "testing"

func TestDB(t *testing.T) {
	// TODO: hardcoding this is bad.
	db, err := New("postgres://tompston:postgres@localhost:5432/postgres")
	if err != nil {
		t.Fatalf("DB.New() error = %v, want nil", err)
	}
	defer db.Close()
	if db.Conn() == nil {
		t.Errorf("DB.Conn() = %v, want not nil", db.Conn())
	}
}
