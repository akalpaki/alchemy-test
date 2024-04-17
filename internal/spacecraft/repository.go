package spacecraft

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
)

var (
	errNotFound = errors.New("entry not found")
)

type Repository struct {
	conn *sql.DB
}

func NewRepository(conn *sql.DB) *Repository {
	return &Repository{
		conn: conn,
	}
}

// Create an entry for a new spaceship.
func (r *Repository) Create(ctx context.Context, craft SpacecraftRequest) error {
	qSpaceship := `INSERT INTO spacecrafts (name, class, status, image, crew, value) VALUES (?, ?, ?, ?, ?, ?)`
	qArmaments := `INSERT INTO armaments (title, quanity) VALUES (?, ?)`

	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("spacecraft_repo: begin tx: %w", err)
	}

	insertArmament, err := tx.Prepare(qArmaments)
	if err != nil {
		return fmt.Errorf("spacecraft_repo: prepare stmt: %w", err)
	}

	_, err = tx.ExecContext(ctx, qSpaceship, craft.Name, craft.Class, craft.Status, craft.Image, craft.Crew, craft.Value)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("spacecraft_repo: insert spacecraft: %w", err)
	}

	for _, armament := range craft.Armament {
		_, err := insertArmament.ExecContext(ctx, armament.Title, armament.Quantity)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("spacecraft_repo: insert armament: %w", err)
		}
	}

	return nil
}

func (r *Repository) Update(ctx context.Context, id int, craft SpacecraftRequest) error {
	q := "UPDATE spacecrafts SET name = ?, class = ?, status = ?, image = ?, crew = ?, value = ? WHERE id = ?"

	res, err := r.conn.ExecContext(ctx, q, craft.Name, craft.Class, craft.Status, craft.Image, craft.Crew, craft.Value)
	if err != nil {
		return fmt.Errorf("spacecraft_repo: update spacecraft: %w", err)
	}
	rAff, _ := res.RowsAffected()
	if rAff == 0 {
		return errNotFound
	}
	return nil
}
func (r *Repository) Delete(ctx context.Context, id int) error {
	q := "DELETE FROM spacecrafts WHERE id = ?"

	res, err := r.conn.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("spacecraft_repo: delete spacecraft: %w", err)
	}
	rAff, _ := res.RowsAffected()
	if rAff == 0 {
		return errNotFound
	}
	return nil
}
func (r *Repository) GetByID(ctx context.Context, id int) (Spacecraft, error) {
	qSpaceship := "SELECT (id, name, class, status, image, crew, value) FROM spacecrafts WHERE id = ?"
	qArmaments := "SELECT (id, craft_id, title, quantity) FROM armaments WHERE craft_id = ?"

	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return Spacecraft{}, fmt.Errorf("spacecraft_repo: begin tx: %w", err)
	}

	var craft Spacecraft
	row := tx.QueryRowContext(ctx, qSpaceship, id)
	if err := row.Scan(&craft.ID, &craft.Name, &craft.Class, &craft.Status, &craft.Image, &craft.Crew, &craft.Value); err != nil {
		tx.Rollback()
		return Spacecraft{}, fmt.Errorf("spacecraft_repo: retrieve spacecraft: %w", err)
	}

	armRows, err := tx.QueryContext(ctx, qArmaments, craft.ID)
	if err != nil {
		tx.Rollback()
		return Spacecraft{}, fmt.Errorf("spacecraft_repo: retrieve armaments: %w", err)
	}

	armaments := make([]Armament, 0)
	for armRows.Next() {
		var armament Armament
		if err := armRows.Scan(&armament.ID, &armament.CraftID, &armament.Title, &armament.Quantity); err != nil {
			tx.Rollback()
			return Spacecraft{}, fmt.Errorf("spacecraft_repo: retrieve armaments: %w", err)
		}
		armaments = append(armaments, armament)
	}
	craft.Armament = armaments

	return craft, nil
}

func (r *Repository) Get(ctx context.Context, filters url.Values) ([]Spacecraft, error) {
	q := `
	SELECT (id, name, class, status , image, crew, value) FROM spacecrafts
	WHERE (LOWER(name) = LOWER($1)) OR $1 = ''
	AND (LOWER(class) = LOWER($2)) OR $2 = ''
	AND (LOWER(status) = LOWER($3)) OR $3 = ''
	ORDER BY id
	`
	name := filters.Get("name")
	class := filters.Get("class")
	status := filters.Get("status")

	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("spacecraft_repo: begin tx: %w", err)
	}

	selectArmaments, err := tx.Prepare("SELECT (id, craft_id, title, quantity) FROM armaments WHERE craft_id = ?")
	if err != nil {
		return nil, fmt.Errorf("spacecraft_repo: preparing stmt: %w", err)
	}

	rows, err := tx.QueryContext(ctx, q, name, class, status)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("spacecraft_repo: retrieve spacecrafts: %w", err)
	}

	spacecrafts := make([]Spacecraft, 0)
	for rows.Next() {
		var spacecraft Spacecraft
		if err := rows.Scan(&spacecraft.ID, &spacecraft.Name, &spacecraft.Class, &spacecraft.Status, &spacecraft.Image, &spacecraft.Crew, &spacecraft.Value); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("spacecraft_repo: retrieve spacecrafts: %w", err)
		}
		spacecrafts = append(spacecrafts, spacecraft)
	}

	// I know this is bad but I'm running out of time.
	for _, spacecraft := range spacecrafts {
		armaments := make([]Armament, 0)
		armRows, err := selectArmaments.Query(spacecraft.ID)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("spacecraft_repo: retrieve armaments: %w", err)
		}
		for armRows.Next() {
			var armament Armament
			if err := armRows.Scan(&armament.ID, &armament.CraftID, &armament.Title, &armament.Quantity); err != nil {
				return nil, fmt.Errorf("spacecraft_repo: retrieve armaments: %w", err)
			}
			armaments = append(armaments, armament)
		}
		copy(spacecraft.Armament, armaments)
		clear(armaments)
	}

	return spacecrafts, nil
}
