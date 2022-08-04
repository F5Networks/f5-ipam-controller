/*-
 * Copyright (c) 2021, F5 Networks, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sqlite

import (
	"database/sql"
	"fmt"
	"github.com/F5Networks/f5-ipam-controller/pkg/utils"
	"os"

	"github.com/F5Networks/f5-ipam-controller/pkg/provider"
	log "github.com/F5Networks/f5-ipam-controller/pkg/vlogger"
	_ "github.com/mattn/go-sqlite3"
)

type DBStore struct {
	db *sql.DB
}

const (
	ALLOCATED = 0
	AVAILABLE = 1

	dbFileName = "/app/ipamdb/cis_ipam.sqlite3"

	ReferenceLength = 16
)

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			log.Debugf("[STORE] Creating IPAM DB file in mount path")
			file, err := os.OpenFile(name, os.O_CREATE, 0660)
			if err != nil {
				log.Errorf("[STORE] Unable to create IPAM DB file: %v", err)
			}
			defer func() {
				if err == nil {
					_ = file.Close()
				}
			}()
			return err == nil
		} else if os.IsPermission(err) {
			log.Errorf("[STORE] Unable to read IPAM DB file due to permission issue: %v", err)
		} else {
			log.Errorf("[STORE] Unable to verify IPAM DB file: %v", err)
		}
		return false
	}
	log.Debugf("[STORE] Using IPAM DB file from mount path")
	return true
}

func NewStore() provider.StoreProvider {
	if !fileExists(dbFileName) {
		return nil
	}
	dsn := "file:" + dbFileName + "?cache=shared&mode=rw"
	db, err := sql.Open("sqlite3", dsn)
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

func (store *DBStore) executeStatement(stmt string, params ...interface{}) error {
	statement, err := store.db.Prepare(stmt)
	if err != nil {
		return err
	}
	_, err = statement.Exec(params...)
	if err != nil {
		return err
	}
	return nil
}

func (store *DBStore) CreateTables() bool {
	// Create `label_map` table
	err := store.executeStatement(
		`CREATE TABLE IF NOT EXISTS label_map (
		"ipam_label" TEXT PRIMARY_KEY,
		"range" TEXT
	  );`,
	)
	if err != nil {
		log.Errorf("[STORE] Unable to Create  Table 'label_map' in Database")
		return false
	}

	err = store.executeStatement(
		`CREATE TABLE IF NOT EXISTS ipaddress_range (
		"ipaddress" TEXT PRIMARY KEY,
		"status" INT,
		"ipam_label" TEXT,
        "reference" TEXT UNIQUE
	  );`,
	)
	if err != nil {
		log.Errorf("[STORE] Unable to Create Table 'ipaddress_range' in Database. Error %v", err)
		return false
	}

	err = store.executeStatement(
		`CREATE TABLE IF NOT EXISTS a_records (
		"ipaddress" TEXT PRIMARY_KEY,
		"hostname" TEXT
	  );`,
	)
	if err != nil {
		log.Errorf("[STORE] Unable to Create Table 'a_records' in Database: %v", err)
		return false
	}

	return true
}

func (store *DBStore) InsertIPs(ips []string, ipamLabel string) {
	for _, ip := range ips {
		err := store.executeStatement(
			`INSERT INTO ipaddress_range(ipaddress, status, ipam_label, reference) VALUES (?, ?, ?, ?)`,
			ip, AVAILABLE, ipamLabel, utils.RandomString(ReferenceLength))
		if err != nil {
			log.Errorf("[STORE] Unable to Insert row in Table 'ipaddress_range': %v", err)
		}
	}
}

func (store *DBStore) DisplayIPRecords() {

	row, err := store.db.Query("SELECT * FROM ipaddress_range")
	if err != nil {
		log.Debugf("%v ", err)
	}
	columns, err := row.Columns()
	if err != nil {
		log.Debugf(" err : %v", err)
	}
	log.Debugf("[STORE] %v", columns)
	defer row.Close()
	for row.Next() {
		var ipaddress string
		var status int
		var ipamLabel string
		var ref string
		row.Scan(&ipaddress, &status, &ipamLabel, &ref)
		log.Debugf("[STORE] %v %v %v %v", ipaddress, status, ipamLabel, ref)
	}
}

func (store *DBStore) AllocateIP(ipamLabel, reference string) string {
	var ipaddress string

	queryString := fmt.Sprintf(
		"SELECT ipaddress FROM ipaddress_range where status=%d AND ipam_label=\"%s\" order by ipaddress ASC limit 1",
		AVAILABLE,
		ipamLabel,
	)
	err := store.db.QueryRow(queryString).Scan(&ipaddress)
	if err != nil {
		log.Infof("[STORE] No Available IP Addresses to Allocate: %v", err)
		return ""
	}

	allocateIPSql := fmt.Sprintf("UPDATE ipaddress_range set status=%d, reference=\"%s\" WHERE ipaddress=?",
		ALLOCATED,
		reference,
	)
	err = store.executeStatement(allocateIPSql, ipaddress)
	if err != nil {
		log.Errorf("[STORE] Unable to update row in Table 'ipaddress_range': %v", err)
	}
	return ipaddress
}

func (store *DBStore) GetIPAddressFromARecord(ipamLabel, hostname string) string {
	var ipaddress string
	var status int

	queryString := fmt.Sprintf(
		"SELECT ipaddress FROM a_records where hostname=\"%s\" order by ipaddress ASC limit 1",
		hostname,
	)
	err := store.db.QueryRow(queryString).Scan(&ipaddress)
	if err != nil {
		return ""
	}

	queryString = fmt.Sprintf("SELECT status FROM ipaddress_range where ipaddress=\"%s\" AND ipam_label=\"%s\" order by ipaddress ASC LIMIT 1",
		ipaddress,
		ipamLabel,
	)
	err = store.db.QueryRow(queryString).Scan(&status)
	if err != nil {
		log.Errorf("Unable to fetch IPAddress from 'A' record %s with error %v", err)
		return ""
	} else if status == AVAILABLE {
		return ""
	}

	return ipaddress
}

func (store *DBStore) GetIPAddressFromReference(ipamLabel, reference string) string {
	var ipaddress string
	var status int

	queryString := fmt.Sprintf("SELECT ipaddress, status FROM ipaddress_range where reference=\"%s\" AND ipam_label=\"%s\" order by ipaddress ASC limit 1",
		reference,
		ipamLabel,
	)
	err := store.db.QueryRow(queryString).Scan(&ipaddress, &status)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Errorf("Unable to fetch IPAddress for host %s with error %v", reference, err)
		}
		return ""
	}
	if status == AVAILABLE {
		return ""
	}

	return ipaddress
}

func (store *DBStore) ReleaseIP(ip string) {
	deallocateIPSql := fmt.Sprintf("UPDATE ipaddress_range set status=%d, reference=\"%s\" where ipaddress=?",
		AVAILABLE,
		utils.RandomString(ReferenceLength),
	)

	err := store.executeStatement(deallocateIPSql, ip)
	if err != nil {
		log.Errorf("[STORE] Unable to update row in Table 'ipaddress_range': %v", err)
	}
}

func (store *DBStore) CreateARecord(hostname, ipAddr string) bool {
	insertARecordSQL := `INSERT INTO a_records(ipaddress, hostname) VALUES (?, ?)`

	err := store.executeStatement(insertARecordSQL, ipAddr, hostname)
	if err != nil {
		log.Errorf("[STORE] Unable to Insert row in Table 'a_records': %v", err)
		return false
	}
	return true
}

func (store *DBStore) DeleteARecord(hostname, ipAddr string) bool {
	deleteARecord := "DELETE FROM a_records WHERE ipaddress=? AND hostname=?"

	err := store.executeStatement(deleteARecord, ipAddr, hostname)
	if err != nil {
		log.Errorf("[STORE] Unable to Delete row from Table 'a_records': %v", err)
		return false
	}
	return true
}

func (store *DBStore) GetLabelMap() map[string]string {
	row, err := store.db.Query("SELECT * FROM label_map")
	if err != nil {
		log.Debugf("%v ", err)
	}
	rangeMap := make(map[string]string)
	defer row.Close()
	for row.Next() {
		var ipamLabel string
		var ipamRange string
		row.Scan(&ipamLabel, &ipamRange)
		rangeMap[ipamLabel] = ipamRange
	}

	return rangeMap
}

func (store *DBStore) AddLabel(label, ipRange string) bool {
	err := store.executeStatement(
		`INSERT INTO label_map(ipam_label, range) VALUES (?, ?)`,
		label,
		ipRange,
	)
	if err != nil {
		log.Errorf("[STORE] Unable to Insert row in Table 'label_map': %v", err)
		return false
	}
	return true
}

func (store *DBStore) RemoveLabel(label string) bool {
	err := store.executeStatement(
		"DELETE FROM label_map WHERE ipam_label=?",
		label,
	)
	if err != nil {
		log.Errorf("[STORE] Unable to Delete label: %v: %v", label, err)
		return false
	}
	return true
}

// CleanUpLabel performs DELETE CASCADE of associated rows in all tables
// TODO: Should be replaced by standard DB cascade deletion
func (store *DBStore) CleanUpLabel(label string) {
	row, err := store.db.Query(fmt.Sprintf("SELECT ipaddress FROM ipaddress_range WHERE ipam_label = \"%s\"", label))
	if err != nil {
		log.Debugf("%v", err)
	}

	defer row.Close()
	for row.Next() {
		var ipAddr string

		err = row.Scan(&ipAddr)
		if err != nil {
			continue
		}
		_ = store.executeStatement(
			"DELETE FROM a_records WHERE ipaddress=?",
			ipAddr,
		)
	}

	_ = store.executeStatement(
		"DELETE FROM ipaddress_range WHERE ipam_label=?",
		label,
	)

	_ = store.RemoveLabel(label)
}
