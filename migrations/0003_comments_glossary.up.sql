CREATE TABLE glossary_terms (
    id INT AUTO_INCREMENT PRIMARY KEY,
    category ENUM('character', 'place', 'item', 'term', 'other') NOT NULL DEFAULT 'other',
    source_term VARCHAR(255) NOT NULL,
    target_term VARCHAR(255) NOT NULL,
    description TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE comments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    entity_type ENUM('translation', 'glossary_term') NOT NULL,
    entity_id INT NOT NULL,
    parent_id INT NULL,
    user_id INT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE
);

-- Glossary Links (optional)
CREATE TABLE IF NOT EXISTS glossary_links (
    id INT AUTO_INCREMENT PRIMARY KEY,
    term_id INT NOT NULL,
    string_id INT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (term_id) REFERENCES glossary_terms(id) ON DELETE CASCADE,
    FOREIGN KEY (string_id) REFERENCES strings(id) ON DELETE CASCADE,
    UNIQUE KEY (term_id, string_id)
);
