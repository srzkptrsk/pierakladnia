CREATE TABLE strings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    project_id VARCHAR(64) NULL,
    `key` VARCHAR(255) NULL,
    source_text TEXT NOT NULL,
    context TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE translations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    string_id INT NOT NULL,
    locale VARCHAR(32) NOT NULL DEFAULT 'target',
    current_text TEXT NOT NULL,
    status ENUM('todo', 'draft', 'needs_review', 'done') NOT NULL DEFAULT 'todo',
    updated_by INT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (string_id) REFERENCES strings(id) ON DELETE CASCADE,
    FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL,
    UNIQUE KEY unique_string_locale (string_id, locale)
);

CREATE TABLE translation_revisions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    translation_id INT NOT NULL,
    old_text TEXT NOT NULL,
    new_text TEXT NOT NULL,
    changed_by INT NULL,
    changed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (translation_id) REFERENCES translations(id) ON DELETE CASCADE,
    FOREIGN KEY (changed_by) REFERENCES users(id) ON DELETE SET NULL
);
