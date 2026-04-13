package storage

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	app "github.com/vasary/job-search-mcp/internal/application"
)

func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *PostgresRepository) CreateApplication(ctx context.Context, input app.CreateApplicationInput, now time.Time) (app.ApplicationDetails, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return app.ApplicationDetails{}, err
	}
	defer tx.Rollback(ctx)

	const insertApplication = `
		insert into %s (
			company_name,
			position_title,
			source_url,
			work_type,
			salary,
			position_description,
			tech_stack,
			created_at,
			updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		returning id
	`

	var applicationID int64
	err = tx.QueryRow(
		ctx,
		fmt.Sprintf(insertApplication, r.applicationsTable()),
		input.CompanyName,
		nullIfEmpty(input.PositionTitle),
		nullIfEmpty(input.SourceURL),
		nullIfEmpty(input.WorkType),
		nullIfEmpty(input.Salary),
		nullIfEmpty(input.PositionDescription),
		nullIfEmpty(input.TechStack),
		now,
	).Scan(&applicationID)
	if err != nil {
		return app.ApplicationDetails{}, err
	}

	if _, err := r.insertStatusChange(ctx, tx, app.ChangeStatusInput{
		ApplicationID: applicationID,
		Status:        input.InitialStatus,
		Note:          input.InitialStatusNote,
	}, now); err != nil {
		return app.ApplicationDetails{}, err
	}

	if input.InitialComment != "" {
		if _, err := r.insertComment(ctx, tx, app.AddCommentInput{
			ApplicationID: applicationID,
			Body:          input.InitialComment,
		}, now); err != nil {
			return app.ApplicationDetails{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return app.ApplicationDetails{}, err
	}

	return r.GetApplication(ctx, applicationID)
}

func (r *PostgresRepository) UpdateApplication(ctx context.Context, input app.UpdateApplicationInput, now time.Time) (app.ApplicationDetails, error) {
	current, err := r.GetApplication(ctx, input.ID)
	if err != nil {
		return app.ApplicationDetails{}, err
	}

	if input.CompanyName != nil {
		current.CompanyName = *input.CompanyName
	}
	if input.PositionTitle != nil {
		current.PositionTitle = *input.PositionTitle
	}
	if input.SourceURL != nil {
		current.SourceURL = *input.SourceURL
	}
	if input.WorkType != nil {
		current.WorkType = *input.WorkType
	}
	if input.Salary != nil {
		current.Salary = *input.Salary
	}
	if input.PositionDescription != nil {
		current.PositionDescription = *input.PositionDescription
	}
	if input.TechStack != nil {
		current.TechStack = *input.TechStack
	}

	const query = `
		update %s
		set
			company_name = $2,
			position_title = $3,
			source_url = $4,
			work_type = $5,
			salary = $6,
			position_description = $7,
			tech_stack = $8,
			updated_at = $9
		where id = $1
	`

	tag, err := r.pool.Exec(
		ctx,
		fmt.Sprintf(query, r.applicationsTable()),
		input.ID,
		current.CompanyName,
		nullIfEmpty(current.PositionTitle),
		nullIfEmpty(current.SourceURL),
		nullIfEmpty(current.WorkType),
		nullIfEmpty(current.Salary),
		nullIfEmpty(current.PositionDescription),
		nullIfEmpty(current.TechStack),
		now,
	)
	if err != nil {
		return app.ApplicationDetails{}, err
	}
	if tag.RowsAffected() == 0 {
		return app.ApplicationDetails{}, app.ErrNotFound
	}

	return r.GetApplication(ctx, input.ID)
}

func (r *PostgresRepository) GetApplication(ctx context.Context, id int64) (app.ApplicationDetails, error) {
	const applicationQuery = `
		select
			a.id,
			a.company_name,
			a.position_title,
			a.source_url,
			a.work_type,
			a.salary,
			a.position_description,
			a.tech_stack,
			a.created_at,
			a.updated_at,
			sh.status,
			sh.changed_at
		from %s a
		join lateral (
			select status, changed_at
			from %s
			where application_id = a.id
			order by changed_at desc, id desc
			limit 1
		) sh on true
		where a.id = $1
	`

	var details app.ApplicationDetails
	var (
		positionTitle       *string
		sourceURL           *string
		workType            *string
		salary              *string
		positionDescription *string
		techStack           *string
		currentStatus       string
	)

	err := r.pool.QueryRow(ctx, fmt.Sprintf(applicationQuery, r.applicationsTable(), r.statusHistoryTable()), id).Scan(
		&details.ID,
		&details.CompanyName,
		&positionTitle,
		&sourceURL,
		&workType,
		&salary,
		&positionDescription,
		&techStack,
		&details.CreatedAt,
		&details.UpdatedAt,
		&currentStatus,
		&details.LastStatusChangedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return app.ApplicationDetails{}, app.ErrNotFound
		}
		return app.ApplicationDetails{}, err
	}

	details.PositionTitle = derefString(positionTitle)
	details.SourceURL = derefString(sourceURL)
	details.WorkType = derefString(workType)
	details.Salary = derefString(salary)
	details.PositionDescription = derefString(positionDescription)
	details.TechStack = derefString(techStack)
	details.CurrentStatus = app.ApplicationStatus(currentStatus)

	if details.Comments, err = r.listComments(ctx, id); err != nil {
		return app.ApplicationDetails{}, err
	}
	if details.StatusHistory, err = r.listStatusHistory(ctx, id); err != nil {
		return app.ApplicationDetails{}, err
	}
	if details.Documents, err = r.listDocuments(ctx, id); err != nil {
		return app.ApplicationDetails{}, err
	}

	return details, nil
}

func (r *PostgresRepository) ListApplications(ctx context.Context, input app.ListApplicationsInput) (app.ListApplicationsOutput, error) {
	query, args := r.buildListApplicationsQuery(input)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return app.ListApplicationsOutput{}, err
	}
	defer rows.Close()

	out := app.ListApplicationsOutput{Items: []app.ApplicationSummary{}}
	for rows.Next() {
		item, total, err := scanApplicationSummary(rows)
		if err != nil {
			return app.ListApplicationsOutput{}, err
		}
		out.Total = total
		out.Items = append(out.Items, item)
	}

	if err := rows.Err(); err != nil {
		return app.ListApplicationsOutput{}, err
	}
	return out, nil
}

func (r *PostgresRepository) SearchApplications(ctx context.Context, input app.SearchApplicationsInput) (app.ListApplicationsOutput, error) {
	query, args := r.buildSearchApplicationsQuery(input)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return app.ListApplicationsOutput{}, err
	}
	defer rows.Close()

	out := app.ListApplicationsOutput{Items: []app.ApplicationSummary{}}
	for rows.Next() {
		item, total, err := scanApplicationSummary(rows)
		if err != nil {
			return app.ListApplicationsOutput{}, err
		}
		out.Total = total
		out.Items = append(out.Items, item)
	}

	if err := rows.Err(); err != nil {
		return app.ListApplicationsOutput{}, err
	}
	return out, nil
}

func (r *PostgresRepository) GetRecentApplications(ctx context.Context, input app.RecentApplicationsInput) (app.ListApplicationsOutput, error) {
	query, args := r.buildRecentApplicationsQuery(input)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return app.ListApplicationsOutput{}, err
	}
	defer rows.Close()

	out := app.ListApplicationsOutput{Items: []app.ApplicationSummary{}}
	for rows.Next() {
		item, total, err := scanApplicationSummary(rows)
		if err != nil {
			return app.ListApplicationsOutput{}, err
		}
		out.Total = total
		out.Items = append(out.Items, item)
	}

	if err := rows.Err(); err != nil {
		return app.ListApplicationsOutput{}, err
	}
	return out, nil
}

func (r *PostgresRepository) GetApplicationStats(ctx context.Context) (app.ApplicationStats, error) {
	query := fmt.Sprintf(`
		select sh.status, count(*)
		from %s a
		join lateral (
			select status
			from %s
			where application_id = a.id
			order by changed_at desc, id desc
			limit 1
		) sh on true
		group by sh.status
	`, r.applicationsTable(), r.statusHistoryTable())

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return app.ApplicationStats{}, err
	}
	defer rows.Close()

	stats := app.ApplicationStats{ByStatus: map[app.ApplicationStatus]int{}}
	for rows.Next() {
		var (
			status string
			count  int
		)
		if err := rows.Scan(&status, &count); err != nil {
			return app.ApplicationStats{}, err
		}
		stats.ByStatus[app.ApplicationStatus(status)] = count
		stats.Total += count
	}

	if err := rows.Err(); err != nil {
		return app.ApplicationStats{}, err
	}
	return stats, nil
}

func (r *PostgresRepository) buildListApplicationsQuery(input app.ListApplicationsInput) (string, []any) {
	var (
		args       []any
		conditions []string
	)

	limit := input.Limit
	if limit <= 0 {
		limit = 100
	}
	args = append(args, limit, input.Offset)

	if input.CompanyName != "" {
		args = append(args, "%"+strings.ToLower(input.CompanyName)+"%")
		conditions = append(conditions, fmt.Sprintf("lower(a.company_name) like $%d", len(args)))
	}
	if input.PositionTitle != "" {
		args = append(args, "%"+strings.ToLower(input.PositionTitle)+"%")
		conditions = append(conditions, fmt.Sprintf("lower(coalesce(a.position_title, '')) like $%d", len(args)))
	}
	if input.CurrentStatus != "" {
		args = append(args, string(input.CurrentStatus))
		conditions = append(conditions, fmt.Sprintf("sh.status = $%d", len(args)))
	}

	where := ""
	if len(conditions) > 0 {
		where = " where " + strings.Join(conditions, " and ")
	}

	query := `
		select
			a.id,
			a.company_name,
			a.position_title,
			a.source_url,
			a.work_type,
			a.salary,
			a.position_description,
			a.tech_stack,
			a.created_at,
			a.updated_at,
			sh.status,
			sh.changed_at,
			count(*) over()
		from %s a
		join lateral (
			select status, changed_at
			from %s
			where application_id = a.id
			order by changed_at desc, id desc
			limit 1
		) sh on true` + where + `
		order by sh.changed_at desc, a.id desc
		limit $1 offset $2
	`

	return fmt.Sprintf(query, r.applicationsTable(), r.statusHistoryTable()), args
}

func (r *PostgresRepository) buildSearchApplicationsQuery(input app.SearchApplicationsInput) (string, []any) {
	var (
		args       []any
		conditions []string
	)

	limit := input.Limit
	if limit <= 0 {
		limit = 100
	}
	args = append(args, limit, input.Offset)

	if input.Query != "" {
		args = append(args, "%"+strings.ToLower(input.Query)+"%")
		param := "$" + strconv.Itoa(len(args))
		conditions = append(conditions, "("+strings.Join([]string{
			"lower(a.company_name) like " + param,
			"lower(coalesce(a.position_title, '')) like " + param,
			"lower(coalesce(a.tech_stack, '')) like " + param,
		}, " or ")+")")
	}
	if input.CurrentStatus != "" {
		args = append(args, string(input.CurrentStatus))
		conditions = append(conditions, fmt.Sprintf("sh.status = $%d", len(args)))
	}

	where := ""
	if len(conditions) > 0 {
		where = " where " + strings.Join(conditions, " and ")
	}

	query := `
		select
			a.id,
			a.company_name,
			a.position_title,
			a.source_url,
			a.work_type,
			a.salary,
			a.position_description,
			a.tech_stack,
			a.created_at,
			a.updated_at,
			sh.status,
			sh.changed_at,
			count(*) over()
		from %s a
		join lateral (
			select status, changed_at
			from %s
			where application_id = a.id
			order by changed_at desc, id desc
			limit 1
		) sh on true` + where + `
		order by sh.changed_at desc, a.id desc
		limit $1 offset $2
	`

	return fmt.Sprintf(query, r.applicationsTable(), r.statusHistoryTable()), args
}

func (r *PostgresRepository) buildRecentApplicationsQuery(input app.RecentApplicationsInput) (string, []any) {
	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}

	query := `
		select
			a.id,
			a.company_name,
			a.position_title,
			a.source_url,
			a.work_type,
			a.salary,
			a.position_description,
			a.tech_stack,
			a.created_at,
			a.updated_at,
			sh.status,
			sh.changed_at,
			count(*) over()
		from %s a
		join lateral (
			select status, changed_at
			from %s
			where application_id = a.id
			order by changed_at desc, id desc
			limit 1
		) sh on true
		order by a.created_at desc, a.id desc
		limit $1 offset $2
	`

	return fmt.Sprintf(query, r.applicationsTable(), r.statusHistoryTable()), []any{limit, input.Offset}
}

func scanApplicationSummary(row interface {
	Scan(dest ...any) error
}) (app.ApplicationSummary, int, error) {
	var (
		item                app.ApplicationSummary
		positionTitle       *string
		sourceURL           *string
		workType            *string
		salary              *string
		positionDescription *string
		techStack           *string
		status              string
		total               int
	)

	if err := row.Scan(
		&item.ID,
		&item.CompanyName,
		&positionTitle,
		&sourceURL,
		&workType,
		&salary,
		&positionDescription,
		&techStack,
		&item.CreatedAt,
		&item.UpdatedAt,
		&status,
		&item.LastStatusChangedAt,
		&total,
	); err != nil {
		return app.ApplicationSummary{}, 0, err
	}

	item.PositionTitle = derefString(positionTitle)
	item.SourceURL = derefString(sourceURL)
	item.WorkType = derefString(workType)
	item.Salary = derefString(salary)
	item.PositionDescription = derefString(positionDescription)
	item.TechStack = derefString(techStack)
	item.CurrentStatus = app.ApplicationStatus(status)

	return item, total, nil
}
