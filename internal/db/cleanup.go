package db

import (
	"context"
	"log"
	"time"
)

// StartRefreshTokenCleanup deletes expired rows from refresh_tokens every interval.
// Expired tokens are already rejected by the JWT expiry check, so this only
// keeps the table from growing forever.
func (db *DB) StartRefreshTokenCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		tag, err := db.Pool.Exec(
			ctx,
			`DELETE FROM refresh_tokens
			WHERE expires_at < NOW()`,
		)
		if err != nil {
			log.Printf("refresh token cleanup failed: %v", err)
		} else if tag.RowsAffected() > 0 {
			log.Printf("refresh token cleanup: deleted %d expired tokens", tag.RowsAffected())
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
