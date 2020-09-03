/*
Package repository holds event sourced repositories
*/
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/vardius/go-api-boilerplate/cmd/user/internal/domain/user"
	"github.com/vardius/go-api-boilerplate/pkg/application"
	"github.com/vardius/go-api-boilerplate/pkg/errors"
	"github.com/vardius/go-api-boilerplate/pkg/eventbus"
	"github.com/vardius/go-api-boilerplate/pkg/eventstore"
)

type userRepository struct {
	eventStore eventstore.EventStore
	eventBus   eventbus.EventBus
}

// NewUserRepository creates new user event sourced repository
func NewUserRepository(store eventstore.EventStore, bus eventbus.EventBus) user.Repository {
	return &userRepository{store, bus}
}

// Save current user changes to event store and publish each event with an event bus
func (r *userRepository) Save(ctx context.Context, u user.User) error {
	if err := r.eventStore.Store(ctx, u.Changes()); err != nil {
		return errors.Wrap(err)
	}

	for _, event := range u.Changes() {
		if err := r.eventBus.Publish(ctx, event); err != nil {
			return errors.Wrap(err)
		}
	}

	return nil
}

// Save current user changes to event store and publish each event with an event bus
// blocks until event handlers are finished
func (r *userRepository) SaveAndAcknowledge(ctx context.Context, u user.User) error {
	if err := r.eventStore.Store(ctx, u.Changes()); err != nil {
		return errors.Wrap(err)
	}

	for _, event := range u.Changes() {
		if err := r.eventBus.PublishAndAcknowledge(ctx, event); err != nil {
			return errors.Wrap(err)
		}
	}

	return nil
}

// Get user with current state applied
func (r *userRepository) Get(ctx context.Context, id uuid.UUID) (user.User, error) {
	events, err := r.eventStore.GetStream(ctx, id, user.StreamName)
	if err != nil {
		return user.User{}, errors.Wrap(err)
	}

	if len(events) == 0 {
		return user.User{}, application.ErrNotFound
	}

	return user.FromHistory(events), nil
}
