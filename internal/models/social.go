package models

import (
	"time"

	"github.com/google/uuid"
)

type FriendCard struct {
	User         UserResponse `json:"user"`
	ContactAlias *string      `json:"contact_alias,omitempty"`
	Record       *string      `json:"record,omitempty"`
	Since        time.Time    `json:"since"`
}

type FeedResponse struct {
	IncomingChallenges []DuelCard       `json:"incoming_challenges"`
	ActiveDuels        []DuelCard       `json:"active_duels"`
	Activity           []FeedActivity   `json:"activity"`
}

type FeedActivity struct {
	ID        uuid.UUID `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
