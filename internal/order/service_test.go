package order

import (
	"context"
	"errors"
	"testing"
)

func TestServiceGetFallbackUsesLegacyCode(t *testing.T) {
	legacy := 2
	repository := &fakeRepository{order: Order{ID: 1, StatusCode: &legacy}}

	got, err := NewService(repository, ReadModeFallback, WriteModeDual).Get(context.Background(), 1)

	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "shipped" {
		t.Fatalf("status = %q, want shipped", got.Status)
	}
}

func TestServiceGetNewDoesNotFallBackToLegacyCode(t *testing.T) {
	legacy := 2
	repository := &fakeRepository{order: Order{ID: 1, StatusCode: &legacy}}

	got, err := NewService(repository, ReadModeNew, WriteModeDual).Get(context.Background(), 1)

	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "" {
		t.Fatalf("status = %q, want empty", got.Status)
	}
}

func TestServiceUpdateStatusUsesConfiguredWriteMode(t *testing.T) {
	repository := &fakeRepository{}

	err := NewService(repository, ReadModeFallback, WriteModeDual).UpdateStatus(context.Background(), 1, "shipped")

	if err != nil {
		t.Fatal(err)
	}
	if repository.mode != WriteModeDual || repository.status.Code != "shipped" {
		t.Fatalf("got mode %q and status %#v", repository.mode, repository.status)
	}
}

func TestServiceUpdateStatusRejectsUnknownStatus(t *testing.T) {
	repository := &fakeRepository{}

	err := NewService(repository, ReadModeFallback, WriteModeLegacy).UpdateStatus(context.Background(), 1, "unknown")

	if err == nil {
		t.Fatal("expected an error")
	}
	if repository.updated {
		t.Fatal("repository must not be updated")
	}
}

type fakeRepository struct {
	order   Order
	status  Status
	mode    WriteMode
	updated bool
	err     error
}

func (f *fakeRepository) FindByID(context.Context, int64) (Order, error) {
	if f.err != nil {
		return Order{}, f.err
	}
	return f.order, nil
}

func (f *fakeRepository) UpdateStatus(_ context.Context, _ int64, status Status, mode WriteMode) error {
	if f.err != nil {
		return errors.New("repository error")
	}
	f.status = status
	f.mode = mode
	f.updated = true
	return nil
}
