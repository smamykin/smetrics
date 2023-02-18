// Integration tests of the db_storage
package storage

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/smamykin/smetrics/internal/server/handlers"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestDBStorage_GetAllCounters(t *testing.T) {
	skipIfNoDatabaseURL(t)

	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	require.Nil(t, err)
	defer db.Close()

	dbStorage, err := NewDBStorage(db)
	require.Nil(t, err)

	prepareDBBeforeTest(db, t)

	counters, err := dbStorage.GetAllCounters()
	require.Nil(t, err)

	require.Equal(t, []handlers.CounterMetric{
		{Name: "metric-a", Value: 11},
		{Name: "metric-b", Value: 22},
	}, counters)
}

func TestDBStorage_GetAllGauge(t *testing.T) {
	skipIfNoDatabaseURL(t)

	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	require.Nil(t, err)
	defer db.Close()

	dbStorage, err := NewDBStorage(db)
	require.Nil(t, err)

	prepareDBBeforeTest(db, t)

	counters, err := dbStorage.GetAllGauge()
	require.Nil(t, err)

	require.Equal(t, []handlers.GaugeMetric{
		{Name: "metric-c", Value: 33.44},
		{Name: "metric-d", Value: 55.66},
	}, counters)
}

func TestDBStorage_GetCounter(t *testing.T) {
	skipIfNoDatabaseURL(t)

	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	require.Nil(t, err)
	defer db.Close()

	dbStorage, err := NewDBStorage(db)
	require.Nil(t, err)
	prepareDBBeforeTest(db, t)

	counter, err := dbStorage.GetCounter("metric-a")
	require.Nil(t, err)
	require.Equal(t, int64(11), counter)

	_, err = dbStorage.GetCounter("metric-non-existed")
	require.NotNil(t, err)
	require.Equal(t, handlers.ErrMetricNotFound, err)
}

func TestDBStorage_GetGauge(t *testing.T) {
	skipIfNoDatabaseURL(t)

	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	require.Nil(t, err)
	defer db.Close()

	dbStorage, err := NewDBStorage(db)
	require.Nil(t, err)
	prepareDBBeforeTest(db, t)

	counter, err := dbStorage.GetGauge("metric-c")
	require.Nil(t, err)
	require.Equal(t, 33.44, counter)

	_, err = dbStorage.GetGauge("metric-non-existed")
	require.NotNil(t, err)
	require.Equal(t, handlers.ErrMetricNotFound, err)
}

func TestDBStorage_UpsertCounter(t *testing.T) {
	skipIfNoDatabaseURL(t)

	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	require.Nil(t, err)
	defer db.Close()

	dbStorage, err := NewDBStorage(db)
	require.Nil(t, err)
	prepareDBBeforeTest(db, t)

	//insert
	metric := handlers.CounterMetric{Name: "metric-z", Value: 222}
	err = dbStorage.UpsertCounter(metric)
	require.Nil(t, err)

	actual, err := dbStorage.GetCounter("metric-z")
	require.Nil(t, err)
	require.Equal(t, metric.Value, actual)

	//update
	metric = handlers.CounterMetric{Name: "metric-z", Value: 333}
	err = dbStorage.UpsertCounter(metric)
	require.Nil(t, err)

	actual, err = dbStorage.GetCounter("metric-z")
	require.Nil(t, err)
	require.Equal(t, metric.Value, actual)

}

func TestDBStorage_UpsertGauge(t *testing.T) {
	skipIfNoDatabaseURL(t)

	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	require.Nil(t, err)
	defer db.Close()

	dbStorage, err := NewDBStorage(db)
	require.Nil(t, err)
	prepareDBBeforeTest(db, t)

	//insert
	metric := handlers.GaugeMetric{Name: "metric-z", Value: 222.333}
	err = dbStorage.UpsertGauge(metric)
	require.Nil(t, err)

	actual, err := dbStorage.GetGauge("metric-z")
	require.Nil(t, err)
	require.Equal(t, metric.Value, actual)

	//update
	metric = handlers.GaugeMetric{Name: "metric-z", Value: 444.555}
	err = dbStorage.UpsertGauge(metric)
	require.Nil(t, err)

	actual, err = dbStorage.GetGauge("metric-z")
	require.Nil(t, err)
	require.Equal(t, metric.Value, actual)

}
func TestDBStorage_UpsertMany(t *testing.T) {
	skipIfNoDatabaseURL(t)

	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	require.Nil(t, err)
	defer db.Close()

	dbStorage, err := NewDBStorage(db)
	require.Nil(t, err)
	prepareDBBeforeTest(db, t)

	//insert
	metricGauge := handlers.GaugeMetric{Name: "metric-z", Value: 222.333}
	metricCounter := handlers.CounterMetric{Name: "metric-y", Value: 222}
	metrics := []interface{}{metricGauge, metricCounter}
	err = dbStorage.UpsertMany(context.Background(), metrics)
	require.Nil(t, err)

	actualGauge, err := dbStorage.GetGauge("metric-z")
	require.Nil(t, err)
	require.Equal(t, metricGauge.Value, actualGauge)
	actualCounter, err := dbStorage.GetCounter("metric-y")
	require.Nil(t, err)
	require.Equal(t, metricCounter.Value, actualCounter)

	//update
	metricGauge = handlers.GaugeMetric{Name: "metric-z", Value: 444.555}
	metricCounter = handlers.CounterMetric{Name: "metric-y", Value: 333}
	metrics = []interface{}{metricGauge, metricCounter}
	err = dbStorage.UpsertMany(context.Background(), metrics)
	require.Nil(t, err)

	actualGauge, err = dbStorage.GetGauge("metric-z")
	require.Nil(t, err)
	require.Equal(t, metricGauge.Value, actualGauge)
	actualCounter, err = dbStorage.GetCounter("metric-y")
	require.Nil(t, err)
	require.Equal(t, metricCounter.Value, actualCounter)
}

func TestDBStorage_init(t *testing.T) {
	skipIfNoDatabaseURL(t)

	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	dropTableIfExists(db, t)

	dbStorage := DBStorage{db, nil}
	err = dbStorage.init()
	if err != nil {
		t.Error(err)
	}
	assertTableExist(db, t)

	//second run, table already exists
	dbStorage = DBStorage{db, nil}
	err = dbStorage.init()
	if err != nil {
		t.Error(err)
	}
}

func assertTableExist(db *sql.DB, t *testing.T) {
	sqlStmt := "SELECT EXISTS ( SELECT FROM pg_tables WHERE tablename  = 'metric');"

	var result bool
	row := db.QueryRow(sqlStmt)
	err := row.Scan(&result)
	if err != nil {
		t.Error(err)
	}

	require.True(t, result)

}

func skipIfNoDatabaseURL(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping integration test with db.")
	}
}

func dropTableIfExists(db *sql.DB, t *testing.T) {
	_, err := db.Exec("DROP TABLE IF EXISTS metric")
	if err != nil {
		t.Error("error while droping db")
	}
}

func truncateTable(db *sql.DB, t *testing.T) {
	_, err := db.Exec("TRUNCATE TABLE metric")

	if err != nil {
		t.Error("error while truncating db")
	}
}

func prepareDBBeforeTest(db *sql.DB, t *testing.T) {
	truncateTable(db, t)
	_, err := db.Exec(`
		INSERT INTO metric (name, type, delta, value) 
		VALUES 
		    ('metric-a', 'counter', 11, null),
		    ('metric-c', 'gauge', null, 33.44),
		    ('metric-b', 'counter', 22, null), 
		    ('metric-d', 'gauge', null, 55.66)
	`)
	require.Nil(t, err)
}
