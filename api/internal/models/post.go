package models

import "time"

type Post struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Platform    string     `json:"platform"` // instagram, facebook, linkedin
	Status      string     `json:"status"`   // draft, scheduled, published
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PostInput es lo que recibe la API al crear/editar.
type PostInput struct {
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Platform    string     `json:"platform"`
	Status      string     `json:"status"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

var validPlatforms = map[string]bool{"instagram": true, "facebook": true, "linkedin": true}
var validStatuses = map[string]bool{"draft": true, "scheduled": true, "published": true}

// Validate devuelve un mapa campo -> error (vacío si todo es correcto).
func (in PostInput) Validate() map[string]string {
	errs := map[string]string{}
	if in.Title == "" {
		errs["title"] = "es obligatorio"
	} else if len(in.Title) > 200 {
		errs["title"] = "máximo 200 caracteres"
	}
	if !validPlatforms[in.Platform] {
		errs["platform"] = "debe ser instagram, facebook o linkedin"
	}
	if !validStatuses[in.Status] {
		errs["status"] = "debe ser draft, scheduled o published"
	}
	if in.Status == "scheduled" && in.ScheduledAt == nil {
		errs["scheduled_at"] = "es obligatorio si status es scheduled"
	}
	return errs
}
