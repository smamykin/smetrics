package storage

import (
	"database/sql"
	"github.com/smamykin/smetrics/internal/server/handlers"
)

func NewDbStorage(db *sql.DB) (*DbStorage, error) {
	result := &DbStorage{db: db}
	err := result.init()
	return result, err
}

type DbStorage struct {
	db        *sql.DB
	observers []Observer
}

func (d *DbStorage) init() error {
	tableExistsSql := "SELECT EXISTS ( SELECT FROM pg_tables WHERE tablename  = 'metric');"
	var isTableExists bool
	err := d.db.QueryRow(tableExistsSql).Scan(&isTableExists)
	if err != nil {
		return err
	}
	if isTableExists {
		return nil
	}

	_, err = d.db.Exec(`
		CREATE TABLE metric (id SERIAL, name varchar(255) NOT NULL, type varchar(255) NOT NULL, value DOUBLE PRECISION, delta INT, PRIMARY KEY(id));
		CREATE UNIQUE INDEX name_type_unique ON metric (name, type);
	`)

	if err != nil {
		return err
	}

	return nil
}

func (d *DbStorage) UpsertGauge(metric handlers.GaugeMetric) error {
	upsertSql := `
		INSERT INTO metric (name, type, value) 
		VALUES ($1, $2, $3)
		ON CONFLICT (name, type) DO UPDATE 
		    SET value = EXCLUDED.value
	`
	_, err := d.db.Exec(upsertSql, metric.Name, handlers.MetricTypeGauge, metric.Value)

	if err != nil {
		return err
	}

	return d.notifyObservers(AfterUpsertEvent{
		Event{metric},
	})
}

func (d *DbStorage) UpsertCounter(metric handlers.CounterMetric) error {
	upsertSql := `
		INSERT INTO metric (name, type, delta) 
		VALUES ($1, $2, $3)
		ON CONFLICT (name, type) DO UPDATE 
		    SET delta = EXCLUDED.delta
	`
	_, err := d.db.Exec(upsertSql, metric.Name, handlers.MetricTypeCounter, metric.Value)
	if err != nil {
		return err
	}

	return d.notifyObservers(AfterUpsertEvent{
		Event{metric},
	})
}

func (d *DbStorage) GetGauge(name string) (float64, error) {
	getOneSql := `
		SELECT value
		FROM metric
		WHERE type = $1 AND name = $2
	`
	row := d.db.QueryRow(getOneSql, handlers.MetricTypeGauge, name)
	if row.Err() != nil {
		return 0, row.Err()
	}
	var gauge float64
	err := row.Scan(&gauge)
	if err != nil {
		return 0, err
	}

	return gauge, nil
}

func (d *DbStorage) GetCounter(name string) (int64, error) {
	getOneSql := `
		SELECT delta
		FROM metric
		WHERE type = $1 AND name = $2
	`
	row := d.db.QueryRow(getOneSql, handlers.MetricTypeCounter, name)
	if row.Err() != nil {
		return 0, row.Err()
	}
	var counter int64
	err := row.Scan(&counter)
	if err != nil {
		return 0, err
	}

	return counter, nil
}

func (d *DbStorage) GetAllGauge() (metrics []handlers.GaugeMetric, err error) {
	getAllSql := `
		SELECT name, value
		FROM metric
		WHERE type = $1
	`
	rows, err := d.db.Query(getAllSql, handlers.MetricTypeGauge)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m handlers.GaugeMetric
		err = rows.Scan(&m.Name, &m.Value)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, m)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return metrics, err
}

func (d *DbStorage) GetAllCounters() (metrics []handlers.CounterMetric, err error) {
	getAllSql := `
		SELECT name, delta
		FROM metric
		WHERE type = $1
	`
	rows, err := d.db.Query(getAllSql, handlers.MetricTypeCounter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m handlers.CounterMetric
		err = rows.Scan(&m.Name, &m.Value)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, m)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return metrics, err
}

func (d *DbStorage) AddObserver(o Observer) {
	d.observers = append(d.observers, o)
}

func (d *DbStorage) notifyObservers(event IEvent) error {
	for _, observer := range d.observers {
		if err := observer.HandleEvent(event); err != nil {
			return err
		}
	}
	return nil
}
