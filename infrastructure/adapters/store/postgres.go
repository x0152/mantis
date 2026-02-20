package store

import (
	"context"

	"github.com/uptrace/bun"

	"mantis/core/types"
)

type Postgres[ID comparable, Entity any, Row any] struct {
	db      *bun.DB
	getID   func(Entity) ID
	toRow   func(Entity) Row
	fromRow func(Row) Entity
}

func NewPostgres[ID comparable, Entity any, Row any](
	db *bun.DB,
	getID func(Entity) ID,
	toRow func(Entity) Row,
	fromRow func(Row) Entity,
) *Postgres[ID, Entity, Row] {
	return &Postgres[ID, Entity, Row]{db: db, getID: getID, toRow: toRow, fromRow: fromRow}
}

func (s *Postgres[ID, Entity, Row]) Create(ctx context.Context, items []Entity) ([]Entity, error) {
	rows := toRows(items, s.toRow)
	_, err := s.db.NewInsert().Model(&rows).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return fromRows(rows, s.fromRow), nil
}

func (s *Postgres[ID, Entity, Row]) Get(ctx context.Context, ids []ID) (map[ID]Entity, error) {
	var rows []Row
	err := s.db.NewSelect().Model(&rows).Where("id IN (?)", bun.In(ids)).Scan(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[ID]Entity, len(rows))
	for _, row := range rows {
		e := s.fromRow(row)
		result[s.getID(e)] = e
	}
	return result, nil
}

func (s *Postgres[ID, Entity, Row]) List(ctx context.Context, query types.ListQuery) ([]Entity, error) {
	var rows []Row
	q := s.db.NewSelect().Model(&rows)
	for field, value := range query.Filter {
		q = q.Where("? = ?", bun.Ident(field), value)
	}
	for field, value := range query.FilterNot {
		q = q.Where("? != ?", bun.Ident(field), value)
	}
	for _, item := range query.Sort {
		q = q.OrderExpr("? ?", bun.Ident(item.Field), bun.Safe(string(item.Dir)))
	}
	if query.Page.Limit > 0 {
		q = q.Limit(query.Page.Limit)
	}
	if query.Page.Offset > 0 {
		q = q.Offset(query.Page.Offset)
	}
	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}
	return fromRows(rows, s.fromRow), nil
}

func (s *Postgres[ID, Entity, Row]) Update(ctx context.Context, items []Entity) ([]Entity, error) {
	result := make([]Entity, len(items))
	for i, item := range items {
		row := s.toRow(item)
		_, err := s.db.NewUpdate().Model(&row).WherePK().Exec(ctx)
		if err != nil {
			return nil, err
		}
		result[i] = s.fromRow(row)
	}
	return result, nil
}

func (s *Postgres[ID, Entity, Row]) Delete(ctx context.Context, ids []ID) error {
	var row Row
	_, err := s.db.NewDelete().Model(&row).Where("id IN (?)", bun.In(ids)).Exec(ctx)
	return err
}

func toRows[Entity any, Row any](items []Entity, fn func(Entity) Row) []Row {
	rows := make([]Row, len(items))
	for i, item := range items {
		rows[i] = fn(item)
	}
	return rows
}

func fromRows[Row any, Entity any](rows []Row, fn func(Row) Entity) []Entity {
	items := make([]Entity, len(rows))
	for i, row := range rows {
		items[i] = fn(row)
	}
	return items
}
