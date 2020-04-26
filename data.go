package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"strings"
)

// Database properties
const (
	driver     = "postgres"
	dbhost     = "ec2-107-21-98-89.compute-1.amazonaws.com"
	dbuser     = "gydhbcepmvfojy"
	dbName     = "dccpn9636r8od"
	dbpassword = "4ce46cf1c6eabcd1d325dcd0fd31bddf3d573cdc6393f4e5077eead7fa3f53c8"
	sslmode    = "require"
)

type stmtConfig struct {
	stmt *sql.Stmt
	q    string
}

// Statements
const (
	GET               = "get"
	INSERT            = "insert"
	UPDATE            = "update"
	DELETE            = "delete"
	LIST              = "list"
	GET_MATCH_RAWDATA = "getMatchRawData"
)

type Data struct {
	Db    *sql.DB
	Stmts map[string]*stmtConfig
}

var data Data

func InitDb() error {
	var err error
	data.Db, err = sql.Open(driver, fmt.Sprintf(
		"host=%s user=%s dbname=%s password=%s sslmode=%s",
		dbhost, dbuser, dbName, dbpassword, sslmode))
	if err != nil {
		return err
	}
	data.Stmts = map[string]*stmtConfig{
		LIST: {q: "select id, municipality, name, address, department, latitude, longitude, map_url, phone," +
			" open_hour, close_hour from \"store\";"},
		GET: {q: "select municipality, name, address, department, latitude, longitude, map_url, phone," +
			" open_hour, close_hour from \"store\" where id=$1;"},
		GET_MATCH_RAWDATA: {q: "select  municipality, name, address, department, latitude, longitude, map_url, phone," +
			" open_hour, close_hour from \"store\" where raw_data like '%' || $1 || '%';"},
		INSERT: {q: "Insert into \"store\" ( municipality, name, address, department, latitude, longitude, map_url," +
			" phone, open_hour, close_hour, raw_data) values ( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11 );"},
		UPDATE: {q: "update \"store\" set municipality=$2, name=$3, address=$4, department=$5, latitude=$6, longitude=$7," +
			" map_url=$8, phone=$9, open_hour=$10, close_hour=$11 where id=$1;"},
		DELETE: {q: "delete from \"store\" where id=$1"},
	}
	for k, v := range data.Stmts {
		data.Stmts[k].stmt, _ = data.Db.Prepare(v.q)
	}
	return nil
}

func (d Data) Insert(s Store) (int64, error) {
	insertUser := data.Stmts[INSERT].stmt
	rawData := getRawData(s)
	err := insertUser.QueryRow(s.Municipality, s.Name, s.Address, s.Department, s.Geolocation.Latitude, s.Geolocation.Longitude,
		s.MapUrl, s.Phone, s.Open, s.Close, rawData).Scan(&s.ID)
	if err != nil {
		return 0, err
	}

	return s.ID, nil
}

func (d Data) Get(id int64) (*Store, error) {
	getUser := d.Stmts[GET].stmt
	store := Store{}
	err := getUser.QueryRow(id).
		Scan(&store.Municipality, &store.Name, &store.Address, &store.Department,
			&store.Geolocation.Latitude, &store.Geolocation.Longitude, &store.MapUrl, &store.Phone,
			&store.Open, &store.Close)
	if err != nil {
		return nil, err
	}
	store.ID = id
	return &store, nil
}

func (d Data) List() ([]Store, error) {
	listUser := d.Stmts[LIST].stmt
	rows, err := listUser.Query()
	if err != nil {
		return nil, err
	}

	var storeList = make([]Store, 0)
	for rows.Next() {
		var store = Store{}
		if err := rows.Scan(&store.ID, &store.Municipality, &store.Name, &store.Address, &store.Department,
			&store.Geolocation.Latitude, &store.Geolocation.Longitude, &store.MapUrl, &store.Phone,
			&store.Open, &store.Close); err != nil {
			return nil, err
		}
		storeList = append(storeList, store)
	}
	return storeList, nil
}

func (d Data) Delete(id int64) error {
	delUser := d.Stmts[DELETE].stmt
	_, err := delUser.Exec(id)
	return err
}

func (d Data) Update(id int64, new Store) error {
	updUser := d.Stmts[UPDATE].stmt
	_, err := updUser.Exec(id, new.Municipality, new.Name, new.Address, new.Department,
		new.Geolocation.Latitude, new.Geolocation.Longitude, new.MapUrl, new.Phone,
		new.Open, new.Close)

	return err
}

func (d Data) GetWhenMatchWithRawData(value string) ([]Store, error) {
	getUser := d.Stmts[GET_MATCH_RAWDATA].stmt
	var storeList = make([]Store, 0)
	rowList, err := getUser.Query(value)
	if err != nil {
		return nil, err
	}

	for rowList.Next() {
		var store = Store{}
		if err := rowList.Scan(&store.Municipality, &store.Name, &store.Address, &store.Department,
			&store.Geolocation.Latitude, &store.Geolocation.Longitude, &store.MapUrl, &store.Phone,
			&store.Open, &store.Close); err != nil {
			return nil, err
		}
		storeList = append(storeList, store)
	}

	return storeList, nil
}

func (d *Data) Close() error {
	for s := range d.Stmts {
		err := d.Stmts[s].stmt.Close()
		if err != nil {
			return err
		}
	}

	return d.Db.Close()
}

func getRawData(s Store) string {
	slice := []string{
		s.Name,
		s.Department,
		s.Municipality,
	}
	for i := range slice {
		slice[i] = strings.Trim(strings.ToLower(slice[i]), " \n\t\f\r!?#$%&'\"()*+,-./:;<=>@[\\^_`{|}~]")
	}
	return strings.Join(slice, ".")
}
