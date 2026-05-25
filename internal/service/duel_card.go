package service

import (
	"context"

	"github.com/GalahadKingsman/clutch/internal/models"
	"github.com/GalahadKingsman/clutch/internal/repository"
	"github.com/google/uuid"
)

type DuelCardService struct {
	users *repository.UserRepository
}

func NewDuelCardService(users *repository.UserRepository) *DuelCardService {
	return &DuelCardService{users: users}
}

func (s *DuelCardService) Enrich(ctx context.Context, d models.Duel) (models.DuelCard, error) {
	card := models.DuelCard{Duel: d}
	creator, err := s.users.GetByID(ctx, d.CreatorID)
	if err != nil || creator == nil {
		return card, err
	}
	card.Creator = participantFromUser(*creator)
	if d.OpponentID != nil {
		opp, err := s.users.GetByID(ctx, *d.OpponentID)
		if err != nil {
			return card, err
		}
		if opp != nil {
			p := participantFromUser(*opp)
			card.Opponent = &p
		}
	}
	return card, nil
}

func (s *DuelCardService) EnrichMany(ctx context.Context, duels []models.Duel) ([]models.DuelCard, error) {
	out := make([]models.DuelCard, 0, len(duels))
	for _, d := range duels {
		c, err := s.Enrich(ctx, d)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func participantFromUser(u models.User) models.DuelParticipant {
	return models.DuelParticipant{
		ID:               u.ID,
		FirstName:        u.FirstName,
		TelegramUsername: u.TelegramUsername,
		PhotoURL:         u.PhotoURL,
	}
}

func UserInDuel(d models.Duel, userID uuid.UUID) bool {
	if d.CreatorID == userID {
		return true
	}
	return d.OpponentID != nil && *d.OpponentID == userID
}

func OtherUserID(d models.Duel, userID uuid.UUID) *uuid.UUID {
	if d.CreatorID == userID && d.OpponentID != nil {
		return d.OpponentID
	}
	if d.OpponentID != nil && *d.OpponentID == userID {
		id := d.CreatorID
		return &id
	}
	return nil
}
