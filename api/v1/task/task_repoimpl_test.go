package task

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, error) {
    // Create a new SQL mock
    sqlDB, mock, err := sqlmock.New()
    if err != nil {
        return nil, nil, err
    }

    dialector := postgres.New(postgres.Config{
        Conn:       sqlDB,
        DriverName: "postgres",
    })

    db, err := gorm.Open(dialector, &gorm.Config{})
    if err != nil {
        return nil, nil, err
    }

    return db, mock, nil
}

func TestTaskRepositoryImpl_Create(t *testing.T) {
    // Initialize test helper
    helper := testutils.NewTestHelper(t)

    // Setup test database
    db, mock, err := setupTestDB(t)
    helper.AssertNoError(err)

    // Create repository instance
    repo := &TaskRepositoryImpl{Db: db}

    // Test case
    task := &models.Task{
        Title:       "Test Task",
        Description: "Test Description",
        Done:        false,
        Owner:       1,
    }

    // Setup mock expectations
    mock.ExpectBegin()
    mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "tasks" ("created_at","updated_at","deleted_at","title","description","done","owner") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`)).
        WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, task.Title, task.Description, task.Done, task.Owner).
        WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
    mock.ExpectCommit()

    // Execute test
    err = repo.Create(task)
    helper.AssertNoError(err)

    // Verify all expectations were met
    err = mock.ExpectationsWereMet()
    helper.AssertNoError(err)
}

func TestTaskRepositoryImpl_List(t *testing.T) {
    helper := testutils.NewTestHelper(t)
    db, mock, err := setupTestDB(t)
    helper.AssertNoError(err)

    repo := &TaskRepositoryImpl{Db: db}

    // Setup mock data
    rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "title", "description", "done", "owner"}).
        AddRow(1, time.Now(), time.Now(), nil, "Task 1", "Description 1", false, 1).
        AddRow(2, time.Now(), time.Now(), nil, "Task 2", "Description 2", true, 1)

    // Setup mock expectations
    mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tasks" WHERE "tasks"."deleted_at" IS NULL LIMIT 10 OFFSET 0`)).
        WillReturnRows(rows)

    // Execute test
    tasks, err := repo.List(1, 10)
    helper.AssertNoError(err)
    assert.Len(t, tasks, 2)

    err = mock.ExpectationsWereMet()
    helper.AssertNoError(err)
}

func TestTaskRepositoryImpl_ListById(t *testing.T) {
    helper := testutils.NewTestHelper(t)
    db, mock, err := setupTestDB(t)
    helper.AssertNoError(err)

    repo := &TaskRepositoryImpl{Db: db}

    // Setup mock data
    rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "title", "description", "done", "owner"}).
        AddRow(1, time.Now(), time.Now(), nil, "Task 1", "Description 1", false, 1)

    // Setup mock expectations
    mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tasks" WHERE "tasks"."deleted_at" IS NULL AND "tasks"."id" = $1 ORDER BY "tasks"."id" LIMIT 1`)).
        WithArgs(1).
        WillReturnRows(rows)

    // Execute test
    task, err := repo.ListById(1)
    helper.AssertNoError(err)
    assert.NotNil(t, task)
    assert.Equal(t, "Task 1", task.Title)

    err = mock.ExpectationsWereMet()
    helper.AssertNoError(err)
}