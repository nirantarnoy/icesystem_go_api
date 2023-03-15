package entity

type User struct {
	ID                  uint64 `json:"id"`
	Dns_user            string `json:"dns_user"`
	Token               string `gorm:"-" json:"token,omitempty"`
	CurrentTeamID       string `json:"current_team_id"`
	CurrentSafetyTeamID string `json:"current_safety_team_id"`
	Photo               string `json:"photo"`
	SectionCode         string `json:"section_code"`
}
