package service

import "duekeep/internal/model"

// requireOwner: чужая строка неотличима от отсутствия (404, не 403).
func requireOwner(ownerID, actorID string) error {
	if ownerID == "" || actorID == "" || ownerID != actorID {
		return model.ErrNotFound
	}
	return nil
}
