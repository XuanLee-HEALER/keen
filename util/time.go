package util

import "time"

const (
	OLD_DATE        = "1963-11-22 12:30:00"
	COMMON_TIME_FMT = "2006-01-02 15:04:05"
	SERIAL_FMT      = "20060102150405"
)

var (
	OldDate, _ = time.ParseInLocation(COMMON_TIME_FMT, OLD_DATE, time.Local)
)

// Between 两个时间点的差距，返回非负值
func Between(t1, t2 time.Time) time.Duration {
	if t1.Before(t2) {
		return t2.Sub(t1)
	} else {
		return t1.Sub(t2)
	}
}

func LocalParse(layout, strTime string) (time.Time, error) {
	return time.ParseInLocation(layout, strTime, time.Local)
}

func LocalFormat(layout string, t time.Time) string {
	return t.Local().Format(layout)
}
