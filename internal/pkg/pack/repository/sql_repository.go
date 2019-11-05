package repository

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2019_2_RabbitRoar/internal/pkg/models"
	"github.com/go-park-mail-ru/2019_2_RabbitRoar/internal/pkg/pack"
	"github.com/jackc/pgx/v4/pgxpool"
)

type sqlPackRepository struct {
	conn *pgxpool.Pool
}

func NewSqlPackRepository(conn *pgxpool.Pool) pack.Repository {
	return &sqlPackRepository{
		conn: conn,
	}
}

func (repo sqlPackRepository) GetByID(packID int) (*models.Pack, error) {
	row := repo.conn.QueryRow(
		context.Background(),
		"SELECT id, name, description, img, rating, author, private, tags"+
			"FROM svoyak.\"Pack\""+
			"WHERE id = $1::integer;",
		packID,
	)

	var pack models.Pack
	err := row.Scan(&pack.ID, &pack.Name, &pack.Description, &pack.Img, &pack.Rating, &pack.Author, &pack.Private, &pack.Tags)

	return &pack, err
}

func (repo sqlPackRepository) GetQuestions(pack models.Pack) (*[]models.Question, error) {
	rows, err := repo.conn.Query(
		context.Background(),
		"SELECT id, text, media, answer, rating, author, tags"+
			"FROM svoyak.\"Question\""+
			"WHERE id = ANY(SELECT Question_id"+
			"FROM svoyak.\"PackQuestion\""+
			"WHERE QuestionPack_id = $1::integer);",
		pack.ID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var questions []models.Question

	for rows.Next() {
		var question models.Question

		err := rows.Scan(&question.ID, &question.Text, &question.Media, &question.Answer, &question.Rating, &question.Author, &question.Tags)

		if err != nil {
			return nil, err
		}

		questions = append(questions, question)
	}

	return &questions, rows.Err()
}

func (repo sqlPackRepository) FetchOrderedByRating(desc bool, page, pageSize int) (*[]models.Pack, error) {
	var order string
	if desc {
		order = "DESC"
	} else {
		order = "ASC"
	}

	rows, err := repo.conn.Query(
		context.Background(),
		"SELECT id, name, description, img, rating, author, private, tags"+
			"FROM svoyak.\"Pack\""+
			"ORDER BY rating $1::text;",
		order,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var packs []models.Pack

	for rows.Next() {
		var pack models.Pack

		err := rows.Scan(&pack.ID, &pack.Name, &pack.Description, &pack.Img, &pack.Rating, &pack.Author, &pack.Private, &pack.Tags)

		if err != nil {
			return nil, err
		}

		packs = append(packs, pack)
	}

	return &packs, rows.Err()
}

func (repo sqlPackRepository) FetchByTags(tags string, page, pageSize int) (*[]models.Pack, error) {
	rows, err := repo.conn.Query(
		context.Background(),
		"SELECT id, name, description, img, rating, author, private, tags"+
			"FROM svoyak.\"Pack\""+
			"WHERE tags = $1::varchar"+
			"OFFSET $2::integer LIMIT $3::integer;",
		tags, (page * pageSize), pageSize,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var packs []models.Pack

	for rows.Next() {
		pack := models.Pack{
			Tags: tags,
		}

		err := rows.Scan(&pack.ID, &pack.Name, &pack.Description, &pack.Img, &pack.Rating, &pack.Author, &pack.Private)

		if err != nil {
			return nil, err
		}

		packs = append(packs, pack)
	}

	return &packs, rows.Err()
}

func (repo *sqlPackRepository) Create(pack models.Pack) (*models.Pack, error) {
	idRow := repo.conn.QueryRow(
		context.Background(),
		"INSERT INTO svoyak.\"Pack\" (id, name, description, img, rating, author, private, tags)"+
			"VALUES (DEFAULT, $1::varchar, $2::text, $3::varchar, $4::integer, $5::integer, $6::boolean, $7::varchar)"+
			"RETURNING id;",
		pack.Name, pack.Description, pack.Img, pack.Rating, pack.Author, pack.Private, pack.Tags,
	)

	err := idRow.Scan(&pack.ID)

	return &pack, err
}

func (repo *sqlPackRepository) Update(pack models.Pack) error {
	commandTag, err := repo.conn.Exec(
		context.Background(),
		"UPDATE svoyak.Pack"+
			"SET name = $1::varchar, description = $2::text, img = $3::varchar, rating = $4::integer, author = $5::integer, private = $6::boolean, tags = $7::varchar"+
			"WHERE id = $8::integer;",
		pack.Name, pack.Description, pack.Img, pack.Rating, pack.Author, pack.Private, pack.Tags, pack.ID,
	)

	if commandTag.RowsAffected() != 1 {
		return errors.New("Unable to update pack: No pack found")
	}

	return err
}

func (repo *sqlPackRepository) Delete(packID int) error {
	commandTag, err := repo.conn.Exec(
		context.Background(),
		"DELETE FROM svoyak.Pack"+
			"WHERE id = $1::integer;",
		packID,
	)

	if commandTag.RowsAffected() != 1 {
		return errors.New("Unable to delete pack: No pack found")
	}

	return err
}
