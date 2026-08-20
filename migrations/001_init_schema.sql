CREATE TABLE users (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  email VARCHAR(128) NOT NULL,
  nickname VARCHAR(8) NOT NULL,
  avatar_url VARCHAR(512) NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uk_users_email (email),
  UNIQUE KEY uk_users_nickname (nickname)
);

CREATE TABLE email_verification_codes (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NULL,
  email VARCHAR(128) NOT NULL,
  purpose VARCHAR(32) NOT NULL,
  code VARCHAR(16) NOT NULL,
  used_at DATETIME NULL,
  expires_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL,
  KEY idx_email_codes_email_purpose_created_at (email, purpose, created_at)
);

CREATE TABLE rooms (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(80) NOT NULL,
  owner_id BIGINT UNSIGNED NOT NULL,
  max_members TINYINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  KEY idx_rooms_created_at (created_at)
);

CREATE TABLE room_members (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  room_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  mic_status VARCHAR(16) NOT NULL DEFAULT 'off',
  joined_at DATETIME NOT NULL,
  UNIQUE KEY uk_room_members_user (user_id),
  KEY idx_room_members_room_joined_at (room_id, joined_at)
);

CREATE TABLE messages (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  room_id BIGINT UNSIGNED NOT NULL,
  sender_id BIGINT UNSIGNED NOT NULL,
  type VARCHAR(16) NOT NULL DEFAULT 'text',
  content TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  KEY idx_messages_room_id_id (room_id, id)
);
CREATE TABLE app_versions (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  version_code BIGINT UNSIGNED NOT NULL,
  apk_url VARCHAR(512) NOT NULL,
  apk_sha256 CHAR(64) NOT NULL,
  release_notes TEXT NOT NULL,
  is_published TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uk_app_versions_version_code (version_code),
  KEY idx_app_versions_published_version (is_published, version_code)
);