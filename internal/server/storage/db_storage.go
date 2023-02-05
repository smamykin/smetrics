package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/smamykin/smetrics/internal/server/handlers"
)

func NewDBStorage(db *sql.DB) (*DBStorage, error) {
	result := &DBStorage{db: db}
	err := result.init()
	return result, err
}

type DBStorage struct {
	db        *sql.DB
	observers []Observer
}

func (d *DBStorage) init() error {
	tableExistsSQL := "SELECT EXISTS ( SELECT FROM pg_tables WHERE tablename  = 'metric');"
	var isTableExists bool
	err := d.db.QueryRow(tableExistsSQL).Scan(&isTableExists)
	if err != nil {
		return err
	}
	if isTableExists {
		return nil
	}

	_, err = d.db.Exec(`
		CREATE TABLE metric (id SERIAL, name varchar(255) NOT NULL, type varchar(255) NOT NULL, value DOUBLE PRECISION, delta BIGINT, PRIMARY KEY(id));
		CREATE UNIQUE INDEX name_type_unique ON metric (name, type);
	`)

	if err != nil {
		return err
	}

	return nil
}

var upsertCounterSQL = `
	INSERT INTO metric (name, type, delta) 
	VALUES ($1, $2, $3)
	ON CONFLICT (name, type) DO UPDATE 
		SET delta = EXCLUDED.delta
`
var upsertGaugeSQL = `
	INSERT INTO metric (name, type, value) 
	VALUES ($1, $2, $3)
	ON CONFLICT (name, type) DO UPDATE 
		SET value = EXCLUDED.value
`

func (d *DBStorage) UpsertGauge(metric handlers.GaugeMetric) error {
	_, err := d.db.Exec(upsertGaugeSQL, metric.Name, handlers.MetricTypeGauge, metric.Value)

	if err != nil {
		return err
	}

	return d.notifyObservers(AfterUpsertEvent{
		Event{metric},
	})
}

func (d *DBStorage) UpsertCounter(metric handlers.CounterMetric) error {
	_, err := d.db.Exec(upsertCounterSQL, metric.Name, handlers.MetricTypeCounter, metric.Value)
	if err != nil {
		return err
	}

	return d.notifyObservers(AfterUpsertEvent{
		Event{metric},
	})
}

func (d *DBStorage) GetGauge(name string) (float64, error) {
	getOneSQL := `
		SELECT value
		FROM metric
		WHERE type = $1 AND name = $2
	`
	row := d.db.QueryRow(getOneSQL, handlers.MetricTypeGauge, name)
	if row.Err() != nil {
		return 0, row.Err()
	}
	var gauge float64
	err := row.Scan(&gauge)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, handlers.ErrMetricNotFound
		}
		return 0, err
	}

	return gauge, nil
}

func (d *DBStorage) GetCounter(name string) (int64, error) {
	getOneSQL := `
		SELECT delta
		FROM metric
		WHERE type = $1 AND name = $2
	`
	row := d.db.QueryRow(getOneSQL, handlers.MetricTypeCounter, name)
	if row.Err() != nil {
		return 0, row.Err()
	}
	var counter int64
	err := row.Scan(&counter)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, handlers.ErrMetricNotFound
		}
		return 0, err
	}

	return counter, nil
}

func (d *DBStorage) GetAllGauge() (metrics []handlers.GaugeMetric, err error) {
	getAllSQL := `
		SELECT name, value
		FROM metric
		WHERE type = $1
	`
	rows, err := d.db.Query(getAllSQL, handlers.MetricTypeGauge)
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

func (d *DBStorage) GetAllCounters() (metrics []handlers.CounterMetric, err error) {
	getAllSQL := `
		SELECT name, delta
		FROM metric
		WHERE type = $1
	`
	rows, err := d.db.Query(getAllSQL, handlers.MetricTypeCounter)
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

func (d *DBStorage) UpsertMany(ctx context.Context, metrics []interface{}) error {

	// шаг 1 — объявляем транзакцию
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	// шаг 1.1 — если возникает ошибка, откатываем изменения
	defer tx.Rollback()

	// шаг 2 — готовим инструкцию
	stmtGauge, err := tx.PrepareContext(ctx, upsertGaugeSQL)
	if err != nil {
		return err
	}
	defer stmtGauge.Close()
	stmtCounter, err := tx.PrepareContext(ctx, upsertCounterSQL)
	if err != nil {
		return err
	}
	defer stmtCounter.Close()

	// шаг 3 — указываем, что каждое видео будет добавлено в транзакцию
	for _, metric := range metrics {
		switch metric := metric.(type) {
		case handlers.GaugeMetric:
			if _, err = stmtGauge.ExecContext(ctx, metric.Name, handlers.MetricTypeGauge, metric.Value); err != nil {
				return err
			}
		case handlers.CounterMetric:
			if _, err = stmtCounter.ExecContext(ctx, metric.Name, handlers.MetricTypeCounter, metric.Value); err != nil {
				return err
			}
		default:
			return errors.New("unknown metric type")
		}
	}

	// шаг 4 — сохраняем изменения
	err = tx.Commit()
	if err != nil {
		return err
	}

	return d.notifyObservers(AfterUpsertEvent{
		Event{metrics},
	})
}

func (d *DBStorage) AddObserver(o Observer) {
	d.observers = append(d.observers, o)
}

func (d *DBStorage) notifyObservers(event IEvent) error {
	for _, observer := range d.observers {
		if err := observer.HandleEvent(event); err != nil {
			return err
		}
	}
	return nil
}
