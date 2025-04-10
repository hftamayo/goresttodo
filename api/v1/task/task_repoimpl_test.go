package task

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func TestTaskRepositoryImpl_Create(t *testing.T) {
    helper := testutils.NewTestHelper(t)

    db, mock, err := testutils.SetupTestDB()
    helper.AssertNoError(err)

    repo := &TaskRepositoryImpl{Db: db}

    task := &models.Task{
        Title:       "Test Task",
        Description: "Test Description",
        Done:        false,
        Owner:       1,
    }

    mock.ExpectBegin()
    mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "tasks" ("created_at","updated_at","deleted_at","title","description","done","owner") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`)).
        WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, task.Title, task.Description, task.Done, task.Owner).
        WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
    mock.ExpectCommit()

    err = repo.Create(task)
    helper.AssertNoError(err)

    err = mock.ExpectationsWereMet()
    helper.AssertNoError(err)
}

func TestTaskRepositoryImpl_List(t *testing.T) {
    helper := testutils.NewTestHelper(t)
    db, mock, err := testutils.SetupTestDB()
    helper.AssertNoError(err)

    repo := &TaskRepositoryImpl{Db: db}

    rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "title", "description", "done", "owner"}).
        AddRow(1, time.Now(), time.Now(), nil, "Task 1", "Description 1", false, 1).
        AddRow(2, time.Now(), time.Now(), nil, "Task 2", "Description 2", true, 1)

    mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tasks" WHERE "tasks"."deleted_at" IS NULL LIMIT 10 OFFSET 0`)).
        WillReturnRows(rows)

    tasks, err := repo.List(1, 10)
    helper.AssertNoError(err)
    assert.Len(t, tasks, 2)

    err = mock.ExpectationsWereMet()
    helper.AssertNoError(err)
}

func TestTaskRepositoryImpl_ListById(t *testing.T) {
    helper := testutils.NewTestHelper(t)
    db, mock, err := testutils.SetupTestDB()
    helper.AssertNoError(err)

    repo := &TaskRepositoryImpl{Db: db}

    rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "title", "description", "done", "owner"}).
        AddRow(1, time.Now(), time.Now(), nil, "Task 1", "Description 1", false, 1)

    mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tasks" WHERE "tasks"."deleted_at" IS NULL AND "tasks"."id" = $1 ORDER BY "tasks"."id" LIMIT 1`)).
        WithArgs(1).
        WillReturnRows(rows)

    task, err := repo.ListById(1)
    helper.AssertNoError(err)
    assert.NotNil(t, task)
    assert.Equal(t, "Task 1", task.Title)

    err = mock.ExpectationsWereMet()
    helper.AssertNoError(err)
}