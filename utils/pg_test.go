package utils

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUUIDToPg(t *testing.T) {
	uid := uuid.New()
	pgUUID, err := UUIDToPg(uid)
	require.NoError(t, err)
	assert.True(t, pgUUID.Valid)
	assert.Equal(t, uid.String(), uuid.UUID(pgUUID.Bytes).String())
}

func TestUUIDToPg_Nil(t *testing.T) {
	_, err := UUIDToPg(uuid.Nil)
	require.NoError(t, err)
}

func TestDurationToInterval(t *testing.T) {
	d := 5*time.Hour + 30*time.Minute
	interval := DurationToInterval(d)
	assert.True(t, interval.Valid)
	assert.Equal(t, d.Microseconds(), interval.Microseconds)
}

func TestDurationToInterval_Zero(t *testing.T) {
	interval := DurationToInterval(0)
	assert.True(t, interval.Valid)
	assert.Equal(t, int64(0), interval.Microseconds)
}

func TestTimeToTimestamptz(t *testing.T) {
	now := time.Now()
	ts := TimeToTimestamptz(now)
	assert.True(t, ts.Valid)
	assert.True(t, ts.Time.Equal(now.UTC()))
}

func TestTimeToTimestamptzRaw_Valid(t *testing.T) {
	now := time.Now()
	ts := TimeToTimestamptzRaw(&now)
	assert.True(t, ts.Valid)
	assert.True(t, ts.Time.Equal(now.UTC()))
}

func TestTimeToTimestamptzRaw_Nil(t *testing.T) {
	ts := TimeToTimestamptzRaw(nil)
	assert.False(t, ts.Valid)
}
