package shorten

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type PGStore struct {
	db *sql.DB
}

func NewPGStore(db *sql.DB) *PGStore {
	return &PGStore{db: db}
}

func (store *PGStore) Save(link ShortLink) error {
	_, err := store.db.Exec(`
	INSERT INTO links (short_id, original_url, hits, created_at)
	VALUES ($1, $2, $3, NOW())
	`, link.ID, link.URL, link.Hits)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return ErrDuplicateID
			}
		}
		return err
	}

	return nil
}

func (store *PGStore) Get(id string) (ShortLink, error) {
	var link ShortLink

	err := store.db.QueryRow(`
	SELECT short_id, original_url, hits, created_at 
	FROM links 
	WHERE short_id = $1
	`, id).Scan(
		&link.ID,
		&link.URL,
		&link.Hits,
		&link.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ShortLink{}, ErrNotFound
		}
		return ShortLink{}, err
	}

	return link, nil
}

func (store *PGStore) IncrementHits(id string) error {
	result, err := store.db.Exec(`
	UPDATE links
	SET hits = hits + 1
	WHERE short_id = $1
	`, id)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
