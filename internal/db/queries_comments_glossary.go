package db

import "database/sql"

func GetCommentsForEntity(db *sql.DB, entityType string, entityID int) ([]Comment, error) {
	rows, err := db.Query(`
		SELECT c.id, c.entity_type, c.entity_id, c.parent_id, c.user_id, c.content, c.created_at, c.updated_at,
		       u.email as author_email
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.entity_type = ? AND c.entity_id = ?
		ORDER BY c.created_at ASC
	`, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(
			&c.ID, &c.EntityType, &c.EntityID, &c.ParentID, &c.UserID, &c.Content,
			&c.CreatedAt, &c.UpdatedAt, &c.AuthorEmail,
		); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func CreateComment(db *sql.DB, entityType string, entityID int, parentID *int, userID int, content string) error {
	_, err := db.Exec(`
		INSERT INTO comments (entity_type, entity_id, parent_id, user_id, content)
		VALUES (?, ?, ?, ?, ?)
	`, entityType, entityID, parentID, userID, content)
	return err
}

func GetAllGlossaryTerms(db *sql.DB, projectID int) ([]GlossaryTerm, error) {
	rows, err := db.Query(`
		SELECT id, project_id, category, source_term, target_term, description, created_at, updated_at
		FROM glossary_terms
		WHERE project_id = ?
		ORDER BY source_term ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var terms []GlossaryTerm
	for rows.Next() {
		var t GlossaryTerm
		if err := rows.Scan(
			&t.ID, &t.ProjectID, &t.Category, &t.SourceTerm, &t.TargetTerm, &t.Description, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		terms = append(terms, t)
	}
	return terms, nil
}

func CountGlossaryTerms(db *sql.DB, projectID int, category string) (int, error) {
	query := `SELECT COUNT(*) FROM glossary_terms WHERE project_id = ?`
	args := []interface{}{projectID}

	if category != "" {
		query += ` AND category = ?`
		args = append(args, category)
	}

	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	return count, err
}

func GetGlossaryTermsPaginated(db *sql.DB, projectID int, category string, limit int, offset int) ([]GlossaryTerm, error) {
	query := `
		SELECT id, project_id, category, source_term, target_term, description, created_at, updated_at
		FROM glossary_terms
		WHERE project_id = ?
	`
	args := []interface{}{projectID}

	if category != "" {
		query += ` AND category = ?`
		args = append(args, category)
	}

	query += ` ORDER BY source_term ASC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var terms []GlossaryTerm
	for rows.Next() {
		var t GlossaryTerm
		if err := rows.Scan(
			&t.ID, &t.ProjectID, &t.Category, &t.SourceTerm, &t.TargetTerm, &t.Description, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		terms = append(terms, t)
	}
	return terms, nil
}

func GetGlossaryTermByID(db *sql.DB, projectID int, id int) (*GlossaryTerm, error) {
	var t GlossaryTerm
	err := db.QueryRow(`
		SELECT id, project_id, category, source_term, target_term, description, created_at, updated_at
		FROM glossary_terms WHERE project_id = ? AND id = ?
	`, projectID, id).Scan(
		&t.ID, &t.ProjectID, &t.Category, &t.SourceTerm, &t.TargetTerm, &t.Description, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func CreateGlossaryTerm(db *sql.DB, projectID int, category, sourceTerm, targetTerm string, description *string) error {
	_, err := db.Exec(`
		INSERT INTO glossary_terms (project_id, category, source_term, target_term, description)
		VALUES (?, ?, ?, ?, ?)
	`, projectID, category, sourceTerm, targetTerm, description)
	return err
}

func UpdateGlossaryTerm(db *sql.DB, projectID int, id int, category, sourceTerm, targetTerm string, description *string) error {
	_, err := db.Exec(`
		UPDATE glossary_terms SET category = ?, source_term = ?, target_term = ?, description = ?
		WHERE project_id = ? AND id = ?
	`, category, sourceTerm, targetTerm, description, projectID, id)
	return err
}
