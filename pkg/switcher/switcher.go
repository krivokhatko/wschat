package switcher

// switch tracks the set of active users and spread messages to them.
type Switcher struct {
	// Logged-in users.
	users map[*User]bool

	// In messages from the users.
	spread chan *ChatMessage

	// Login requests from the users.
	Login chan *User

	// Loggout requests from users.
	logout chan *User
}

func NewSwitcher() *Switcher {
	return &Switcher{
		spread: make(chan *ChatMessage),
		Login:  make(chan *User),
		logout: make(chan *User),
		users:  make(map[*User]bool),
	}
}

func (s *Switcher) Proceed() {
	for {
		select {
		case message := <-s.spread:
			for user := range s.users {
				select {
				case user.Send <- message:
				default:
					close(user.Send)
					delete(s.users, user)
				}
			}
		case user := <-s.Login:
			s.users[user] = true
		case user := <-s.logout:
			if _, ok := s.users[user]; ok {
				delete(s.users, user)
				close(user.Send)
			}
		}
	}
}
