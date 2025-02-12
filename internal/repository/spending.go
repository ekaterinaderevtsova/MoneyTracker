package repository

import (
	"cmd/main.go/internal/domain"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SpendingRepository struct {
	db *pgxpool.Pool
}

func NewSpendingRepository(db *pgxpool.Pool) *SpendingRepository {
	return &SpendingRepository{db: db}
}

func (sr *SpendingRepository) AddSpending(ctx context.Context, payload *domain.AddSpending) error {
	_, err := sr.db.Exec(ctx, `
		INSERT INTO spendings (date, sum)
		VALUES ($1, $2);
	`, payload.Date, payload.Sum)
	if err != nil {
		return err
	}
	return nil
}

func (sr *SpendingRepository) GetDaySpendings(ctx context.Context, date time.Time) (int32, error) {
	var total int32
	year := date.Year()
	month := date.Month()
	day := date.Day()

	err := sr.db.QueryRow(ctx, `
		SELECT SUM(sum)
		FROM spendings
		WHERE EXTRACT(YEAR FROM date) = $1
		AND EXTRACT(MONTH FROM date) = $2
		AND EXTRACT(DAY FROM date) = $3;
		`, year, month, day).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (sr *SpendingRepository) GetWeekSpendings(ctx context.Context, date time.Time) (*domain.WeekSpendings, error) {
	var weekSpendings domain.WeekSpendings
	date = date.UTC()
	today := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	startOfWeek := today.AddDate(0, 0, -int(today.Weekday()-1))

	endOfWeek := time.Date(
		startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day()+6,
		23, 59, 59, 0,
		time.UTC,
	)

	rows, err := sr.db.Query(ctx, `
	   SELECT date_series.date AS date,
       COALESCE(SUM(spendings.sum), 0) AS total
       FROM generate_series($1::date, $2::date, '1 day'::interval) AS date_series
       LEFT JOIN spendings ON date_series.date = spendings.date
       GROUP BY date_series.date
       ORDER BY date_series.date;
		`, startOfWeek.Format("2006-01-02"), endOfWeek.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		var daySpendings domain.DaySpendings
		var date time.Time
		err := rows.Scan(&date, &daySpendings.Sum)
		if err != nil {
			return nil, err
		}
		daySpendings.Day = date.Format("02-01")
		weekSpendings.DaySpendings[i] = daySpendings
		weekSpendings.Total += daySpendings.Sum
		i++
	}

	return &weekSpendings, nil
}

func (sr *SpendingRepository) GetMonthSpendings(ctx context.Context, date time.Time) ([]domain.WeekTotalSpending, error) {
	return nil, nil
}
