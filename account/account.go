package account

import "fmt"

// Account is a narrow interface for any domain account type.
type Account interface {
	Role() string
	Describe() string
	CanAccess(resource string) bool
}

// Person is the shared base data for all accounts.
type Person struct {
	Name  string
	Email string
}

func (p Person) Role() string {
	return "person"
}

func (p Person) Describe() string {
	return fmt.Sprintf("%s <%s>", p.Name, p.Email)
}

func (p Person) CanAccess(resource string) bool {
	return false
}

// User embeds Person to promote fields and methods, then adds user-specific state.
type User struct {
	Person
	LastLogin string
}

func (u User) Role() string {
	return "user"
}

func (u User) CanAccess(resource string) bool {
	return resource == "dashboard"
}

// Admin embeds User, inheriting promoted members through composition.
type Admin struct {
	User
	Department string
}

func (a Admin) Role() string {
	return "admin"
}

func (a Admin) Describe() string {
	return fmt.Sprintf("%s [admin: %s]", a.User.Describe(), a.Department)
}

func (a Admin) CanAccess(resource string) bool {
	return true
}

// ComposeSummary explains embedding and override behavior.
func ComposeSummary() string {
	return `Composition patterns in this package:

- User embeds Person  -> Name and Email are promoted to User
- Admin embeds User   -> Person fields/methods promote through User
- Outer types override embedded methods like Role() and Describe()
- No inheritance keyword; behavior is reused by struct embedding`
}
