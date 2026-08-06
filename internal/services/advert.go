package services

import (
	"encoding/json"

	"player-node/internal/core/enums"
	"player-node/internal/db/models"
)

// ResolveInlineAdverts returns only enabled adverts that can be embedded directly
// in PLAYER_CONFIG. A nil result omits the adverts field entirely.
func ResolveInlineAdverts(planType string, domain *models.CustomDomain, spaceID *string) *models.DomainAdverts {
	if planType == "" || planType == enums.PlanTypeHobby {
		return activeAdverts(readHobbyAdverts())
	}

	if domain != nil {
		return activeAdverts(domain.Adverts)
	}
	if spaceID != nil && *spaceID != "" {
		if advertDomain := FindDomainBySpaceID(*spaceID); advertDomain != nil {
			return activeAdverts(advertDomain.Adverts)
		}
	}
	return nil
}

func readHobbyAdverts() *models.DomainAdverts {
	settings, err := ReadSettingFile()
	if err != nil {
		return nil
	}
	raw, exists := settings["advert_hobby"]
	if !exists {
		return nil
	}

	var adverts models.DomainAdverts
	if err := json.Unmarshal(raw, &adverts); err != nil {
		return nil
	}
	return &adverts
}

func activeAdverts(source *models.DomainAdverts) *models.DomainAdverts {
	if source == nil {
		return nil
	}

	result := &models.DomainAdverts{
		Video:  activeAdvertCategory(source.Video),
		Image:  activeAdvertCategory(source.Image),
		Script: activeAdvertCategory(source.Script),
	}
	if !result.Video.Enabled && !result.Image.Enabled && !result.Script.Enabled {
		return nil
	}
	return result
}

func activeAdvertCategory(source models.DomainAdvertCategory) models.DomainAdvertCategory {
	if !source.Enabled {
		return models.DomainAdvertCategory{List: []models.AdContent{}}
	}

	list := make([]models.AdContent, 0, len(source.List))
	for _, advert := range source.List {
		if advert.Enabled {
			list = append(list, advert)
		}
	}
	return models.DomainAdvertCategory{
		Enabled: len(list) > 0,
		List:    list,
	}
}
