package web

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/url"
	"regexp"
	"strings"
	"tas/backend/database"
	"time"
	"unicode/utf8"
)

var websiteKinds = map[string]bool{"home": true, "teams": true, "events": true, "sponsors": true, "publications": true, "blog": true}
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var youtubeID = regexp.MustCompile(`^[a-zA-Z0-9_-]{11}$`)

func textValid(t database.LocalizedText, max int, required bool) bool {
	for _, value := range []string{t.DE, t.EN} {
		if !utf8.ValidString(value) || strings.ContainsRune(value, 0) || utf8.RuneCountInString(value) > max || (required && strings.TrimSpace(value) == "") {
			return false
		}
	}
	return true
}
func youtubeURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil {
		return "", errors.New("Use an HTTPS YouTube video link.")
	}
	var id string
	switch strings.ToLower(u.Host) {
	case "youtu.be":
		id = strings.Trim(u.Path, "/")
	case "youtube.com", "www.youtube.com", "m.youtube.com":
		if u.Path == "/watch" {
			id = u.Query().Get("v")
		} else {
			parts := strings.Split(strings.Trim(u.Path, "/"), "/")
			if len(parts) == 2 && (parts[0] == "shorts" || parts[0] == "embed" || parts[0] == "live") {
				id = parts[1]
			}
		}
	}
	if !youtubeID.MatchString(id) {
		return "", errors.New("Use a valid YouTube video link (watch, Shorts, or youtu.be).")
	}
	return "https://www.youtube.com/watch?v=" + id, nil
}
func validateWebsiteEntry(e *database.WebsiteEntry) error {
	if !websiteKinds[e.Kind] {
		return errors.New("Unknown website section.")
	}
	if e.Kind == "home" {
		e.Slug = "home"
	}
	if len(e.Slug) > 120 || !slugPattern.MatchString(e.Slug) {
		return errors.New("Use a URL slug containing lowercase letters, numbers and hyphens.")
	}
	if e.SortOrder < 0 || e.SortOrder > 100000 {
		return errors.New("Display order must be between 0 and 100000.")
	}
	c := &e.Content
	if !textValid(c.Name, 200, e.Kind != "home") || !textValid(c.Description, 5000, e.Published && e.Kind != "home") || !textValid(c.About, 20000, e.Published && e.Kind == "home") {
		return errors.New("Check text lengths and provide both German and English names; published content also needs both descriptions (or About us for Home).")
	}
	if len(c.Images) > 50 || len(c.Awards) > 200 || len(c.Videos) > 100 {
		return errors.New("Too many images, results or videos.")
	}
	seen := map[string]bool{}
	for _, image := range c.Images {
		if _, err := uuid.Parse(image.ID); err != nil || seen[image.ID] {
			return errors.New("Select valid, distinct images from the library.")
		}
		seen[image.ID] = true
		if !textValid(image.Alt, 1000, e.Published) {
			return errors.New("Published images need German and English alt text (up to 1000 characters).")
		}
	}
	if c.CoverImageID != "" && !seen[c.CoverImageID] {
		return errors.New("The cover image must be one of the selected images.")
	}
	if e.Published && (e.Kind == "teams" || e.Kind == "blog") && c.CoverImageID == "" {
		return errors.New("Choose a cover image before publishing.")
	}
	if e.Kind == "teams" && c.TeamStatus != "active" && c.TeamStatus != "retired" {
		return errors.New("Team status must be active or retired.")
	}
	if e.Kind == "events" {
		if _, err := time.Parse("2006-01-02", c.Date); err != nil {
			return errors.New("Provide an event date (YYYY-MM-DD); the year is derived from it.")
		}
	}
	for i := range c.Videos {
		canonical, err := youtubeURL(c.Videos[i].URL)
		if err != nil {
			return err
		}
		c.Videos[i].URL = canonical
		if !textValid(c.Videos[i].Title, 200, e.Published) {
			return errors.New("Provide a video title in German and English.")
		}
	}
	for _, award := range c.Awards {
		if _, err := uuid.Parse(award.EventID); err != nil {
			return errors.New("Select an event for every team result.")
		}
		if !textValid(award.League, 200, true) || !textValid(award.Result, 500, true) {
			return errors.New("Provide a league and result in German and English.")
		}
	}
	if c.URL != "" {
		u, err := url.Parse(c.URL)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || len(c.URL) > 2048 {
			return errors.New("External links must be valid HTTPS URLs.")
		}
	}
	if e.Kind == "blog" && e.Published && e.PublishAt == nil {
		return errors.New("Set the article publication date before publishing.")
	}
	if len(c.Blocks) > 200 {
		return errors.New("Use at most 200 article blocks.")
	}
	blockIDs := map[string]bool{}
	for i := range c.Blocks {
		block := &c.Blocks[i]
		if _, err := uuid.Parse(block.ID); err != nil || blockIDs[block.ID] {
			return errors.New("Article blocks need distinct IDs.")
		}
		blockIDs[block.ID] = true
		if !textValid(block.Text, 20000, e.Published && (block.Type == "text" || block.Type == "heading" || block.Type == "video")) {
			return errors.New("Complete article text in both languages before publishing.")
		}
		switch block.Type {
		case "text":
			block.Images = nil
			block.URL = ""
			block.Level = 0
		case "heading":
			if block.Level != 2 && block.Level != 3 {
				return errors.New("Heading level must be 2 or 3.")
			}
			block.Images = nil
			block.URL = ""
		case "image", "gallery":
			if len(block.Images) < 1 || len(block.Images) > 20 || (block.Type == "image" && len(block.Images) != 1) {
				return errors.New("Choose one image per image block, or 1–20 for a gallery.")
			}
			block.URL = ""
		case "video":
			canonical, err := youtubeURL(block.URL)
			if err != nil {
				return err
			}
			block.URL = canonical
			block.Images = nil
		default:
			return errors.New("Unsupported article block type.")
		}
		for _, image := range block.Images {
			if _, err := uuid.Parse(image.ID); err != nil || !textValid(image.Alt, 1000, e.Published) {
				return errors.New("Choose valid block images and provide both alt texts before publishing.")
			}
		}
	}
	if e.Kind == "blog" && e.Published && len(c.Blocks) == 0 {
		return errors.New("Add article content before publishing.")
	}
	// Only relevant fields are accepted so hidden data cannot create hidden dependencies.
	if e.Kind != "home" {
		c.About = database.LocalizedText{}
		c.Videos = nil
	}
	if e.Kind != "teams" {
		c.TeamStatus = ""
		c.Awards = nil
	}
	if e.Kind != "events" {
		c.Date = ""
	}
	if e.Kind != "blog" {
		c.Blocks = nil
		e.PublishAt = nil
	}
	if e.Kind != "sponsors" && e.Kind != "publications" {
		c.URL = ""
	}
	return nil
}

func websiteLanguage(raw string) (string, error) {
	if raw == "" {
		return "en", nil
	}
	if raw != "de" && raw != "en" {
		return "", fmt.Errorf("Unsupported language %q. Use lang=de or lang=en.", raw)
	}
	return raw, nil
}
