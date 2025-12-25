package repository

import (
    "health-bar/shared/models"
    "github.com/jmoiron/sqlx"
    "github.com/google/uuid"
)

type AuthRepository struct {
    db *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) *AuthRepository {
    return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUser(email, passwordHash string, role models.UserRole) (*models.User, error) {
    user := &models.User{
        ID:           uuid.New().String(),
        Email:        email,
        PasswordHash: passwordHash,
        Role:         role,
    }

    query := `
        INSERT INTO users (id, email, password_hash, role)
        VALUES ($1, $2, $3, $4)
        RETURNING id, email, role, created_at, updated_at
    `

    err := r.db.QueryRowx(query, user.ID, user.Email, user.PasswordHash, user.Role).
        StructScan(user)

    if err != nil {
        return nil, err
    }

    return user, nil
}

func (r *AuthRepository) GetUserByEmail(email string) (*models.User, error) {
    user := &models.User{}
    query := `SELECT id, email, password_hash, role, created_at, updated_at FROM users WHERE email = $1`
    
    err := r.db.Get(user, query, email)
    if err != nil {
        return nil, err
    }

    return user, nil
}

func (r *AuthRepository) GetUserByID(id string) (*models.User, error) {
    user := &models.User{}
    query := `SELECT id, email, role, created_at, updated_at FROM users WHERE id = $1`
    
    err := r.db.Get(user, query, id)
    if err != nil {
        return nil, err
    }

    return user, nil
}