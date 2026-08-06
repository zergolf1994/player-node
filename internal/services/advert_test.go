package services

import (
	"testing"

	"player-node/internal/db/models"
)

func TestActiveAdvertsFiltersDisabledEntries(t *testing.T) {
	source := &models.DomainAdverts{
		Video: models.DomainAdvertCategory{
			Enabled: true,
			List: []models.AdContent{
				{ID: "active", Enabled: true},
				{ID: "disabled", Enabled: false},
			},
		},
		Image: models.DomainAdvertCategory{
			Enabled: false,
			List:    []models.AdContent{{ID: "hidden", Enabled: true}},
		},
	}

	got := activeAdverts(source)
	if got == nil {
		t.Fatal("activeAdverts() returned nil")
	}
	if !got.Video.Enabled || len(got.Video.List) != 1 || got.Video.List[0].ID != "active" {
		t.Fatalf("unexpected active video adverts: %#v", got.Video)
	}
	if got.Image.Enabled || len(got.Image.List) != 0 {
		t.Fatalf("disabled image category was not removed: %#v", got.Image)
	}
}

func TestActiveAdvertsReturnsNilWhenNothingIsActive(t *testing.T) {
	source := &models.DomainAdverts{
		Video: models.DomainAdvertCategory{
			Enabled: true,
			List:    []models.AdContent{{ID: "disabled", Enabled: false}},
		},
	}

	if got := activeAdverts(source); got != nil {
		t.Fatalf("activeAdverts() = %#v, want nil", got)
	}
}
