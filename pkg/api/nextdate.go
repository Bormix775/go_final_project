package api

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "20060102"

var (
	ErrInvalidDateFormat    = errors.New("неверный формат даты")
	ErrInvalidRepeatFormat  = errors.New("неверный формат правила повторения")
	ErrUnsupportedRule      = errors.New("неподдерживаемое правило повторения")
	ErrInvalidIntervalValue = errors.New("недопустимое значение интервала")
)

func afterNow(date, now time.Time) bool {
	t := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	n := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return t.After(n)
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	date, err := time.Parse(dateLayout, dstart)
	if err != nil {
		return "", ErrInvalidDateFormat
	}

	r := strings.TrimSpace(repeat)
	if r == "" {
		return "", ErrInvalidRepeatFormat
	}

	parts := strings.SplitN(r, " ", 2)
	rule := parts[0]

	if rule == "y" {
		for {
			date = date.AddDate(1, 0, 0)
			if afterNow(date, now) {
				return date.Format(dateLayout), nil
			}
		}
	}

	if rule == "d" {
		if len(parts) < 2 {
			return "", ErrInvalidRepeatFormat
		}
		interval, err := strconv.Atoi(parts[1])
		if err != nil || interval <= 0 || interval > 400 {
			return "", ErrInvalidIntervalValue
		}

		date = date.AddDate(0, 0, interval)
		for {
			if afterNow(date, now) {
				return date.Format(dateLayout), nil
			}
			date = date.AddDate(0, 0, interval)
		}
	}

	if rule == "w" {
		if len(parts) < 2 {
			return "", ErrInvalidRepeatFormat
		}

		daysList := strings.Split(parts[1], ",")
		var dayMask [8]bool

		for _, dStr := range daysList {
			d, err := strconv.Atoi(dStr)
			if err != nil || d < 1 || d > 7 {
				return "", ErrInvalidRepeatFormat
			}
			dayMask[d] = true
		}

		for {
			wd := int(date.Weekday())

			wdNorm := (wd+6)%7 + 1

			if dayMask[wdNorm] && afterNow(date, now) {
				return date.Format(dateLayout), nil
			}

			date = date.AddDate(0, 0, 1)
		}
	}

	if rule == "m" {
		if len(parts) < 2 {
			return "", ErrInvalidRepeatFormat
		}

		subParts := strings.Split(parts[1], " ")
		if len(subParts) == 0 {
			return "", ErrInvalidRepeatFormat
		}

		days := strings.Split(subParts[0], ",")
		var validDays [32]bool
		hasLast := false
		hasSecondLast := false

		for _, dStr := range days {
			d, err := strconv.Atoi(dStr)
			if err != nil {
				return "", ErrInvalidRepeatFormat
			}
			if d > 0 && d <= 31 {
				validDays[d] = true
			} else if d == -1 {
				hasLast = true
			} else if d == -2 {
				hasSecondLast = true
			} else {
				return "", ErrInvalidRepeatFormat
			}
		}

		var validMonths [13]bool
		monthsOk := true
		if len(subParts) > 1 {
			monthList := strings.Split(subParts[1], ",")
			for _, mStr := range monthList {
				m, err := strconv.Atoi(mStr)
				if err != nil || m < 1 || m > 12 {
					monthsOk = false
					break
				}
				validMonths[m] = true
			}
			if !monthsOk {
				return "", ErrInvalidRepeatFormat
			}
		}

		for {
			dy := date.Day()
			mo := int(date.Month())
			yr := date.Year()
			loc := date.Location()

			matchDay := validDays[dy]

			lastDay := time.Date(yr, time.Month(mo), 1, 0, 0, 0, 0, loc).AddDate(0, 1, -1).Day()

			if hasLast && dy == lastDay {
				matchDay = true
			}
			if hasSecondLast && dy == lastDay-1 {
				matchDay = true
			}

			matchMonth := true
			if len(subParts) > 1 {
				matchMonth = validMonths[mo]
			}

			if matchDay && matchMonth && afterNow(date, now) {
				return date.Format(dateLayout), nil
			}

			date = date.AddDate(0, 0, 1)
		}
	}

	return "", ErrUnsupportedRule
}
