package main

import (
	"database/sql"
	"encoding/json"
	"errors"
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

//TV struct
type TV struct {
	ID           int64      `json:"id"`
	Brand        NullString `json:"brand"`
	Manufacturer string     `json:"manufacturer"`
	Model        string     `json:"model"`
	Year         int        `json:"year"`
}

//InPutTv struct
type InPutTv struct {
	Brand        NullString `json:"brand"`
	Manufacturer string     `json:"manufacturer"`
	Model        string     `json:"model"`
	Year         int        `json:"year"`
}

//GetTV get information about the TV by id
func GetTV(db *sql.DB, id int64) (*TV, error) {

	row := db.QueryRow("select * from tv where tv.id = $1", id)

	tv := TV{}
	err := row.Scan(&tv.ID, &tv.Brand, &tv.Manufacturer, &tv.Model, &tv.Year)

	if err == nil {
		return &tv, nil
	}
	return nil, err

}

//UpdateTV put data from tv table
func UpdateTV(db *sql.DB, id int64, inTv *InPutTv) (int64, error) {
	result, err := db.Exec("update tv set brand = $1, manufacturer = $2, model = $3, year = $4  where id = $5",
		inTv.Brand, inTv.Manufacturer, inTv.Model, inTv.Year, id)

	if err != nil {
		return -1, err
	}

	updateRows, err := result.RowsAffected()

	if err != nil || updateRows < 1 {
		return -1, err
	}
	return id, err
}

//AddTV insert data from tv table
func AddTV(db *sql.DB, id int64, inTv *InPutTv) (int64, error) {
	result, err := db.Exec("insert into tv (id, brand, manufacturer, model, year) values($1, $2, $3, $4, $5)",
		id, inTv.Brand, inTv.Manufacturer, inTv.Model, inTv.Year)

	if err != nil {
		return -1, err
	}

	addRows, err := result.RowsAffected()

	if err != nil || addRows < 1 {
		return -1, err
	}
	return id, err
}

//DelTV delete data from tv table
func DelTV(db *sql.DB, id int64) (int64, error) {
	result, err := db.Exec("delete from tv where tv.id = $1", id)
	if err != nil {
		return -1, err
	}
	remowedRows, err := result.RowsAffected()

	if err != nil || remowedRows < 1 {
		return -1, err
	}
	return id, err
}

//UpdtateReturns recalculation of available and sold goods
func UpdtateReturns(db *sql.DB, id int64, returns int64) (int64, error) {
	row := db.QueryRow("select sold_count, available from soldTv where soldTv.tv_id = $1", id)

	var availible int64
	var sold int64
	err := row.Scan(&sold, &availible)

	if err != nil {
		return -1, err
	}

	if returns > sold {
		return -1, errors.New("the quantity of returned goods is greater than the quantity sold")
	}

	availible += returns
	sold -= returns

	result, err := db.Exec("update soldTv set sold_count = $1, available = $2 where soldTv.tv_id = $3",
		sold, availible, id)

	if err != nil {
		return -1, err
	}

	updateRows, err := result.RowsAffected()

	if err != nil || updateRows < 1 {
		return -1, err
	}
	return availible, err
}
