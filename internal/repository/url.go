package repository

import (
	"context"
	"errors"
	"go-shorten/internal/store"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	URLCleanupInterval = 5 * time.Minute
)

func CleanUpExpiredURLs(ctx context.Context) {
	if rows, err := deleteExpiredURLs(ctx); err != nil {
		log.Println("error while cleaning up urls:", err.Error())
	} else {
		log.Println("cleaned up", rows, "expired urls")
	}
}

func StartOldURLsCleanup() {
	go func() {
		CleanUpExpiredURLs(context.Background())

		ticker := time.NewTicker(URLCleanupInterval)
		for range ticker.C {
			CleanUpExpiredURLs(context.Background())
		}
	}()
}

var (
	ErrorURLExpired = errors.New("url code expired")
)

func AddShortenedUrl(c context.Context, url, code string, days int) error {
	conn, err := store.New().GetConnection(c)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `INSERT INTO URL (url_string, url_code, expire_at) VALUES ($1, $2, NOW() + make_interval(days => $3))`
	_, err = conn.Conn().Exec(c, query, url, code, days)
	if err != nil {
		return err
	}

	return nil
}

func GetShortenedURL(c context.Context, code string) (string, error) {
	var url string
	var now, expireAt time.Time

	conn, err := store.New().GetConnection(c)
	if err != nil {
		return "", err
	}
	defer conn.Release()

	query := `SELECT url_string, NOW(), expire_at FROM URL WHERE url_code = $1 AND expired = false;`
	rows := conn.Conn().QueryRow(c, query, code)
	if err := rows.Scan(&url, &now, &expireAt); err != nil {
		return "", err
	}

	if now.Compare(expireAt) == 1 {
		if err := markExpired(c, code); err != nil {
			return "", ErrorURLExpired
		}

		return "", nil
	}

	return url, nil
}

func markExpired(c context.Context, code string) error {
	query := `UPDATE URL SET expired = true WHERE url_code = $1 AND NOW() > expire_at;`
	conn, err := store.New().GetConnection(c)
	if err != nil {
		return err
	}
	defer conn.Release()

	if _, err := conn.Exec(c, query, code); err != nil {
		return err
	}

	return nil
}

func CheckCodeExists(c context.Context, code string) (int, error) {
	query := `SELECT 1 FROM URL WHERE url_code = $1;`
	conn, err := store.New().GetConnection(c)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	rows := conn.QueryRow(c, query, code)
	if err := rows.Scan(nil); errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	return 1, nil
}

func deleteExpiredURLs(c context.Context) (int64, error) {
	query := `DELETE FROM URL WHERE expired = true OR NOW() > expire_at;`
	conn, err := store.New().GetConnection(c)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	if tag, err := conn.Exec(c, query); err != nil {
		return 0, err
	} else {
		return tag.RowsAffected(), nil
	}
}
