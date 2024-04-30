package domain

import "time"

type PointChange struct {
	Id           int64
	Uid          int64
	ChangeAmount int64
	Reason       string
	Source       string
	Utime        time.Time
	Ctime        time.Time
}

type UserPointInfo struct {
	Uid             int64
	Points          int64
	Level           int64
	NextLevelPoints int64
}
