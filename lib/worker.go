package spsw

import (
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Worker struct {
	UUID             string
	ScheduledTasksIn chan *ScheduledTask
	DataChunksOut    chan *DataChunk
	Done             chan interface{}
}

func NewWorker() *Worker {
	return &Worker{
		UUID:             uuid.New().String(),
		ScheduledTasksIn: make(chan *ScheduledTask),
		DataChunksOut:    make(chan *DataChunk),
		Done:             make(chan interface{}),
	}
}

func (w *Worker) executeTask(task *Task) error {
	err := task.Run()
	if err != nil {
		log.Error(fmt.Sprintf("Task %v failed with error: %v", task, err))
		// TODO: send error
		return err
	}

	for _, outDP := range task.Outputs {
		if len(outDP.Queue) == 0 {
			continue
		}

		x := outDP.Remove()

		if item, okItem := x.(*Item); okItem {
			chunk, _ := NewDataChunk(item)
			w.DataChunksOut <- chunk
		}

		if promise, okPromise := x.(*TaskPromise); okPromise {
			chunk, _ := NewDataChunk(promise)
			w.DataChunksOut <- chunk
		}
	}

	return nil
}

func (w *Worker) Run() error {
	log.Info(fmt.Printf("Starting runloop for worker %s", w.UUID))

	for {
		select {
		case scheduledTask := <-w.ScheduledTasksIn:
			log.Info(fmt.Printf("Worker %s got scheduled task %v", w.UUID, scheduledTask))

			task := NewTaskFromScheduledTask(scheduledTask)
			log.Info(fmt.Sprintf("Worker %s running task %v", w.UUID, task))

			w.executeTask(task)
		case <-w.Done:
			return nil
		}
	}

}
