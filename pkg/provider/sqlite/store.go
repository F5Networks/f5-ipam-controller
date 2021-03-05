package sqlite

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/F5Networks/f5-ipam-controller/pkg/vlogger"
)

type DBStore struct {
	db *sql.DB
}

const (
	ALLOCATED = 0
	AVAILABLE = 1
)

func NewStore() *DBStore {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		log.Errorf("[STORE] Unable to Initialise DB, %v", err)
		return nil
	}

	err = db.Ping()
	if err != nil {
		log.Errorf("[STORE] Unable to Establish Connection to DB, %v", err)
		return nil
	}

	store := &DBStore{db: db}
	if !store.CreateTables() {
		return nil
	}

	return store
}

func (store *DBStore) CreateTables() bool {
	createIPAddressTableSQL := `CREATE TABLE ipaddress_range (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"ipaddress" TEXT,
		"status" INT,
		"cidr" TEXT	
	  );`

	statement, _ := store.db.Prepare(createIPAddressTableSQL)

	_, err := statement.Exec()
	if err != nil {
		log.Errorf("[STORE] Unable to Create Table 'ipaddress_range' in Database")
		return false
	}
	createARecodsTableSQL := `CREATE TABLE a_records (
		"ipaddress" TEXT PRIMARY_KEY,
		"hostname" TEXT	
	  );`

	statement, _ = store.db.Prepare(createARecodsTableSQL)

	_, err = statement.Exec()
	if err != nil {
		log.Errorf("[STORE] Unable to Create  Table 'a_records' in Database")
		return false
	}
	return true
}

func (store *DBStore) InsertIP(ips []string, cidr string) {
	for _, j := range ips {
		insertIPSQL := `INSERT INTO ipaddress_range(ipaddress, status, cidr) VALUES (?, ?, ?)`

		statement, _ := store.db.Prepare(insertIPSQL)

		_, err := statement.Exec(j, AVAILABLE, cidr)
		if err != nil {
			log.Error("[STORE] Unable to Insert row in Table 'ipaddress_range'")
		}
	}
}

func (store *DBStore) DisplayIPRecords() {

	row, err := store.db.Query("SELECT * FROM ipaddress_range ORDER BY id")
	if err != nil {
		log.Debugf(" ", err)
	}
	columns, err := row.Columns()
	if err != nil {
		log.Debugf(" err : ", err)
	}
	log.Debugf("[STORE] Column names: %v", columns)
	defer row.Close()
	for row.Next() {
		var id int
		var ipaddress string
		var status int
		var cidr string
		row.Scan(&id, &ipaddress, &status, &cidr)
		log.Debugf("[STORE] ipaddress_range: %v\t %v\t%v\t%v", id, ipaddress, status, cidr)
	}
}

func (store *DBStore) AllocateIP(cidr string) string {
	var ipaddress string
	var id int

	queryString := fmt.Sprintf(
		"SELECT ipaddress,id FROM ipaddress_range where status=%d AND cidr=\"%s\" order by id ASC limit 1",
		AVAILABLE,
		cidr,
	)
	err := store.db.QueryRow(queryString).Scan(&ipaddress, &id)
	if err != nil {
		log.Infof("[STORE] No Available IP Addresses to Allocate: %v", err)
		return ""
	}

	allocateIPSql := fmt.Sprintf("UPDATE ipaddress_range set status = %d where id = ?", ALLOCATED)
	statement, _ := store.db.Prepare(allocateIPSql)

	_, err = statement.Exec(id)
	if err != nil {
		log.Errorf("[STORE] Unable to update row in Table 'ipaddress_range': %v", err)
	}
	return ipaddress
}

func (store *DBStore) MarkIPAsAllocated(cidr, ipAddr string) bool {
	var id int

	queryString := fmt.Sprintf(
		"SELECT id FROM ipaddress_range where status=%d AND cidr=\"%s\" AND ipaddress=\"%s\" order by id ASC limit 1",
		AVAILABLE,
		cidr,
		ipAddr,
	)
	err := store.db.QueryRow(queryString).Scan(&id)
	if err != nil {
		log.Infof("[STORE] No Available IP Addresses to Allocate: %v", err)
		return false
	}

	allocateIPSql := fmt.Sprintf("UPDATE ipaddress_range set status = %d where id = ?", ALLOCATED)
	statement, _ := store.db.Prepare(allocateIPSql)

	_, err = statement.Exec(id)
	if err != nil {
		log.Errorf("[STORE] Unable to update row in Table 'ipaddress_range': %v", err)
		return false
	}
	return true
}

func (store *DBStore) GetIPAddress(hostname string) string {
	var ipaddress string

	queryString := fmt.Sprintf(
		"SELECT ipaddress FROM a_records where hostname=\"%s\" order by ipaddress ASC limit 1",
		hostname,
	)
	err := store.db.QueryRow(queryString).Scan(&ipaddress)
	if err != nil {
		log.Infof("[STORE] No A record with Host: %v", hostname)
		return ""
	}
	return ipaddress
}

func (store *DBStore) ReleaseIP(ip string) {
	unallocateIPSql := fmt.Sprintf("UPDATE ipaddress_range set status = %d where ipaddress = ?", AVAILABLE)
	statement, _ := store.db.Prepare(unallocateIPSql)

	_, err := statement.Exec(ip)
	if err != nil {
		log.Errorf("[STORE] Unable to update row in Table 'ipaddress_range': %v", err)
	}
}

func (store *DBStore) CreateARecord(hostname, ipAddr string) bool {
	insertARecordSQL := `INSERT INTO a_records(ipaddress, hostname) VALUES (?, ?)`

	statement, _ := store.db.Prepare(insertARecordSQL)

	_, err := statement.Exec(ipAddr, hostname)
	if err != nil {
		log.Error("[STORE] Unable to Insert row in Table 'a_records'")
		return false
	}
	return true
}

func (store *DBStore) DeleteARecord(hostname, ipAddr string) bool {
	deleteARecord := "DELETE FROM a_records WHERE ipaddress=? AND hostname=?"

	statement, _ := store.db.Prepare(deleteARecord)

	_, err := statement.Exec(ipAddr, hostname)
	if err != nil {
		log.Error("[STORE] Unable to Delete row from Table 'a_records'")
		return false
	}
	return true
}
