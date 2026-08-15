package account

import "testing"

func TestPromotedFields(t *testing.T) {
	user := User{
		Person:    Person{Name: "Ayse", Email: "ayse@example.com"},
		LastLogin: "2026-08-15",
	}

	if user.Name != "Ayse" {
		t.Fatalf("promoted Name got %q", user.Name)
	}
}

func TestMethodOverride(t *testing.T) {
	person := Person{Name: "Base", Email: "base@example.com"}
	user := User{Person: person}
	admin := Admin{
		User:       User{Person: Person{Name: "Root", Email: "root@example.com"}},
		Department: "Ops",
	}

	if person.Role() != "person" {
		t.Fatalf("person role got %q", person.Role())
	}
	if user.Role() != "user" {
		t.Fatalf("user role got %q", user.Role())
	}
	if admin.Role() != "admin" {
		t.Fatalf("admin role got %q", admin.Role())
	}
}

func TestDescribeOverride(t *testing.T) {
	admin := Admin{
		User: User{
			Person: Person{Name: "Mehmet", Email: "admin@example.com"},
		},
		Department: "Platform",
	}

	got := admin.Describe()
	if got != "Mehmet <admin@example.com> [admin: Platform]" {
		t.Fatalf("unexpected describe: %q", got)
	}
}

func TestCanAccessComposition(t *testing.T) {
	user := User{Person: Person{Name: "Ayse"}}
	admin := Admin{User: User{Person: Person{Name: "Mehmet"}}, Department: "Ops"}

	if user.CanAccess("dashboard") != true || user.CanAccess("settings") != false {
		t.Fatal("user access rules unexpected")
	}
	if !admin.CanAccess("settings") {
		t.Fatal("admin should access all resources")
	}
}

func TestAccountInterface(t *testing.T) {
	accounts := []Account{
		User{Person: Person{Name: "Ayse", Email: "ayse@example.com"}},
		Admin{
			User:       User{Person: Person{Name: "Mehmet", Email: "admin@example.com"}},
			Department: "Ops",
		},
	}

	if len(accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(accounts))
	}
}
