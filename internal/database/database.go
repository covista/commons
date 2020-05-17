package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/covista/commons/internal/config"
	"github.com/covista/commons/internal/logging"
	"github.com/covista/commons/proto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func max(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func eninToTimestamp(enin uint32) time.Time {
	return time.Unix(int64(enin*600), 0).UTC()
}

func timestampInRange(ts, start, end time.Time) bool {
	return ts.Equal(start) || ts.Equal(end) || (ts.Before(end) && ts.After(start))
}

// Database object providing pooled connections to the underlying postgres database
type Database struct {
	pool *pgxpool.Pool
}

// Creates a new Database instance from the insecure defaults given in the docker compose file.
// Helpful for testing.
func NewWithInsecureDefaults(ctx context.Context) (*Database, error) {
	cfg := &config.Config{
		Database: config.Database{
			Host:     "localhost",
			Database: "covid19",
			User:     "covid19",
			Password: "covid19databasepassword",
			Port:     "5434",
		},
	}
	return NewFromConfig(ctx, cfg)
}

// Creates a new Database instance from the given configuration
func NewFromConfig(ctx context.Context, cfg *config.Config) (*Database, error) {
	if err := checkConfig(cfg); err != nil {
		return nil, fmt.Errorf("Invalid config to connect to database: %w", err)
	}
	db_connection_url := fmt.Sprintf("postgres://%s/%s?sslmode=disable&user=%s&password=%s&port=%s",
		cfg.Database.Host, cfg.Database.Database, cfg.Database.User, url.QueryEscape(cfg.Database.Password), cfg.Database.Port)

	log := logging.FromContext(ctx)
	// loop until database is live
	var pool *pgxpool.Pool
	var err error
	for {
		pool, err = pgxpool.Connect(ctx, db_connection_url)
		if err != nil {
			log.Warnf("Failed to connect to database (%s); retrying in 5 seconds", err.Error())
			time.Sleep(5 * time.Second)
		}
		break
	}
	log.Infof("Connected to postgres at %s", cfg.Database.Host)
	return &Database{
		pool: pool,
	}, nil
}

func (db *Database) Close() {
	db.pool.Close()
}

func (db *Database) RunAsTransaction(ctx context.Context, f func(txn pgx.Tx) error) error {
	// start transaction in a new pooled connection
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("Could not acquire connection from pool: %w", err)
	}
	defer conn.Release()
	txn, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Could not begin transaction: %w", err)
	}
	if err := f(txn); err != nil {
		if rberr := txn.Rollback(ctx); rberr != nil {
			return fmt.Errorf("Error (%w) occured during transaction. Could not rollback: %w", err, rberr)
		}
		return fmt.Errorf("Error occured during transaction execution: %w", err)
	}
	if err := txn.Commit(ctx); err != nil {
		return fmt.Errorf("Error occured during transaction commit: %w", err)
	}
	return nil
}

// Create a one-time use authorization key to be given to a patient.
func (db *Database) CreateAuthorizationKey(ctx context.Context, request *proto.TokenRequest) ([]byte, error) {
	var one_time_auth_key uuid.UUID
	// check sanity of tokenrequest
	if err := checkTokenRequest(request); err != nil {
		return nil, fmt.Errorf("Invalid TokenRequest: %w", err)
	}

	log := logging.FromContext(ctx)

	err := db.RunAsTransaction(ctx, func(txn pgx.Tx) error {
		// Checks the validity of the health authority API key by looking at the health_authorities table
		var (
			authority_id []byte
			name         string
		)
		err := txn.QueryRow(ctx, "SELECT authority_id, name FROM health_authorities where api_key=$1", request.ApiKey).Scan(&authority_id, &name)
		if err != nil {
			return fmt.Errorf("Invalid api_key: %w", err)
		}
		// api_key is valid!
		log.Infof("Generating one-time auth key for authority %s (%x)", name, authority_id)

		// generate a new one-time authorization key
		one_time_auth_key, err = uuid.NewRandom()
		if err != nil {
			return fmt.Errorf("Could not generate one-time auth key: %w", err)
		}
		// parse timestamps
		permitted_start, err := time.Parse(time.RFC3339, request.PermittedRangeStart)
		if err != nil {
			return fmt.Errorf("Could not parse permitted_range_start: %w", err)
		}
		permitted_end, err := time.Parse(time.RFC3339, request.PermittedRangeEnd)
		if err != nil {
			return fmt.Errorf("Could not parse permitted_range_end: %w", err)
		}

		// insert the one-time auth key into the authorization_keys table
		_, err = txn.Exec(ctx, `INSERT INTO authorization_keys
					   (authorization_key, api_key, key_type, permitted_start, permitted_end) 
					   VALUES ($1, $2, $3, $4, $5) ON CONFLICT (authorization_key) DO NOTHING`,
			one_time_auth_key[:], request.ApiKey, "DIAGNOSED", permitted_start.UTC(), permitted_end.UTC())
		if err != nil {
			return fmt.Errorf("Could not insert new one-time auth key: %w", err)
		}

		return nil
	})
	return one_time_auth_key[:], err
}

func (db *Database) AddReport(ctx context.Context, report *proto.Report) error {
	if err := checkReport(report); err != nil {
		return fmt.Errorf("Invalid Report: %w", err)
	}
	log := logging.FromContext(ctx)

	err := db.RunAsTransaction(ctx, func(txn pgx.Tx) error {
		var permitted_start, permitted_end time.Time

		// validate that authorization_key is valid
		log.Infof("New report with auth key %x", report.AuthorizationKey)
		err := txn.QueryRow(ctx, `SELECT permitted_start, permitted_end FROM authorization_keys
						   WHERE authorization_key = $1`, report.AuthorizationKey).Scan(&permitted_start, &permitted_end)
		if err != nil {
			return fmt.Errorf("Could not validate authorization key: %w", err)
		}

		// For each TEK, ENIN pair, check that it is within the valid range permitted by the authorization key
		for idx, tstek := range report.Reports {
			timestamp := eninToTimestamp(tstek.ENIN)
			if !timestampInRange(timestamp, permitted_start, permitted_end) {
				return fmt.Errorf("Report %d (%s) was not in valid range [%s, %s])", idx, timestamp, permitted_start, permitted_end)
			}
			// insert the TEK, ENIN into the database if it is valid. If there are any errors, this will all be rolled
			// back and no values from this report will be inserted
			_, err := txn.Exec(ctx, `INSERT INTO reported_keys(TEK, ENIN, authorization_key)
									 VALUES($1, $2, $3) ON CONFLICT (TEK) DO NOTHING`, tstek.TEK, timestamp, report.AuthorizationKey)
			if err != nil {
				return fmt.Errorf("Could not insert report %d into database: %w", idx, err)
			}
		}

		return nil
	})
	return err
}

func (db *Database) GetDiagnosisKeys(ctx context.Context, request *proto.GetKeyRequest) (chan *proto.TimestampedTEK, chan error) {
	results := make(chan *proto.TimestampedTEK)
	errchan := make(chan error, 1)

	if err := checkGetKeyRequest(request); err != nil {
		errchan <- fmt.Errorf("Invalid GetKeyRequest: %w", err)
		return results, errchan
	}

	log := logging.FromContext(ctx)
	log.Debugf("Fetching keys for query: health_authority=%x, enin=%d, hrange=%v", request.AuthorityId, request.ENIN, request.Hrange)

	// construct the SQL query for the provided filter
	query, values, err := buildQuery(request)
	if err != nil {
		errchan <- fmt.Errorf("Could not construct query for GetKeyRequest: %w", err)
		return results, errchan
	}

	go func() {
		err := db.RunAsTransaction(ctx, func(txn pgx.Tx) error {
			rows, err := txn.Query(ctx, query, values...)
			if err != nil {
				return fmt.Errorf("Could not get reported keys: %w", err)
			}
			defer rows.Close()
			for rows.Next() {
				var tek []byte
				var enin time.Time
				if err := rows.Scan(&tek, &enin); err != nil {
					return fmt.Errorf("Error getting tek, enin: %w", err)
				}
				results <- &proto.TimestampedTEK{
					TEK:  tek,
					ENIN: uint32(enin.Unix() / 600),
				}
			}
			close(results)
			return rows.Err()
		})
		if err != nil {
			errchan <- fmt.Errorf("Could not download reported keys: %w", err)
		}
		close(errchan)
	}()

	return results, errchan
}
