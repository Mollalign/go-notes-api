package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"go-notes-api/internal/services"
)

type NotesRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) *NotesRepository {
	return &NotesRepository{
		db: db,
	}
}

func (r *NotesRepository) Create(
	ctx context.Context,
	input services.CreateInput,
) (services.Note, error) {
	const query = `
	   INSERT INTO notes (title, content)
	   VALUES ($1, $2)
	   RETURNING id, title, content, created_at
	`

	var note services.Note

	err := r.db.QueryRowContext(
		ctx,
		query,
		input.Title,
		input.Content,
	).Scan(
		&note.ID,
		&note.Title,
		&note.Content,
		&note.CreatedAt,
	)

	if err != nil {
		return services.Note{}, fmt.Errorf(
			"create note: %w",
			err,
		)
	}

	return note, nil
}

func (r *NotesRepository) List(
	ctx context.Context,
) ([]services.Note, error) {
	const query = `
		SELECT id, title, content, created_at
		FROM notes
		ORDER BY created_at DESC, id DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	result := make([]services.Note, 0)

	for rows.Next() {
		var note services.Note

		if err := rows.Scan(
			&note.ID,
			&note.Title,
			&note.Content,
			&note.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan note row: %w",
				err,
			)
		}

		result = append(result, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notes: %w", err)
	}

	return result, nil
}
