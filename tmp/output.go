 Here's an example of tests for the new code. I've copied the iteration over the flagTestCreateHistory array for your convenience.

                     ```go
                     package worker_test

                     import (
                         "context"
                         "testing"

                         "github.com/google/go-cmp/cmp"
                         "github.com/google/go-cmp/cmp/cmpassert"

                         "github.com/stretchr/testify/mock"

                         "go.uber.org/zap"

                         "time"

                         "uservice-worker/internal/entities"
                         "uservice-worker/internal/mocks"
                         "uservice-worker/internal/services"
                     )

                     func TestWorker_PrepareNextTaskToRun(t *testing.T) {
                         ctx := context.Background()

                         var (
                             task = &entities.Task{
                                 ID:         1,
                                 Name:       "test_task",
                                 Plan:       &entities.Plan{},
                                 Step:       &entities.Step{},
                                 WorkflowID: 1,
                             }
                             workerUUID = uuid.New()
                             errDB = errors.New("failed to get next task")
                             errDB2 = errors.New("failed to execute task")
                             errJQ = errors.New("failed to process jq")
                             errJQ2 = errors.New("failed to marshal jq output")
                             errExec = errors.New("failed to execute task finalizer")
                             errExec2 = errors.New("failed to execute task starter")
                             errHistory = errors.New("failed to check task history")
                             errHistory2 = errors.New("failed to get task execution history")
                             errGetDependencies = errors.New("failed to get task dependencies")
                             errUnmarshal = errors.New("failed to unmarshal task name")
                         )

                         flagTestPrepareNextTaskToRun := []struct {
                             name               string
                             deps               *services.WorkerPendingTaskGetterDependencies
                             shouldCallGetNextTask bool
                             resGetNextTask      *entities.Task
                             shouldCallIsTaskDuplicated bool
                             resIsTaskDuplicated bool
                             shouldCallExec      bool
                             resExec             error
                             expectedErr         error
                         }{
                             {
                                 name: "ok",

                                 deps: &services.WorkerPendingTaskGetterDependencies{
                                     TaskDAO: &mocks.MockTaskDAO{},
                                     TaskStarter: &mocks.MockTaskStarter{},
                                     TaskFinalizer: &mocks.MockTaskFinalizer{},
                                     TaskExecutionHistory: &mocks.MockTaskExecutionHistory{},
                                 },

                                 shouldCallGetNextTask: true,
                                 resGetNextTask: task,

                                 shouldCallIsTaskDuplicated: false,
                                 resIsTaskDuplicated: false,

                                 shouldCallExec: true,
                                 resExec: nil,
                             },
                             {
                                 name: "err get next task",

                                 deps: &services.WorkerPendingTaskGetterDependencies{
                                     TaskDAO: &mocks.MockTaskDAO{},
                                     TaskStarter: &mocks.MockTaskStarter{},
                                     TaskFinalizer: &mocks.MockTaskFinalizer{},
                                     TaskExecutionHistory: &mocks.MockTaskExecutionHistory{},
                                 },

                                 shouldCallGetNextTask: true,
                                 resGetNextTask: nil,

                                 shouldCallIsTaskDuplicated: false,
                                 resIsTaskDuplicated: false,

                                 shouldCallExec: false,
                                 expectedErr: errDB,
                             },
                             {
                                 name: "err task duplicated",

                                 deps: &services.WorkerPendingTaskGetterDependencies{
                                     TaskDAO: &mocks.MockTaskDAO{},
                                     TaskStarter: &mocks.MockTaskStarter{},
                                     TaskFinalizer: &mocks.MockTaskFinalizer{},
                                     TaskExecutionHistory: &mocks.MockTaskExecutionHistory{},
                                 },

                                 shouldCallGetNextTask: true,
                                 resGetNextTask: task,

                                 shouldCallIsTaskDuplicated: true,
                                 resIsTaskDuplicated: true,

                                 shouldCallExec: false,
                                 expectedErr: ErrTaskDuplicated,
                             },
                             {
                                 name: "err execute task",

                                 deps: &services.WorkerPendingTaskGetterDependencies{
                                     TaskDAO: &mocks.MockTaskDAO{},
                                     TaskStarter: &mocks.MockTaskStarter{},
                                     TaskFinalizer: &mocks.MockTaskFinalizer{},
                                     TaskExecutionHistory: &mocks.MockTaskExecutionHistory{},
                                 },

                                 shouldCallGetNextTask: true,
                                 resGetNextTask: task,

                                 shouldCallIsTaskDuplicated: false,
                                 resIsTaskDuplicated: false,

                                 shouldCallExec: true,
                                 resExec: nil,

                                 expectedErr: errExec,
                             },
                             {
                                 name: "err execute task finalizer",

                                 deps: &services.WorkerPendingTaskGetterDependencies{
                                     TaskDAO: &mocks.MockTaskDAO{},
                                     TaskStarter: &mocks.MockTaskStarter{},
                                     TaskFinalizer: &mocks.MockTaskFinalizer{},
                                     TaskExecutionHistory: &mocks.MockTaskExecutionHistory{},
                                 },

                                 shouldCallGetNextTask: true,
                                 resGetNextTask: task,

                                 shouldCallIsTaskDuplicated: false,
                                 resIsTaskDuplicated: false,

                                 shouldCallExec: true,
                                 resExec: nil,

                                 expectedErr: errExec2,
                             },
                             {
                                 name: "err get task dependencies",

                                 deps: &services.WorkerPendingTaskGetterDependencies{
                                     TaskDAO: &mocks.MockTaskDAO{},
                                     TaskStarter: &mocks.MockTaskStarter{},
                                     TaskFinalizer: &mocks.MockTaskFinalizer{},
                                     TaskExecutionHistory: &mocks.MockTaskExecutionHistory{},
                                 },

                                 shouldCallGetNextTask: true,
                                 resGetNextTask: task,

                                 shouldCallIsTaskDuplicated: false,
                                 resIsTaskDuplicated: false,

                                 shouldCallExec: false,
                                 expectedErr: errGetDependencies,
                             },
                             {
                                 name: "err unmarshal task name",

                                 deps: &services.WorkerPendingTaskGetterDependencies{
                                     TaskDAO: &mocks.MockTaskDAO{},
                                     TaskStarter: &mocks.MockTaskStarter{},
                                     TaskFinalizer: &mocks.MockTaskFinalizer{},
                                     TaskExecutionHistory: &mocks.MockTaskExecutionHistory{},
                                 },

                                 shouldCallGetNextTask: true,
                                 resGetNextTask: task,

                                 shouldCallIsTaskDuplicated: false,
                                 resIsTaskDuplicated: false,

                                 shouldCallExec: false,
                                 expectedErr: errUnmarshal,
                             },
                             {
                                 name: "err check task history",

                                 deps: &services.WorkerPendingTaskGetterDependencies{
                                     TaskDAO: &mocks.MockTaskDAO{},
                                     TaskStarter: &mocks.MockTaskStarter{},
                                     TaskFinalizer: &mocks.MockTaskFinalizer{},
                                     TaskExecutionHistory: &mocks.MockTaskExecutionHistory{},
                                 },

                                 shouldCallGetNextTask: true,
                                 resGetNextTask: task,

                                 shouldCallIsTaskDuplicated: false,
                                 resIsTaskDuplicated: false,

                                 shouldCallExec: false,
                                 expectedErr: errHistory,
                             },
                             {
                                 name: "err get task execution history",

                                 deps: &services.WorkerPendingTaskGetterDependencies{
                                     TaskDAO: &mocks.MockTaskDAO{},
                                     TaskStarter: &mocks.MockTaskStarter{},
                                     TaskFinalizer: &mocks.MockTaskFinalizer{},
                                     TaskExecutionHistory: &mocks.MockTaskExecutionHistory{},
                                 },

                                 shouldCallGetNextTask: true,
                                 resGetNextTask: task,

                                 shouldCallIsTaskDuplicated: false,
                                 resIsTaskDuplicated: false,

                                 shouldCallExec: false,
                                 expectedErr: errHistory2,
                             },
                         }

                         for _, tt := range flagTestPrepareNextTaskToRun {
                             tt := tt

                             t.Run(tt.name, func(t *testing.T) {
                                 assert := cmpassert.New(t)
                                 dbMock := new(mocks.MockTaskDAO)
                                 starterMock := new(mocks.MockTaskStarter)
                                 finalizerMock := new(mocks.MockTaskFinalizer)
                                 executionHistoryMock := new(mocks.MockTaskExecutionHistory)

                                 if tt.shouldCallGetNextTask {
                                     dbMock.On("GetNextPendingTask", ctx).Return(tt.resGetNextTask, nil)
                                 } else {
                                     dbMock.On("GetNextPendingTask", ctx).Return(nil, tt.expectedErr)
                                 }

                                 starterMock.On("Exec", ctx, tt.resGetNextTask.ID, workerUUID).Return(tt.resExec)
                                 finalizerMock.On("Exec", ctx, tt.resGetNextTask.ID, entities.TaskStatusCancelled, mock.Anything).Return(nil)
                                 finalizerMock.On("Exec", ctx, tt.resGetNextTask.ID, entities.TaskStatusDuplicated, mock.Anything).Return(nil)
                                 executionHistoryMock.On("Contains", ctx, "workflow_name", "task_name", mock.Anything).Return(tt.resIsTaskDuplicated, nil)
                                 executionHistoryMock.On("Contains", ctx, "workflow_name", "task_name", mock.Anything).Return(tt.resIsTaskDuplicated, tt.expectedErr)

                                 deps := tt.deps
                                 deps.TaskDAO = dbMock
                                 deps.TaskStarter = starterMock
                                 deps.TaskFinalizer = finalizerMock
                                 deps.TaskExecutionHistory = executionHistoryMock

                                 worker := services.New()

                                 gotTask, err := worker.PrepareNextTaskToRun(ctx, workerUUID)

                                 if tt.expectedErr != nil {
                                     assert.ErrorIs(err, tt.expectedErr)
                                 } else {
                                     assert.Equal(gotTask, tt.resGetNextTask)
                                 }

                                 dbMock.AssertExpectations(t)
                                 starterMock.AssertExpectations(t)
                                 finalizerMock.AssertExpectations(t)
                                 executionHistoryMock.AssertExpectations(t)
                             })
                         }
                     }

                     func TestWorker_isTaskDuplicated(t *testing.T) {
                         ctx := context.Background()

                         var (
                             task = &entities.Task{
                                 ID:         1,
                                 Name:       "test_task",
                                 Plan:       &entities.Plan{},
                                 Step:       &entities.Step{},
                                 WorkflowID: 1,
                             }
                             workerUUID = uuid.New()
                             errDB = errors.New("failed to get task execution history")
                             errDB2 = errors.New("failed to check task execution history")
                             errUnmarshal = errors.New("failed to unmarshal task name")
                         )

                         flagTestIsTaskDuplicated := []struct {
                             name               string
                             deps               *services.WorkerPendingTaskGetterDependencies
                             task               *entities.Task
                             expectedIsDuplicated bool
                             expectedErr         error
                         }{
                             {
                                 name: "ok",

                                 deps: &services.WorkerPendingTaskGetterDependencies{
                                     TaskDAO: &mocks.MockTaskDAO{},
                                     TaskStarter: &mocks.MockTaskStarter{},
                                     TaskFinalizer: &mocks.MockTaskFinalizer{},
                                     TaskExecutionHistory: &mocks.MockTaskExecutionHistory{},
                                 },

                                 task: task,

                                 expectedIsDuplicated: false,
                                 expectedErr: nil,
                             },
                             {
                                 name: "err get task execution history",

                                 deps: &services.WorkerPendingTaskGetterDependencies{
                                     TaskDAO: &mocks.MockTaskDAO{},
                                     TaskStarter: &mocks.MockTaskStarter{},
                                     TaskFinalizer: &mocks.MockTaskFinalizer{},
                                     TaskExecutionHistory: &mocks.MockTaskExecutionHistory{},
                                 },

                                 task: task,

                                 expectedIsDuplicated: false,
                                 expectedErr: errDB,
                             },
                             {
                                 name: "err check task execution history",

                                 deps: &services.WorkerPendingTaskGetterDependencies{
                                     TaskDAO: &mocks.MockTaskDAO{},
                                     TaskStarter: &mocks.MockTaskStarter{},
                                     TaskFinalizer: &mocks.MockTaskFinalizer{},
                                     TaskExecutionHistory: &mocks.MockTaskExecutionHistory{},
                                 },

                                 task: task,

                                 expectedIsDuplicated: false,
                                 expectedErr: errDB2,
                             },
                         }

                         for _, tt := range flagTestIsTaskDuplicated {
                             tt := tt

                             t.Run(tt.name, func(t *testing.T) {
                                 assert := cmpassert.New(t)
                                 dbMock := new(mocks.MockTaskExecutionHistory)

                                 deps := tt.deps
                                 deps.TaskExecutionHistory = dbMock

                                 worker := services.New()

                                 gotIsDuplicated, err := worker.isTaskDuplicated(ctx, deps, task)

                                 if tt.expectedErr != nil {
                                     assert.ErrorIs(err, tt.expectedErr)
                                 } else {
                                     assert.Equal(gotIsDuplicated, tt.expectedIsDuplicated)
                                 }

                                 dbMock.AssertExpectations(t)
                             })
                         }
                     }
                     ```

                     This test suite covers the `PrepareNextTaskToRun` and `isTaskDuplicated` methods of the `Worker` struct. It includes tests for various error cases and happy paths. The tests make use of mock dependencies to isolate the code under test and assert the expected behavior.