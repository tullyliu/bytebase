package store_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	"github.com/bytebase/bytebase/backend/store"
)

func TestUpdateTaskRunStatusFailedWithNULDetail(t *testing.T) {
	t.Parallel()
	fixture := newStorePostgresFixture(t, `
		INSERT INTO instance (resource_id, workspace) VALUES ('instance-a', 'default');
		INSERT INTO plan (id, creator, project, name, description)
			VALUES (101, 'creator@example.com', 'project-a', 'Plan A', '');
		INSERT INTO task (id, project, plan_id, instance, type)
			VALUES (101, 'project-a', 101, 'instance-a', 'DATABASE_SCHEMA_UPDATE');
		INSERT INTO task_run (id, project, task_id, attempt, status)
			VALUES (101, 'project-a', 101, 0, 'RUNNING');
	`)

	_, err := fixture.store.UpdateTaskRunStatus(fixture.ctx, &store.TaskRunStatusPatch{
		ID:              101,
		ProjectID:       "project-a",
		Status:          storepb.TaskRun_FAILED,
		AllowedStatuses: []storepb.TaskRun_Status{storepb.TaskRun_RUNNING},
		ResultProto:     &storepb.TaskRunResult{Detail: "unsigned expr not supported.\x00\xff"},
	})
	require.NoError(t, err)

	var status, detail string
	require.NoError(t, fixture.db.QueryRowContext(fixture.ctx,
		"SELECT status, result->>'detail' FROM task_run WHERE project = 'project-a' AND id = 101",
	).Scan(&status, &detail))
	require.Equal(t, "FAILED", status)
	require.Equal(t, "unsigned expr not supported.�", detail)
}
