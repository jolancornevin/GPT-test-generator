
type PgxConn interface {
	Rollback(ctx context.Context) error
	Commit(ctx context.Context) error
}

type TaskDAOForWorkerTaskStarter interface {
	GetNextPendingTask(ctx context.Context) (*entities.Task, error)
}

type TaskExecutionHistoryDAOForWorkerTaskStarter interface {
	Contains(ctx context.Context, workflow_name, task_name, unique_tasks_params string) (bool, error)
}

type TaskFinalizerForWorkerPendingTaskGetter interface {
	Exec(ctx context.Context, ID int64, status entities.TaskStatus, output interface{}) error
}

type TaskStarter interface {
	Exec(ctx context.Context, ID int64, workerUUID uuid.UUID) error
}

type WorkerPendingTaskGetterDependencies struct {
	Conn                 PgxConn
	TaskDAO              TaskDAOForWorkerTaskStarter
	TaskStarter          TaskStarter
	TaskFinalizer        TaskFinalizerForWorkerPendingTaskGetter
	TaskExecutionHistory TaskExecutionHistoryDAOForWorkerTaskStarter
}

func NewWorkerPendingTaskGetterDependencies(
	conn PgxConn,
	taskDAO TaskDAOForWorkerTaskStarter,
	taskFinalizer TaskFinalizerForWorkerPendingTaskGetter,
	taskStarter TaskStarter,
	taskExecutionHistory TaskExecutionHistoryDAOForWorkerTaskStarter,
) *WorkerPendingTaskGetterDependencies {
	return &WorkerPendingTaskGetterDependencies{
		Conn:                 conn,
		TaskStarter:          taskStarter,
		TaskFinalizer:        taskFinalizer,
		TaskDAO:              taskDAO,
		TaskExecutionHistory: taskExecutionHistory,
	}
}

func InitializeWorkerPendingTaskGetterDependencies(ctx context.Context) (*WorkerPendingTaskGetterDependencies, error) {
	return &WorkerPendingTaskGetterDependencies{}, nil
}

type Worker struct {
	PrepareNextTaskToRunDeps func(ctx context.Context) (*di.WorkerPendingTaskGetterDependencies, error)
}

func New() *Worker {
	return &Worker{
		PrepareNextTaskToRunDeps: di.InitializeWorkerPendingTaskGetterDependencies,
	}
}

// PrepareNextTaskToRun gets the next task to run and sets its status to RUNNING
func (w *Worker) PrepareNextTaskToRun(ctx context.Context, workerUUID uuid.UUID) (*entities.Task, error) {
	deps, err := w.PrepareNextTaskToRunDeps(ctx)
	if err != nil {
		panic(err)
	}

	defer func() {
		database.CleanTx(ctx, deps.Conn, err)
	}()

	task, err := deps.TaskDAO.GetNextPendingTask(ctx)
	if err != nil {
		return nil, errors.Wrap(err, ErrGetNextPendingTaskFailed)
	}

	task.WorkerUUID = &workerUUID

	// if the task upper plan is out-dated, the task should be cancelled.
	if task.Plan.Upper.Valid && task.Plan.Upper.Time.Before(time.Now()) {
		output := map[string]interface{}{
			"error": "task upper plan is outdated",
		}
		if err := deps.TaskFinalizer.Exec(ctx, task.ID, entities.TaskStatusCancelled, output); err != nil {
			return nil, errors.Wrap(err, ErrGetNextPendingTaskFailed)
		}

		return nil, errors.Wrap(ErrTaskOutdated, ErrGetNextPendingTaskFailed)
	}

	// if another task with the same name and parameters already ran, we cancel it too.
	taskAlreadyRan, err := w.isTaskDuplicated(ctx, deps, task)
	if err != nil {
		return nil, errors.Wrap(err, ErrGetTaskHistory)
	}

	if taskAlreadyRan {
		output := map[string]interface{}{
			"error": "task already ran",
		}
		if err := deps.TaskFinalizer.Exec(ctx, task.ID, entities.TaskStatusDuplicated, output); err != nil {
			return nil, errors.Wrap(err, ErrGetNextPendingTaskFailed)
		}

		return nil, errors.Wrap(ErrTaskDuplicated, ErrGetNextPendingTaskFailed)
	}

	err = deps.TaskStarter.Exec(ctx, task.ID, workerUUID)
	if err != nil {
		return nil, errors.Wrap(err, ErrStartTaskFailed)
	}

	return task, nil
}
func (w *Worker) isTaskDuplicated(ctx context.Context, deps *di.WorkerPendingTaskGetterDependencies, task *entities.Task) (bool, error) {
	deduplicatingValues, _ := w.computeJQForTaskDuplication(ctx, task)
	workflowName, taskName, _ := entities.UnmarshalTaskName(task.Name)
	return deps.TaskExecutionHistory.Contains(ctx, workflowName, taskName, deduplicatingValues)
}
func (w *Worker) computeJQForTaskDuplication(ctx context.Context, referenceTask *entities.Task) (string, error) {
	task, err := w.GetTaskDependencies(ctx, referenceTask)
	if err != nil {
		return "", errors.Annotate(err, "getting task dependencies while checking task duplicated")
	}

	input, err := task.Environment(ctx)
	if err != nil {
		return "", errors.Annotate(err, "get environment while checking task duplicated")
	}

	jq := string(task.Step.Deduplication)
	// if we have no instruction to de-duplicate the task, we skip
	if jq == "" {
		return "", nil
	}

	query, err := gojq.Parse(jq)
	if err != nil {
		return "", errors.Annotate(err, "parse parameters while checking task duplicated")
	}

	deduplicatingValues, ok := query.RunWithContext(ctx, input).Next()
	// if the jq yielded no instruction to de-duplicate the task, we skip
	if !ok {
		return "", nil
	}

	if err, ok := deduplicatingValues.(error); ok {
		return "", errors.Annotate(err, "failed to process jq with context for task duplicated")
	}

	deduplicatingValuesBytes, err := json.Marshal(deduplicatingValues)
	if err != nil {
		return "", errors.Annotate(err, "failed to convert the jq into json while checking task duplicated")
	}

	return string(deduplicatingValuesBytes), nil
}
