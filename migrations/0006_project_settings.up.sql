ALTER TABLE projects
ADD COLUMN export_filename VARCHAR(255) DEFAULT '',
ADD COLUMN po_project_id_version VARCHAR(255) DEFAULT '',
ADD COLUMN po_report_msgid_bugs_to VARCHAR(255) DEFAULT '',
ADD COLUMN po_language_team VARCHAR(255) DEFAULT '',
ADD COLUMN po_language VARCHAR(255) DEFAULT '',
ADD COLUMN po_last_translator VARCHAR(255) DEFAULT '';
