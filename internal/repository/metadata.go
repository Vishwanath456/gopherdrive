package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Metadata struct {
	ID        string `json:"id"`
	FileName  string `json:"filename"`
	FilePath  string `json:"filepath"`
	SHA256    string `json:"sha256"`
	Size      int64  `json:"size"`
	Extension string `json:"extension"`
	Status    string `json:"status"`
}
type MetadataRepo interface {
	Save(ctx context.Context, m *Metadata) error
	Update(ctx context.Context, m *Metadata) error
	GetByID(ctx context.Context, id string) (*Metadata, error)
	ListAll(ctx context.Context) ([]*Metadata, error) // 🔥 MUST BE HERE
}

type mysqlRepo struct {
	db *sql.DB
}

func NewMySQLRepo(db *sql.DB) MetadataRepo {
	return &mysqlRepo{db: db}
}

// -------------------- GetByID --------------------

func (r *mysqlRepo) GetByID(ctx context.Context, id string) (*Metadata, error) {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `SELECT id, filename, filepath, sha256, size, extension, status 
	          FROM metadata WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)

	var m Metadata
	var sha sql.NullString
	var size sql.NullInt64
	var ext sql.NullString

	err := row.Scan(
		&m.ID,
		&m.FileName,
		&m.FilePath,
		&sha,
		&size,
		&ext,
		&m.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file not found: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch metadata: %w", err)
	}

	if sha.Valid {
		m.SHA256 = sha.String
	}
	if size.Valid {
		m.Size = size.Int64
	}
	if ext.Valid {
		m.Extension = ext.String
	}

	return &m, nil
}

// -------------------- Save --------------------

func (r *mysqlRepo) Save(ctx context.Context, m *Metadata) error {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `INSERT INTO metadata (id, filename, filepath, status)
	          VALUES (?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		m.ID,
		m.FileName,
		m.FilePath,
		m.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// -------------------- Update --------------------

func (r *mysqlRepo) Update(ctx context.Context, m *Metadata) error {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `UPDATE metadata 
	          SET sha256 = ?, size = ?, extension = ?, status = ?
	          WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query,
		m.SHA256,
		m.Size,
		m.Extension,
		m.Status,
		m.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	return nil
}

// -------------------- ListAll --------------------

func (r *mysqlRepo) ListAll(ctx context.Context) ([]*Metadata, error) {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `SELECT id, filename, filepath, sha256, size, extension, status FROM metadata`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Metadata

	for rows.Next() {
		var m Metadata
		var sha sql.NullString
		var size sql.NullInt64
		var ext sql.NullString

		err := rows.Scan(
			&m.ID,
			&m.FileName,
			&m.FilePath,
			&sha,
			&size,
			&ext,
			&m.Status,
		)
		if err != nil {
			return nil, err
		}

		if sha.Valid {
			m.SHA256 = sha.String
		}
		if size.Valid {
			m.Size = size.Int64
		}
		if ext.Valid {
			m.Extension = ext.String
		}

		result = append(result, &m)
	}

	return result, nil
}
