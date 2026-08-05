// Package pgutil converts between pgx's nullable wire types and the plain
// Go types (pointers, uuid.UUID, decimal.Decimal) used across the domain
// packages, so sqlc-generated code stays isolated to the db package.
package pgutil

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

func UUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func NullUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return UUID(*id)
}

func ToUUID(u pgtype.UUID) uuid.UUID {
	return uuid.UUID(u.Bytes)
}

func ToNullUUID(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := ToUUID(u)
	return &id
}

func Date(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func ToTime(d pgtype.Date) time.Time {
	return d.Time
}

func Timestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func Int2(v *int16) pgtype.Int2 {
	if v == nil {
		return pgtype.Int2{}
	}
	return pgtype.Int2{Int16: *v, Valid: true}
}

func FromInt2(v pgtype.Int2) *int16 {
	if !v.Valid {
		return nil
	}
	n := v.Int16
	return &n
}

func NullDecimal(v *decimal.Decimal) decimal.NullDecimal {
	if v == nil {
		return decimal.NullDecimal{}
	}
	return decimal.NullDecimal{Decimal: *v, Valid: true}
}

func FromNullDecimal(v decimal.NullDecimal) *decimal.Decimal {
	if !v.Valid {
		return nil
	}
	d := v.Decimal
	return &d
}

func Text(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *v, Valid: true}
}

func FromText(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}
