package db

import (
	"database/sql"
)

func CreateProject(db *sql.DB, name, description string) (int, error) {
	res, err := db.Exec("INSERT INTO projects (name, description) VALUES (?, ?)", name, description)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func UpdateProject(db *sql.DB, id int, name, description, exportFilename, poProjectIdVersion, poReportMsgidBugsTo, poLanguageTeam, poLanguage, poLastTranslator string) error {
	_, err := db.Exec(`
		UPDATE projects 
		SET name = ?, description = ?, export_filename = ?, po_project_id_version = ?, 
		    po_report_msgid_bugs_to = ?, po_language_team = ?, po_language = ?, po_last_translator = ? 
		WHERE id = ?
	`, name, description, exportFilename, poProjectIdVersion, poReportMsgidBugsTo, poLanguageTeam, poLanguage, poLastTranslator, id)
	return err
}

func GetProjectByID(db *sql.DB, id int) (*Project, error) {
	var p Project
	err := db.QueryRow(`
		SELECT id, name, description, export_filename, po_project_id_version, po_report_msgid_bugs_to, 
		       po_language_team, po_language, po_last_translator, created_at, updated_at 
		FROM projects WHERE id = ?
	`, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.ExportFilename, &p.PoProjectIdVersion, &p.PoReportMsgidBugsTo,
		&p.PoLanguageTeam, &p.PoLanguage, &p.PoLastTranslator, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &p, nil
}

func GetProjectsForUser(db *sql.DB, userID int, role string) ([]*Project, error) {
	var rows *sql.Rows
	var err error

	if role == "admin" {
		rows, err = db.Query(`
			SELECT id, name, description, export_filename, po_project_id_version, po_report_msgid_bugs_to, 
			       po_language_team, po_language, po_last_translator, created_at, updated_at 
			FROM projects ORDER BY name ASC
		`)
	} else {
		rows, err = db.Query(`
			SELECT p.id, p.name, p.description, p.export_filename, p.po_project_id_version, p.po_report_msgid_bugs_to, 
			       p.po_language_team, p.po_language, p.po_last_translator, p.created_at, p.updated_at 
			FROM projects p
			JOIN project_users pu ON p.id = pu.project_id
			WHERE pu.user_id = ?
			ORDER BY p.name ASC
		`, userID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		p := &Project{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.ExportFilename, &p.PoProjectIdVersion, &p.PoReportMsgidBugsTo,
			&p.PoLanguageTeam, &p.PoLanguage, &p.PoLastTranslator, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func AssignUserToProject(db *sql.DB, projectID, userID int) error {
	_, err := db.Exec("INSERT INTO project_users (project_id, user_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE user_id=user_id", projectID, userID)
	return err
}

func RemoveUserFromProject(db *sql.DB, projectID, userID int) error {
	_, err := db.Exec("DELETE FROM project_users WHERE project_id = ? AND user_id = ?", projectID, userID)
	return err
}

func GetAllProjects(db *sql.DB) ([]*Project, error) {
	rows, err := db.Query(`
		SELECT id, name, description, export_filename, po_project_id_version, po_report_msgid_bugs_to, 
		       po_language_team, po_language, po_last_translator, created_at, updated_at 
		FROM projects ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		p := &Project{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.ExportFilename, &p.PoProjectIdVersion, &p.PoReportMsgidBugsTo,
			&p.PoLanguageTeam, &p.PoLanguage, &p.PoLastTranslator, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func GetProjectIDsForUser(db *sql.DB, userID int) ([]int, error) {
	rows, err := db.Query("SELECT project_id FROM project_users WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
