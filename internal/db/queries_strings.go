package db

import (
	"database/sql"
	"strconv"
	"strings"
)

func CreateString(db *sql.DB, projectID int, key, sourceText, contextStr string) (int, error) {
	res, err := db.Exec("INSERT INTO strings (project_id, `key`, source_text, context) VALUES (?, ?, ?, ?)", projectID, key, sourceText, contextStr)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func GetAllStrings(db *sql.DB, projectID int) ([]String, error) {
	rows, err := db.Query(`
		SELECT id, project_id, `+"`key`"+`, source_text, context, created_at, updated_at
		FROM strings
		WHERE project_id = ?
		ORDER BY id ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var strings []String
	for rows.Next() {
		var s String
		if err := rows.Scan(
			&s.ID, &s.ProjectID, &s.Key, &s.SourceText, &s.Context, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		strings = append(strings, s)
	}
	return strings, nil
}

func applyIDFilter(q *string, args *[]interface{}, idFilter string) {
	if idFilter == "" {
		return
	}
	idFilter = strings.ReplaceAll(idFilter, " ", "")
	if strings.HasPrefix(idFilter, ">=") {
		if val, err := strconv.Atoi(idFilter[2:]); err == nil {
			*q += ` AND s.id >= ?`
			*args = append(*args, val)
		}
	} else if strings.HasPrefix(idFilter, "<=") {
		if val, err := strconv.Atoi(idFilter[2:]); err == nil {
			*q += ` AND s.id <= ?`
			*args = append(*args, val)
		}
	} else if strings.HasPrefix(idFilter, ">") {
		if val, err := strconv.Atoi(idFilter[1:]); err == nil {
			*q += ` AND s.id > ?`
			*args = append(*args, val)
		}
	} else if strings.HasPrefix(idFilter, "<") {
		if val, err := strconv.Atoi(idFilter[1:]); err == nil {
			*q += ` AND s.id < ?`
			*args = append(*args, val)
		}
	} else if strings.Contains(idFilter, "-") {
		parts := strings.Split(idFilter, "-")
		if len(parts) == 2 {
			if val1, err1 := strconv.Atoi(parts[0]); err1 == nil {
				if val2, err2 := strconv.Atoi(parts[1]); err2 == nil {
					*q += ` AND s.id BETWEEN ? AND ?`
					*args = append(*args, val1, val2)
				}
			}
		}
	} else if strings.Contains(idFilter, ",") {
		var ids []string
		for _, p := range strings.Split(idFilter, ",") {
			if _, err := strconv.Atoi(p); err == nil {
				ids = append(ids, p)
			}
		}
		if len(ids) > 0 {
			*q += ` AND s.id IN (` + strings.Join(ids, ",") + `)`
		}
	} else if val, err := strconv.Atoi(idFilter); err == nil {
		*q += ` AND s.id = ?`
		*args = append(*args, val)
	}
}

func CountStrings(db *sql.DB, projectID int, sourceQuery, targetQuery, status, idFilter string) (int, error) {
	q := `SELECT COUNT(*) FROM strings s`

	if status != "" || sourceQuery != "" || targetQuery != "" {
		q += ` LEFT JOIN translations t ON s.id = t.string_id AND t.locale = 'target'`
	}

	q += ` WHERE s.project_id = ?`
	args := []interface{}{projectID}

	if sourceQuery != "" {
		q += ` AND (LOWER(s.source_text) LIKE ? OR LOWER(s.` + "`key`" + `) LIKE ?)`
		args = append(args, "%"+sourceQuery+"%", "%"+sourceQuery+"%")
	}

	if targetQuery != "" {
		q += ` AND LOWER(t.current_text) LIKE ?`
		args = append(args, "%"+targetQuery+"%")
	}

	if status != "" {
		if status == "untranslated" {
			q += ` AND t.id IS NULL`
		} else {
			q += ` AND t.status = ?`
			args = append(args, status)
		}
	}

	applyIDFilter(&q, &args, idFilter)

	var count int
	err := db.QueryRow(q, args...).Scan(&count)
	return count, err
}

func GetStringsPaginated(db *sql.DB, projectID int, sourceQuery, targetQuery, status, sort, idFilter string, limit, offset int) ([]StringWithTranslation, error) {
	q := `
		SELECT 
			s.id, s.project_id, s.key, s.source_text, s.context, s.created_at, s.updated_at,
			t.id as target_translation_id, t.current_text as target_translation_text, t.status as target_translation_status,
			(SELECT COUNT(*) FROM comments c WHERE c.entity_type = 'translation' AND c.entity_id = s.id) as comments_count
		FROM strings s
		LEFT JOIN translations t ON s.id = t.string_id AND t.locale = 'target'
		WHERE s.project_id = ?
	`
	args := []interface{}{projectID}

	if sourceQuery != "" {
		q += ` AND (LOWER(s.source_text) LIKE ? OR LOWER(s.` + "`key`" + `) LIKE ?)`
		args = append(args, "%"+sourceQuery+"%", "%"+sourceQuery+"%")
	}

	if targetQuery != "" {
		q += ` AND LOWER(t.current_text) LIKE ?`
		args = append(args, "%"+targetQuery+"%")
	}

	if status != "" {
		if status == "untranslated" {
			q += ` AND t.id IS NULL`
		} else {
			q += ` AND t.status = ?`
			args = append(args, status)
		}
	}

	applyIDFilter(&q, &args, idFilter)

	if sort == "comments_desc" {
		q += ` ORDER BY comments_count DESC, s.id ASC LIMIT ? OFFSET ?`
	} else {
		q += ` ORDER BY s.id ASC LIMIT ? OFFSET ?`
	}

	args = append(args, limit, offset)

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var strings []StringWithTranslation
	for rows.Next() {
		var s StringWithTranslation
		if err := rows.Scan(
			&s.ID, &s.ProjectID, &s.Key, &s.SourceText, &s.Context, &s.CreatedAt, &s.UpdatedAt,
			&s.TargetTranslationID, &s.TargetTranslationText, &s.TargetTranslationStatus, &s.CommentsCount,
		); err != nil {
			return nil, err
		}
		strings = append(strings, s)
	}
	return strings, nil
}

func GetStringByID(db *sql.DB, projectID int, id int) (*String, error) {
	var s String
	err := db.QueryRow(`
		SELECT id, project_id, `+"`key`"+`, source_text, context, created_at, updated_at
		FROM strings WHERE project_id = ? AND id = ?
	`, projectID, id).Scan(&s.ID, &s.ProjectID, &s.Key, &s.SourceText, &s.Context, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func GetTranslationsForString(db *sql.DB, stringID int) ([]Translation, error) {
	rows, err := db.Query(`
		SELECT id, string_id, locale, current_text, status, updated_by, updated_at, created_at
		FROM translations WHERE string_id = ?
		ORDER BY locale ASC
	`, stringID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var translations []Translation
	for rows.Next() {
		var t Translation
		if err := rows.Scan(
			&t.ID, &t.StringID, &t.Locale, &t.CurrentText, &t.Status,
			&t.UpdatedBy, &t.UpdatedAt, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		translations = append(translations, t)
	}
	return translations, nil
}

func GetRevisionsForTranslation(db *sql.DB, translationID int) ([]TranslationRevision, error) {
	rows, err := db.Query(`
		SELECT id, translation_id, old_text, new_text, changed_by, changed_at
		FROM translation_revisions WHERE translation_id = ?
		ORDER BY changed_at DESC
	`, translationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var revisions []TranslationRevision
	for rows.Next() {
		var r TranslationRevision
		if err := rows.Scan(
			&r.ID, &r.TranslationID, &r.OldText, &r.NewText, &r.ChangedBy, &r.ChangedAt,
		); err != nil {
			return nil, err
		}
		revisions = append(revisions, r)
	}
	return revisions, nil
}

func UpsertTranslation(db *sql.DB, stringID int, locale, newText string, isImport bool, userID int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if exists
	var existingID int
	var oldText, status string
	err = tx.QueryRow(`SELECT id, current_text, status FROM translations WHERE string_id = ? AND locale = ?`, stringID, locale).Scan(&existingID, &oldText, &status)

	if err == sql.ErrNoRows {
		// New translation
		newStatus := "draft"
		if isImport {
			newStatus = "todo"
		}
		res, err := tx.Exec(`
			INSERT INTO translations (string_id, locale, current_text, status, updated_by)
			VALUES (?, ?, ?, ?, ?)
		`, stringID, locale, newText, newStatus, userID)
		if err != nil {
			return err
		}
		tID, _ := res.LastInsertId()

		_, err = tx.Exec(`
			INSERT INTO translation_revisions (translation_id, old_text, new_text, changed_by)
			VALUES (?, ?, ?, ?)
		`, tID, "", newText, userID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		newStatus := status
		if isImport {
			newStatus = "todo"
		} else if status == "todo" {
			newStatus = "draft"
		}
		
		if oldText == newText && newStatus == status {
			return tx.Commit()
		}

		// Update existing
		_, err = tx.Exec(`
			UPDATE translations SET current_text = ?, status = ?, updated_by = ? WHERE id = ?
		`, newText, newStatus, userID, existingID)
		if err != nil {
			return err
		}

		if oldText != newText {
			_, err = tx.Exec(`
				INSERT INTO translation_revisions (translation_id, old_text, new_text, changed_by)
				VALUES (?, ?, ?, ?)
			`, existingID, oldText, newText, userID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func UpdateTranslationStatus(db *sql.DB, stringID int, locale, status string) error {
	_, err := db.Exec(`
		UPDATE translations SET status = ? WHERE string_id = ? AND locale = ?
	`, status, stringID, locale)
	return err
}

func UpdateStringContext(db *sql.DB, projectID int, stringID int, context string) error {
	_, err := db.Exec(`
		UPDATE strings SET context = ? WHERE project_id = ? AND id = ?
	`, context, projectID, stringID)
	return err
}

func GetAllStringsForExport(db *sql.DB, projectID int) (map[string]string, error) {
	rows, err := db.Query(`
		SELECT s.source_text, COALESCE(t.current_text, '') as translated_text
		FROM strings s
		LEFT JOIN translations t ON s.id = t.string_id AND t.locale = 'target'
		WHERE s.project_id = ?
		ORDER BY s.id ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var source, translated string
		if err := rows.Scan(&source, &translated); err != nil {
			return nil, err
		}
		result[source] = translated
	}
	return result, nil
}

func GetAllStringsWithTranslationsForExportPO(db *sql.DB, projectID int) ([]StringWithTranslation, error) {
	q := `
		SELECT 
			s.id, s.project_id, s.key, s.source_text, s.context, s.created_at, s.updated_at,
			t.id as target_translation_id, t.current_text as target_translation_text, t.status as target_translation_status,
			0 as comments_count
		FROM strings s
		LEFT JOIN translations t ON s.id = t.string_id AND t.locale = 'target'
		WHERE s.project_id = ?
		ORDER BY s.id ASC
	`
	rows, err := db.Query(q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stringsSlice []StringWithTranslation
	for rows.Next() {
		var s StringWithTranslation
		if err := rows.Scan(
			&s.ID, &s.ProjectID, &s.Key, &s.SourceText, &s.Context, &s.CreatedAt, &s.UpdatedAt,
			&s.TargetTranslationID, &s.TargetTranslationText, &s.TargetTranslationStatus, &s.CommentsCount,
		); err != nil {
			return nil, err
		}
		stringsSlice = append(stringsSlice, s)
	}
	return stringsSlice, nil
}

func GetProjectStatistics(db *sql.DB, projectID int) (map[string]int, int, error) {
	q := `
		SELECT COALESCE(t.status, 'untranslated') as current_status, COUNT(*) as count
		FROM strings s
		LEFT JOIN translations t ON s.id = t.string_id AND t.locale = 'target'
		WHERE s.project_id = ?
		GROUP BY current_status
	`
	rows, err := db.Query(q, projectID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	total := 0
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, 0, err
		}
		stats[status] = count
		total += count
	}
	return stats, total, nil
}
