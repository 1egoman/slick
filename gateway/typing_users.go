package gateway

import (
	"time"
	"sort"
)

// How many seconds after getting a user typing event does the user no longer show as typing?
const TYPING_PERSIST_SECONDS = 5

type TypingUsers struct {
	users map[string]time.Time
}

func NewTypingUsers() *TypingUsers {
	return &TypingUsers{users: make(map[string]time.Time)}
}

// Return a slice of users that are typing at any given point in time.
func (t *TypingUsers) Users() []string {
	var users []string
	for user, timestamp := range t.users {
		if timestamp.Add(TYPING_PERSIST_SECONDS*time.Second).Unix() > time.Now().Unix() { // If a typing indicator is < 5s old...
			users = append(users, user) // THe user is "still" typing.
		} else {
			delete(t.users, user) // Remove the user if it's too old.
		}
	}

	// Sort output by timestamp to ensure that the order is deterministic (really here for tests)
	sort.Slice(users, func (i, j int) bool {
		return t.users[users[i]].Before(t.users[users[j]])
	})

	return users
}

// Given a user and a timestamp, store their last typing timestamp.
func (t *TypingUsers) Add(username string, time time.Time) {
	t.users[username] = time
}

// Given a user, if they're typing, stop their typing.
func (t *TypingUsers) Remove(username string) {
	delete(t.users, username)
}
