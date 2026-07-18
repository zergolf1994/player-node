package services

import (
	"player-node/internal/core/enums"
	"player-node/internal/db/models"
)

// ResolveAdSlug returns the ad feed slug: "hobby" or custom domain slug.
// The feed itself (/advert/{slug}.json) is served by content-node.
func ResolveAdSlug(planType string, domain *models.CustomDomain, spaceID *string) string {
	if planType != "" && planType != enums.PlanTypeHobby {
		if domain != nil && domain.Slug != "" {
			return domain.Slug
		}
		if spaceID != nil && *spaceID != "" {
			if slug := FindDomainSlugBySpaceID(*spaceID); slug != "" {
				return slug
			}
		}
		return ""
	}
	return "hobby"
}
