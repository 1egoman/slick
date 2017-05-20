package gateway

import (
	"time"
)

type TypingUsers struct {
	users map[*User]time.Time
}

func NewTypingUsers() *TypingUsers {
	return &TypingUsers{ users: make(map[*User]time.Time) }
}

// Return a slice of users that are typing at any given point in time.
func (t *TypingUsers) Users() []User {
	var users []User
	for user, timestamp := range t.users {
		if user != nil && timestamp.Add(20 * time.Second).Unix() > time.Now().Unix() { // If a typing indicator is < 20s old...
			users = append(users, *user) // THe user is "still" typing.
		} else {
			delete(t.users, user) // Remove the user if it's too old.
		}
	}
	return users
}

// Given a user and a timestamp, store their last typing timestamp.
func (t *TypingUsers) Add(user User, time time.Time) {
	t.users[&user] = time
}
// Given a user, if they're typing, stop their typing.
func (t *TypingUsers) Remove(user User) {
	for u, _ := range t.users {
		if u.Name == user.Name {
			delete(t.users, u)
		}
	}
}
