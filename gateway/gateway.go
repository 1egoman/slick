package gateway

// A Connection is used to represent a message source.
type Connection interface {
	// Each connection has a name.
	Name() string

	Connect() error

	// Called to "refetch" any persistent resources, such as channels.
	Refresh() error

	// Get incoming and outgoing message buffers
	Incoming() chan Event
	Outgoing() chan Event

	MessageHistory() []Message
	AppendMessageHistory(message Message)
	ClearMessageHistory()
	SendMessage(Message, *Channel) (*Message, error)
	ParseMessage(map[string]interface{}, map[string]*User) (*Message, error)

	// Fetch a slice of all channels that are available on this connection
	Channels() []Channel
	FetchChannels() ([]Channel, error)
	SelectedChannel() *Channel
	SetSelectedChannel(*Channel)

	// Fetch the team associated with this connection.
	Team() *Team

	// Fetch user that is authenticated
	Self() *User

	// Given a channel, fetch the message history for that channel
	FetchChannelMessages(Channel) ([]Message, error)

	UserById(string) (*User, error)

	// Post a large block of text in a given channel
	PostText(title string, body string) error
}

// Events are emitted when data comes in from a connection
// and sent when data is to be sent to a connection.
// ie, when another user sends a message, an event would come in:
// Event{
//     Direction: "incoming",
//     Type: "message",
//     Data: map[string]interface{
//       "text": "Hello World!",
//       ...
//     },
// }
type Event struct {
	Direction string // "incoming" or "outgoing"

	Type string `json:"type"`
	Data map[string]interface{}

	// Properties that an event may be associated with.
	Channel Channel
	User    User
}

// A user is a human or bot that sends messages on a channel.
type User struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Avatar   string `json:"avatar"`
	Status   string `json:"status"`
	RealName string `json:"real_name"`
	Email    string `json:"email"`
	Skype    string `json:"skype"`
	Phone    string `json:"phone"`
}

// A Team is a collection of channels.
type Team struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

// A Channel is a independant stream of messages sent by users.
type Channel struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Creator    *User  `json:"creator"`
	Created    int    `json:"created"`
	IsMember   bool   `json:"is_member"`
	IsArchived bool   `json:"is_archived"`
}

// A Reaction is an optional subcollection of a message.
type Reaction struct {
	Name  string  `json:"name"`
	Users []*User `json:"users"`
}

// A file is an optional key on a message.
type File struct {
	Name       string `json:"name"`
	Filetype   string `json:"type"`
	User       *User  `json:"user"`
	PrivateUrl string `json:"url_private"`
	Permalink  string `json:"permalink"`
}

// A Message is a blob of text or media sent by a User within a Channel.
type Message struct {
	Sender    *User      `json:"sender"`
	Text      string     `json:"text"`
	Reactions []Reaction `json:"reactions"`
	Hash      string     `json:"hash"`
	Timestamp int        `json:"timestamp"` // This value is in seconds!
	File      *File      `json:"file,omitempty"`
}
