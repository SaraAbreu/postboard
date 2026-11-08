package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/SaraAbreu/postboard/api/internal/models"
)

type PostgresStore struct {
	db *pgxpool.Pool
}

func NewPostgresStore(db *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{db: db}
}

const cols = `id, title, content, platform, status, scheduled_at, created_at, updated_at`

func scanPost(row pgx.Row) (models.Post, error) {
	var p models.Post
	err := row.Scan(&p.ID, &p.Title, &p.Content, &p.Platform, &p.Status, &p.ScheduledAt, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, ErrNotFound
	}
	return p, err
}

func (s *PostgresStore) List(ctx context.Context) ([]models.Post, error) {
	rows, err := s.db.Query(ctx, `SELECT `+cols+` FROM posts ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (s *PostgresStore) Get(ctx context.Context, id int64) (models.Post, error) {
	return scanPost(s.db.QueryRow(ctx, `SELECT `+cols+` FROM posts WHERE id = $1`, id))
}

func (s *PostgresStore) Create(ctx context.Context, in models.PostInput) (models.Post, error) {
	return scanPost(s.db.QueryRow(ctx,
		`INSERT INTO posts (title, content, platform, status, scheduled_at)
		 VALUES ($1, $2, $3, $4, $5) RETURNING `+cols,
		in.Title, in.Content, in.Platform, in.Status, in.ScheduledAt))
}

func (s *PostgresStore) Update(ctx context.Context, id int64, in models.PostInput) (models.Post, error) {
	return scanPost(s.db.QueryRow(ctx,
		`UPDATE posts SET title=$1, content=$2, platform=$3, status=$4, scheduled_at=$5, updated_at=now()
		 WHERE id=$6 RETURNING `+cols,
		in.Title, in.Content, in.Platform, in.Status, in.ScheduledAt, id))
}

func (s *PostgresStore) Delete(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM posts WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
