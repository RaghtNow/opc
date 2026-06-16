package insight

import (
	"database/sql"
	"sync"

	domaininsight "github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/insight"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) ListSyncRecords(classID string) ([]domaininsight.SyncRecord, error) {
	rows, err := r.db.Query(`
		SELECT target, channel, status, published_at
		FROM insight_sync_records
		WHERE class_id = ?
		ORDER BY created_at DESC
		LIMIT 20
	`, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []domaininsight.SyncRecord{}
	for rows.Next() {
		var record domaininsight.SyncRecord
		if err := rows.Scan(&record.Target, &record.Channel, &record.Status, &record.Time); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (r *MySQLRepository) SaveSyncRecords(classID string, records []domaininsight.SyncRecord) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, record := range records {
		if _, err := tx.Exec(`
			INSERT INTO insight_sync_records (class_id, target, channel, status, published_at)
			VALUES (?, ?, ?, ?, ?)
		`, classID, record.Target, record.Channel, record.Status, record.Time); err != nil {
			return err
		}
	}
	return tx.Commit()
}

type MemoryRepository struct {
	mu      sync.RWMutex
	records map[string][]domaininsight.SyncRecord
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{records: map[string][]domaininsight.SyncRecord{}}
}

func (r *MemoryRepository) ListSyncRecords(classID string) ([]domaininsight.SyncRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domaininsight.SyncRecord{}, r.records[classID]...), nil
}

func (r *MemoryRepository) SaveSyncRecords(classID string, records []domaininsight.SyncRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records[classID] = append(records, r.records[classID]...)
	return nil
}
