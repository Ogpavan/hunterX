package models

import "time"

// Recruiter represents a company or recruiter entity
// extracted from hiring platforms and enriched via external sources.
type Recruiter struct {
	// Source platform: naukri | indeed | internshala | manual
	Source string `json:"source"`

	// Platform-specific identifier (if available)
	SourceID string `json:"source_id,omitempty"`

	// Core identity
	CompanyName string `json:"company_name"`
	RecruiterName string `json:"recruiter_name,omitempty"`

	// Hiring context
	JobTitle string `json:"job_title,omitempty"`
	Location string `json:"location,omitempty"`

	// Metadata
	PostedDate time.Time `json:"posted_date,omitempty"`

	// Enrichment (Google Maps / Web)
	Phone   string `json:"phone,omitempty"`
	Email   string `json:"email,omitempty"`
	Website string `json:"website,omitempty"`

	// Internal
	CollectedAt time.Time `json:"collected_at"`
}
