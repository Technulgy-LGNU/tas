package web

import (
	"regexp"
	"strings"
	"tas/internal/database"

	"github.com/gofiber/fiber/v2"
)

var slugCleanPattern = regexp.MustCompile(`[^a-z0-9]+`)

type teamPayload struct {
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Summary   string `json:"summary"`
	ImageID   *uint  `json:"image_id"`
	Published bool   `json:"published"`
	SortOrder int    `json:"sort_order"`
}

func (a *API) listTeams(c *fiber.Ctx) error {
	var teams []database.Team
	if err := a.DB.Preload("Image").Preload("Prizes.Competition").Order("sort_order asc, name asc").Find(&teams).Error; err != nil {
		return err
	}
	return c.JSON(teams)
}

func (a *API) createTeam(c *fiber.Ctx) error {
	payload, err := bindJSON[teamPayload](c)
	if err != nil {
		return err
	}
	team, err := teamFromPayload(payload)
	if err != nil {
		return err
	}
	if err := a.DB.Create(&team).Error; err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(team)
}

func (a *API) updateTeam(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	payload, err := bindJSON[teamPayload](c)
	if err != nil {
		return err
	}
	team, err := teamFromPayload(payload)
	if err != nil {
		return err
	}
	if err := a.DB.Model(&database.Team{}).Where("id = ?", id).Updates(map[string]any{
		"name":       team.Name,
		"slug":       team.Slug,
		"summary":    team.Summary,
		"image_id":   team.ImageID,
		"published":  team.Published,
		"sort_order": team.SortOrder,
	}).Error; err != nil {
		return err
	}
	if err := a.DB.Preload("Image").Preload("Prizes.Competition").First(&team, id).Error; err != nil {
		return notFoundOrError(err, "team not found")
	}
	return c.JSON(team)
}

func (a *API) deleteTeam(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	if err := a.DB.Delete(&database.Team{}, id).Error; err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type competitionPayload struct {
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Location  string  `json:"location"`
	StartsOn  *string `json:"starts_on"`
	EndsOn    *string `json:"ends_on"`
	Published bool    `json:"published"`
	SortOrder int     `json:"sort_order"`
}

func (a *API) listCompetitions(c *fiber.Ctx) error {
	var competitions []database.Competition
	if err := a.DB.Preload("Prizes.Team").Order("sort_order asc, starts_on desc, name asc").Find(&competitions).Error; err != nil {
		return err
	}
	return c.JSON(competitions)
}

func (a *API) createCompetition(c *fiber.Ctx) error {
	payload, err := bindJSON[competitionPayload](c)
	if err != nil {
		return err
	}
	competition, err := competitionFromPayload(payload)
	if err != nil {
		return err
	}
	if err := a.DB.Create(&competition).Error; err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(competition)
}

func (a *API) updateCompetition(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	payload, err := bindJSON[competitionPayload](c)
	if err != nil {
		return err
	}
	competition, err := competitionFromPayload(payload)
	if err != nil {
		return err
	}
	if err := a.DB.Model(&database.Competition{}).Where("id = ?", id).Updates(map[string]any{
		"name":       competition.Name,
		"slug":       competition.Slug,
		"location":   competition.Location,
		"starts_on":  competition.StartsOn,
		"ends_on":    competition.EndsOn,
		"published":  competition.Published,
		"sort_order": competition.SortOrder,
	}).Error; err != nil {
		return err
	}
	if err := a.DB.Preload("Prizes.Team").First(&competition, id).Error; err != nil {
		return notFoundOrError(err, "competition not found")
	}
	return c.JSON(competition)
}

func (a *API) deleteCompetition(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	if err := a.DB.Delete(&database.Competition{}, id).Error; err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type prizePayload struct {
	Title         string `json:"title"`
	Placement     string `json:"placement"`
	Notes         string `json:"notes"`
	TeamID        uint   `json:"team_id"`
	CompetitionID uint   `json:"competition_id"`
	SortOrder     int    `json:"sort_order"`
}

func (a *API) listPrizes(c *fiber.Ctx) error {
	var prizes []database.Prize
	if err := a.DB.Preload("Team").Preload("Competition").Order("sort_order asc, created_at desc").Find(&prizes).Error; err != nil {
		return err
	}
	return c.JSON(prizes)
}

func (a *API) createPrize(c *fiber.Ctx) error {
	payload, err := bindJSON[prizePayload](c)
	if err != nil {
		return err
	}
	prize, err := prizeFromPayload(payload)
	if err != nil {
		return err
	}
	if err := a.DB.Create(&prize).Error; err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(prize)
}

func (a *API) updatePrize(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	payload, err := bindJSON[prizePayload](c)
	if err != nil {
		return err
	}
	prize, err := prizeFromPayload(payload)
	if err != nil {
		return err
	}
	if err := a.DB.Model(&database.Prize{}).Where("id = ?", id).Updates(map[string]any{
		"title":          prize.Title,
		"placement":      prize.Placement,
		"notes":          prize.Notes,
		"team_id":        prize.TeamID,
		"competition_id": prize.CompetitionID,
		"sort_order":     prize.SortOrder,
	}).Error; err != nil {
		return err
	}
	if err := a.DB.Preload("Team").Preload("Competition").First(&prize, id).Error; err != nil {
		return notFoundOrError(err, "prize not found")
	}
	return c.JSON(prize)
}

func (a *API) deletePrize(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	if err := a.DB.Delete(&database.Prize{}, id).Error; err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type sponsorCategoryPayload struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

func (a *API) listSponsorCategories(c *fiber.Ctx) error {
	var categories []database.SponsorCategory
	if err := a.DB.Preload("Sponsors.LogoImage").Order("sort_order asc, name asc").Find(&categories).Error; err != nil {
		return err
	}
	return c.JSON(categories)
}

func (a *API) createSponsorCategory(c *fiber.Ctx) error {
	payload, err := bindJSON[sponsorCategoryPayload](c)
	if err != nil {
		return err
	}
	category, err := sponsorCategoryFromPayload(payload)
	if err != nil {
		return err
	}
	if err := a.DB.Create(&category).Error; err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(category)
}

func (a *API) updateSponsorCategory(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	payload, err := bindJSON[sponsorCategoryPayload](c)
	if err != nil {
		return err
	}
	category, err := sponsorCategoryFromPayload(payload)
	if err != nil {
		return err
	}
	if err := a.DB.Model(&database.SponsorCategory{}).Where("id = ?", id).Updates(map[string]any{
		"name":       category.Name,
		"sort_order": category.SortOrder,
	}).Error; err != nil {
		return err
	}
	if err := a.DB.First(&category, id).Error; err != nil {
		return notFoundOrError(err, "sponsor category not found")
	}
	return c.JSON(category)
}

func (a *API) deleteSponsorCategory(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	if err := a.DB.Delete(&database.SponsorCategory{}, id).Error; err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type sponsorPayload struct {
	Name        string `json:"name"`
	CategoryID  uint   `json:"category_id"`
	LogoImageID *uint  `json:"logo_image_id"`
	WebsiteURL  string `json:"website_url"`
	Published   bool   `json:"published"`
	SortOrder   int    `json:"sort_order"`
}

func (a *API) listSponsors(c *fiber.Ctx) error {
	var sponsors []database.Sponsor
	if err := a.DB.Preload("Category").Preload("LogoImage").Order("sort_order asc, name asc").Find(&sponsors).Error; err != nil {
		return err
	}
	return c.JSON(sponsors)
}

func (a *API) createSponsor(c *fiber.Ctx) error {
	payload, err := bindJSON[sponsorPayload](c)
	if err != nil {
		return err
	}
	sponsor, err := sponsorFromPayload(payload)
	if err != nil {
		return err
	}
	if err := a.DB.Create(&sponsor).Error; err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(sponsor)
}

func (a *API) updateSponsor(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	payload, err := bindJSON[sponsorPayload](c)
	if err != nil {
		return err
	}
	sponsor, err := sponsorFromPayload(payload)
	if err != nil {
		return err
	}
	if err := a.DB.Model(&database.Sponsor{}).Where("id = ?", id).Updates(map[string]any{
		"name":          sponsor.Name,
		"category_id":   sponsor.CategoryID,
		"logo_image_id": sponsor.LogoImageID,
		"website_url":   sponsor.WebsiteURL,
		"published":     sponsor.Published,
		"sort_order":    sponsor.SortOrder,
	}).Error; err != nil {
		return err
	}
	if err := a.DB.Preload("Category").Preload("LogoImage").First(&sponsor, id).Error; err != nil {
		return notFoundOrError(err, "sponsor not found")
	}
	return c.JSON(sponsor)
}

func (a *API) deleteSponsor(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	if err := a.DB.Delete(&database.Sponsor{}, id).Error; err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type homeArticlePayload struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	ImageID   *uint  `json:"image_id"`
	Published bool   `json:"published"`
}

func (a *API) listHomeArticles(c *fiber.Ctx) error {
	var articles []database.HomeArticle
	if err := a.DB.Preload("Image").Order("slot asc").Find(&articles).Error; err != nil {
		return err
	}
	return c.JSON(articles)
}

func (a *API) upsertHomeArticle(c *fiber.Ctx) error {
	slot64, err := uintParam(c, "slot")
	if err != nil {
		return err
	}
	slot := int(slot64)
	if slot < 1 || slot > 3 {
		return fail(fiber.StatusBadRequest, "home article slot must be 1, 2, or 3")
	}
	payload, err := bindJSON[homeArticlePayload](c)
	if err != nil {
		return err
	}
	title := strings.TrimSpace(payload.Title)
	if title == "" {
		return fail(fiber.StatusBadRequest, "home article title is required")
	}
	article := database.HomeArticle{}
	err = a.DB.Where("slot = ?", slot).First(&article).Error
	if err != nil {
		article = database.HomeArticle{Slot: slot}
	}
	article.Title = title
	article.Body = strings.TrimSpace(payload.Body)
	article.ImageID = payload.ImageID
	article.Published = payload.Published
	if article.ID == 0 {
		if err := a.DB.Create(&article).Error; err != nil {
			return err
		}
	} else if err := a.DB.Save(&article).Error; err != nil {
		return err
	}
	if err := a.DB.Preload("Image").First(&article, article.ID).Error; err != nil {
		return err
	}
	return c.JSON(article)
}

func (a *API) publicTeams(c *fiber.Ctx) error {
	var teams []database.Team
	if err := a.DB.Preload("Image").Preload("Prizes.Competition", "published = ?", true).Where("published = ?", true).Order("sort_order asc, name asc").Find(&teams).Error; err != nil {
		return err
	}
	return c.JSON(teams)
}

func (a *API) publicParticipationHistory(c *fiber.Ctx) error {
	var competitions []database.Competition
	if err := a.DB.Preload("Prizes.Team", "published = ?", true).Where("published = ?", true).Order("sort_order asc, starts_on desc, name asc").Find(&competitions).Error; err != nil {
		return err
	}
	return c.JSON(competitions)
}

func (a *API) publicSponsors(c *fiber.Ctx) error {
	var categories []database.SponsorCategory
	if err := a.DB.Preload("Sponsors", "published = ?", true).Preload("Sponsors.LogoImage").Order("sort_order asc, name asc").Find(&categories).Error; err != nil {
		return err
	}
	return c.JSON(categories)
}

func (a *API) publicHome(c *fiber.Ctx) error {
	var articles []database.HomeArticle
	if err := a.DB.Preload("Image").Where("published = ?", true).Order("slot asc").Find(&articles).Error; err != nil {
		return err
	}
	return c.JSON(articles)
}

func teamFromPayload(payload teamPayload) (database.Team, error) {
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		return database.Team{}, fail(fiber.StatusBadRequest, "team name is required")
	}
	slug := makeSlug(firstNonEmpty(payload.Slug, name))
	if slug == "" {
		return database.Team{}, fail(fiber.StatusBadRequest, "team slug is required")
	}
	return database.Team{Name: name, Slug: slug, Summary: strings.TrimSpace(payload.Summary), ImageID: payload.ImageID, Published: payload.Published, SortOrder: payload.SortOrder}, nil
}

func competitionFromPayload(payload competitionPayload) (database.Competition, error) {
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		return database.Competition{}, fail(fiber.StatusBadRequest, "competition name is required")
	}
	slug := makeSlug(firstNonEmpty(payload.Slug, name))
	if slug == "" {
		return database.Competition{}, fail(fiber.StatusBadRequest, "competition slug is required")
	}
	return database.Competition{
		Name:      name,
		Slug:      slug,
		Location:  strings.TrimSpace(payload.Location),
		StartsOn:  payload.StartsOn,
		EndsOn:    payload.EndsOn,
		Published: payload.Published,
		SortOrder: payload.SortOrder,
	}, nil
}

func prizeFromPayload(payload prizePayload) (database.Prize, error) {
	title := strings.TrimSpace(payload.Title)
	if title == "" {
		return database.Prize{}, fail(fiber.StatusBadRequest, "prize title is required")
	}
	if payload.TeamID == 0 || payload.CompetitionID == 0 {
		return database.Prize{}, fail(fiber.StatusBadRequest, "team_id and competition_id are required")
	}
	return database.Prize{Title: title, Placement: strings.TrimSpace(payload.Placement), Notes: strings.TrimSpace(payload.Notes), TeamID: payload.TeamID, CompetitionID: payload.CompetitionID, SortOrder: payload.SortOrder}, nil
}

func sponsorCategoryFromPayload(payload sponsorCategoryPayload) (database.SponsorCategory, error) {
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		return database.SponsorCategory{}, fail(fiber.StatusBadRequest, "sponsor category name is required")
	}
	return database.SponsorCategory{Name: name, SortOrder: payload.SortOrder}, nil
}

func sponsorFromPayload(payload sponsorPayload) (database.Sponsor, error) {
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		return database.Sponsor{}, fail(fiber.StatusBadRequest, "sponsor name is required")
	}
	if payload.CategoryID == 0 {
		return database.Sponsor{}, fail(fiber.StatusBadRequest, "sponsor category_id is required")
	}
	return database.Sponsor{Name: name, CategoryID: payload.CategoryID, LogoImageID: payload.LogoImageID, WebsiteURL: strings.TrimSpace(payload.WebsiteURL), Published: payload.Published, SortOrder: payload.SortOrder}, nil
}

func makeSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = slugCleanPattern.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}
