package db

import "time"

type User struct {
	ID              int        `json:"id"`
	Email           string     `json:"email"`
	PasswordHash    string     `json:"-"`
	Role            string     `json:"role"`
	CanTranslate    bool       `json:"can_translate"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Project struct {
	ID                  int       `json:"id"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	ExportFilename      string    `json:"export_filename"`
	PoProjectIdVersion  string    `json:"po_project_id_version"`
	PoReportMsgidBugsTo string    `json:"po_report_msgid_bugs_to"`
	PoLanguageTeam      string    `json:"po_language_team"`
	PoLanguage          string    `json:"po_language"`
	PoLastTranslator    string    `json:"po_last_translator"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type ProjectUser struct {
	ProjectID int       `json:"project_id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type String struct {
	ID         int       `json:"id"`
	ProjectID  int       `json:"project_id"`
	Key        *string   `json:"key"`
	SourceText string    `json:"source_text"`
	Context    *string   `json:"context"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type StringWithTranslation struct {
	String
	TargetTranslationText   *string `json:"target_translation_text"`
	TargetTranslationID     *int    `json:"target_translation_id"`
	TargetTranslationStatus *string `json:"target_translation_status"`
	CommentsCount           int     `json:"comments_count"`
}

func (s StringWithTranslation) StatusRaw() string {
	if s.TargetTranslationStatus == nil {
		return ""
	}
	return *s.TargetTranslationStatus
}

type Translation struct {
	ID          int       `json:"id"`
	StringID    int       `json:"string_id"`
	Locale      string    `json:"locale"`
	CurrentText string    `json:"current_text"`
	Status      string    `json:"status"`
	UpdatedBy   *int      `json:"updated_by"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type TranslationRevision struct {
	ID            int       `json:"id"`
	TranslationID int       `json:"translation_id"`
	OldText       string    `json:"old_text"`
	NewText       string    `json:"new_text"`
	ChangedBy     *int      `json:"changed_by"`
	ChangedAt     time.Time `json:"changed_at"`
}

type Comment struct {
	ID         int       `json:"id"`
	EntityType string    `json:"entity_type"`
	EntityID   int       `json:"entity_id"`
	ParentID   *int      `json:"parent_id"`
	UserID     int       `json:"user_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	AuthorEmail string `json:"author_email"` // Loaded via JOIN
}

type GlossaryTerm struct {
	ID          int       `json:"id"`
	ProjectID   int       `json:"project_id"`
	Category    string    `json:"category"`
	SourceTerm  string    `json:"source_term"`
	TargetTerm  string    `json:"target_term"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
