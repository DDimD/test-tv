package main

import (
	"database/sql"
	"encoding/json"
)

// NullString is an alias for sql.NullString data types
type NullString struct {
	sql.NullString
}

// MarshalJSON for NullString
func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

// UnmarshalJSON for NullString
func (ns *NullString) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &ns.String)
	ns.Valid = (err == nil)
	return err
}

type TV struct {
	ID           int64      `json:"id"`
	Brand        NullString `json:"brand"`
	Manufacturer string     `json:"manufacturer"`
	Model        string     `json:"model"`
	Year         int        `json:"year"`
}

type TvID struct {
	ID int64 `json:"id"`
}

//GetTV get information about the TV by id
func GetTV(db *sql.DB, id int64) (*TV, error) {

	row := db.QueryRow("select * from tv where tv.id = $1", id) //todo проверить, что выдаётся в неправильном запросе

	tv := TV{}
	err := row.Scan(&tv.ID, &tv.Brand, &tv.Manufacturer, &tv.Model, &tv.Year)

	if err == nil {
		return &tv, nil
	}
	return nil, err

}

func GetAllTV() {}

func UpdateTV() {}

func AddTV() {}

func DelTV(db *sql.DB, id int64) (int64, error) {
	_, err := db.Exec("delete from tv where tv.id = $1", id)
	if err != nil {
		return -1, err
	}
	return id, err
}
