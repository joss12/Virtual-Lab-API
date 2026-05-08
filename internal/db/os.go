package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Models

type OsCourse struct {
	ID            pgtype.UUID `json:"id"`
	Slug          string      `json:"slug"`
	Track         string      `json:"track"`
	TitleEn       string      `json:"title_en"`
	TitleFr       string      `json:"title_fr"`
	DescriptionEn string      `json:"description_en"`
	DescriptionFr string      `json:"description_fr"`
	OrderIndex    int         `json:"order_index"`
	CreatedAt     time.Time   `json:"created_at"`
}

type OsLesson struct {
	ID          pgtype.UUID `json:"id"`
	CourseSlug  string      `json:"course_slug"`
	Slug        string      `json:"slug"`
	TitleEn     string      `json:"title_en"`
	TitleFr     string      `json:"title_fr"`
	OrderIndex  int         `json:"order_index"`
	HasTerminal bool        `json:"has_terminal"`
	HasQuiz     bool        `json:"has_quiz"`
	CreatedAt   time.Time   `json:"created_at"`
}

type OsProgress struct {
	ID         pgtype.UUID `json:"id"`
	UserID     pgtype.UUID `json:"user_id"`
	LessonSlug string      `json:"lesson_slug"`
	Completed  bool        `json:"completed"`
	QuizScore  *int        `json:"quiz_score"`
	QuizTotal  *int        `json:"quiz_total"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type OsQuizQuestion struct {
	ID           pgtype.UUID `json:"id"`
	LessonSlug   string      `json:"lesson_slug"`
	QuestionEn   string      `json:"question_en"`
	QuestionFr   string      `json:"question_fr"`
	OptionsEn    []string    `json:"options_en"`
	OptionsFr    []string    `json:"options_fr"`
	CorrectIndex int         `json:"correct_index"`
	OrderIndex   int         `json:"order_index"`
}

// Queries

func (q *Queries) GetOsCourses(ctx context.Context) ([]OsCourse, error) {
	rows, err := q.db.Query(ctx, `
		SELECT id, slug, track, title_en, title_fr, description_en, description_fr, order_index, created_at
		FROM os_courses
		ORDER BY order_index ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []OsCourse
	for rows.Next() {
		var c OsCourse
		if err := rows.Scan(&c.ID, &c.Slug, &c.Track, &c.TitleEn, &c.TitleFr, &c.DescriptionEn, &c.DescriptionFr, &c.OrderIndex, &c.CreatedAt); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	if courses == nil {
		courses = []OsCourse{}
	}
	return courses, nil
}

func (q *Queries) GetOsLessonsByCourse(ctx context.Context, courseSlug string) ([]OsLesson, error) {
	rows, err := q.db.Query(ctx, `
		SELECT id, course_slug, slug, title_en, title_fr, order_index, has_terminal, has_quiz, created_at
		FROM os_lessons
		WHERE course_slug = $1
		ORDER BY order_index ASC
	`, courseSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lessons []OsLesson
	for rows.Next() {
		var l OsLesson
		if err := rows.Scan(&l.ID, &l.CourseSlug, &l.Slug, &l.TitleEn, &l.TitleFr, &l.OrderIndex, &l.HasTerminal, &l.HasQuiz, &l.CreatedAt); err != nil {
			return nil, err
		}
		lessons = append(lessons, l)
	}
	if lessons == nil {
		lessons = []OsLesson{}
	}
	return lessons, nil
}

func (q *Queries) GetOsLesson(ctx context.Context, slug string) (OsLesson, error) {
	row := q.db.QueryRow(ctx, `
		SELECT id, course_slug, slug, title_en, title_fr, order_index, has_terminal, has_quiz, created_at
		FROM os_lessons
		WHERE slug = $1
		LIMIT 1
	`, slug)
	var l OsLesson
	err := row.Scan(&l.ID, &l.CourseSlug, &l.Slug, &l.TitleEn, &l.TitleFr, &l.OrderIndex, &l.HasTerminal, &l.HasQuiz, &l.CreatedAt)
	return l, err
}

func (q *Queries) GetOsQuizQuestions(ctx context.Context, lessonSlug string) ([]OsQuizQuestion, error) {
	rows, err := q.db.Query(ctx, `
		SELECT id, lesson_slug, question_en, question_fr, options_en, options_fr, correct_index, order_index
		FROM os_quiz_questions
		WHERE lesson_slug = $1
		ORDER BY order_index ASC
	`, lessonSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []OsQuizQuestion
	for rows.Next() {
		var q OsQuizQuestion
		if err := rows.Scan(&q.ID, &q.LessonSlug, &q.QuestionEn, &q.QuestionFr, &q.OptionsEn, &q.OptionsFr, &q.CorrectIndex, &q.OrderIndex); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	if questions == nil {
		questions = []OsQuizQuestion{}
	}
	return questions, nil
}

func (q *Queries) GetOsProgress(ctx context.Context, userID pgtype.UUID) ([]OsProgress, error) {
	rows, err := q.db.Query(ctx, `
		SELECT id, user_id, lesson_slug, completed, quiz_score, quiz_total, updated_at
		FROM os_progress
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var progress []OsProgress
	for rows.Next() {
		var p OsProgress
		if err := rows.Scan(&p.ID, &p.UserID, &p.LessonSlug, &p.Completed, &p.QuizScore, &p.QuizTotal, &p.UpdatedAt); err != nil {
			return nil, err
		}
		progress = append(progress, p)
	}
	if progress == nil {
		progress = []OsProgress{}
	}
	return progress, nil
}

func (q *Queries) UpsertOsProgress(ctx context.Context, userID pgtype.UUID, lessonSlug string, completed bool, quizScore *int, quizTotal *int) error {
	_, err := q.db.Exec(ctx, `
		INSERT INTO os_progress (user_id, lesson_slug, completed, quiz_score, quiz_total, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (user_id, lesson_slug)
		DO UPDATE SET completed = $3, quiz_score = $4, quiz_total = $5, updated_at = NOW()
	`, userID, lessonSlug, completed, quizScore, quizTotal)
	return err
}
