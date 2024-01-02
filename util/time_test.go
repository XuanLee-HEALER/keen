package util_test

import (
	"testing"
	"time"

	"gitea.fcdm.top/lixuan/keen/util"
	"github.com/stretchr/testify/assert"
)

func TestOldTime(t *testing.T) {
	assert.Equal(t, time.Date(1963, 11, 22, 12, 30, 0, 0, time.Local), util.OldDate)
}

func TestBetween(t *testing.T) {
	d1 := time.Date(2024, 1, 2, 14, 18, 0, 0, time.Local)
	d2 := time.Date(2024, 1, 2, 14, 20, 0, 0, time.Local)

	assert.Equal(t, 120*time.Second, util.Between(d1, d2), "computation fail")

	d3 := time.Date(2024, 1, 2, 14, 22, 0, 0, time.Local)
	assert.Equal(t, 120*time.Second, util.Between(d3, d2), "computation fail")
}
