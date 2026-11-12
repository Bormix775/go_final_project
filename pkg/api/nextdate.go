package api

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "20060102"

const searchLimit = 10

var (
	ErrInvalidDateFormat    = errors.New("неверный формат даты")
	ErrInvalidRepeatFormat  = errors.New("неверный формат правила повторения")
	ErrUnsupportedRule      = errors.New("неподдерживаемое правило повторения")
	ErrInvalidIntervalValue = errors.New("недопустимое значение интервала")
	ErrNoMatchDate          = errors.New("правило повторения не содержит допустимой даты")
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
			wdNorm := (int(date.Weekday())+6)%7 + 1

			if dayMask[wdNorm] && afterNow(date, now) {
				return date.Format(dateLayout), nil
			}

			date = date.AddDate(0, 0, 1)
		}
	}

	if rule == "m" {
		// Формат правила: "m <дни> [месяцы]"
		// Дни — числа месяца через запятую: 1..31, а также -1 и -2
		// (последний и предпоследний день месяца)
		// Месяцы — необязательная вторая часть: номера 1..12 через запятую
		if len(parts) < 2 {
			return "", ErrInvalidRepeatFormat
		}

		subParts := strings.Split(parts[1], " ")

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
		if len(subParts) > 1 {
			for _, mStr := range strings.Split(subParts[1], ",") {
				m, err := strconv.Atoi(mStr)
				if err != nil || m < 1 || m > 12 {
					return "", ErrInvalidRepeatFormat
				}
				validMonths[m] = true
			}
		}
		// Ограничиваем поиск во времени: правило может быть неосуществимо
		// (например, "m 31 2" — 31-го февраля не бывает), и без предела
		// цикл выполнялся бы вечно, подвешивая сервер.
		limit := now.AddDate(searchLimit, 0, 0)
		for date.Before(limit) {
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
			matchMonth := len(subParts) <= 1 || validMonths[mo]

			if matchDay && matchMonth && afterNow(date, now) {
				return date.Format(dateLayout), nil
			}
			date = date.AddDate(0, 0, 1)
		}
		return "", ErrNoMatchDate
	}
	return "", ErrUnsupportedRule
}
