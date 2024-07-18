package account

import (
	"azflow-api/db"
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
)

func (m *Member) Signup() (int, error) {
	var id *int
	err := pgxscan.Get(context.Background(), db.Conn, &id, "INSERT INTO member(ext_id, email) VALUES($1, $2) RETURNING id", m.ExtId, m.Email)

	if err != nil {
		return 0, err
	}

	return *id, err
}
