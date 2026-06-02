package user

import (
	"time"

	"chelbo/backend/internal/models"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// GetByID returns a user by ID
func (r *Repository) GetByID(id uint64) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, name, bio, avatar_url, is_active, is_verified, last_seen, created_at, updated_at
		FROM users 
		WHERE id = ? AND is_active = TRUE
	`
	err := r.db.Get(&user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail returns a user by email
func (r *Repository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, name, bio, avatar_url, is_active, is_verified, last_seen, created_at, updated_at
		FROM users 
		WHERE email = ? AND is_active = TRUE
	`
	err := r.db.Get(&user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateProfile updates user profile
func (r *Repository) UpdateProfile(userID uint64, name *string, bio *string) error {
	updates := make(map[string]interface{})
	if name != nil && *name != "" {
		updates["name"] = *name
	}
	if bio != nil {
		updates["bio"] = *bio
	}
	updates["updated_at"] = time.Now()

	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE users SET "
	args := make([]interface{}, 0)
	i := 0
	for key, value := range updates {
		if i > 0 {
			query += ", "
		}
		query += key + " = ?"
		args = append(args, value)
		i++
	}
	query += " WHERE id = ?"
	args = append(args, userID)

	_, err := r.db.Exec(query, args...)
	return err
}

// UpdateAvatar updates user avatar URL
func (r *Repository) UpdateAvatar(userID uint64, avatarURL string) error {
	query := "UPDATE users SET avatar_url = ?, updated_at = ? WHERE id = ?"
	_, err := r.db.Exec(query, avatarURL, time.Now(), userID)
	return err
}

// UpdateLastSeen updates user's last seen timestamp
func (r *Repository) UpdateLastSeen(userID uint64) error {
	query := "UPDATE users SET last_seen = ? WHERE id = ?"
	_, err := r.db.Exec(query, time.Now(), userID)
	return err
}

// SearchUsers searches users by name or email
func (r *Repository) SearchUsers(queryStr string, limit int, excludeUserID uint64) ([]models.UserResponse, error) {
	var users []models.UserResponse

	sqlQuery := `
		SELECT id, email, name, bio, last_seen
		FROM users 
		WHERE (name LIKE ? OR email LIKE ?) 
		AND id != ? 
		AND is_active = TRUE
		LIMIT ?
	`
	searchPattern := "%" + queryStr + "%"
	err := r.db.Select(&users, sqlQuery, searchPattern, searchPattern, excludeUserID, limit)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// GetOnlineStatuses returns online statuses for multiple users
func (r *Repository) GetOnlineStatuses(userIDs []uint64) (map[uint64]bool, error) {
	if len(userIDs) == 0 {
		return make(map[uint64]bool), nil
	}

	query := "SELECT id, last_seen FROM users WHERE id IN (?)"
	query, args, err := sqlx.In(query, userIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	var users []struct {
		ID       uint64    `db:"id"`
		LastSeen time.Time `db:"last_seen"`
	}
	err = r.db.Select(&users, query, args...)
	if err != nil {
		return nil, err
	}

	statuses := make(map[uint64]bool)
	now := time.Now()
	for _, u := range users {
		statuses[u.ID] = now.Sub(u.LastSeen) < 5*time.Minute
	}
	return statuses, nil
}
