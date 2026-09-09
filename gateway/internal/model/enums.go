package model

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

func (r Role) Valid() bool {
	switch r {
	case RoleUser, RoleAssistant, RoleSystem:
		return true
	}
	return false
}

type MessageStatus string

const (
	MessageStatusStreaming MessageStatus = "streaming"
	MessageStatusComplete  MessageStatus = "complete"
	MessageStatusFailed    MessageStatus = "failed"
)

func (s MessageStatus) Valid() bool {
	switch s {
	case MessageStatusStreaming, MessageStatusComplete, MessageStatusFailed:
		return true
	}
	return false
}

func (s MessageStatus) Terminal() bool {
	return s == MessageStatusComplete || s == MessageStatusFailed
}

type UserStatus string

const (
	UserStatusActive UserStatus = "active"
	UserStatusBanned UserStatus = "banned"
)

func (s UserStatus) Valid() bool {
	switch s {
	case UserStatusActive, UserStatusBanned:
		return true
	}
	return false
}
