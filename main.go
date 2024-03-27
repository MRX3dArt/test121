package main

import (
	"fmt"
	"sync"
	"time"
)

type Ttype struct {
	id         int
	cT         string
	fT         string
	taskRESULT []byte
}

func taskCreturer(a chan Ttype, wg *sync.WaitGroup) {
	defer wg.Done()
	go func() {
		for {
			ft := time.Now().Format(time.RFC3339)
			if time.Now().Nanosecond()%2 > 0 {
				ft = "Some error occurred"
			}
			a <- Ttype{cT: ft, id: int(time.Now().Unix())}
		}
	}()
}

func taskWorker(a Ttype) Ttype {
	tt, _ := time.Parse(time.RFC3339, a.cT)
	if tt.After(time.Now().Add(-20 * time.Second)) {
		a.taskRESULT = []byte("task has been successed")
	} else {
		a.taskRESULT = []byte("something went wrong")
	}
	a.fT = time.Now().Format(time.RFC3339Nano)
	time.Sleep(time.Millisecond * 150)
	return a
}

func main() {
	superChan := make(chan Ttype, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go taskCreturer(superChan, &wg)

	doneTasks := make([]Ttype, 0)
	undoneTasks := make([]error, 0)

	tasksorter := func(t Ttype, wg *sync.WaitGroup) {
		defer wg.Done()
		if string(t.taskRESULT)[:14] == "task has been" {
			doneTasks = append(doneTasks, t)
		} else {
			undoneTasks = append(undoneTasks, fmt.Errorf("Task id %d time %s, error %s", t.id, t.cT, t.taskRESULT))
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for t := range superChan {
			t = taskWorker(t)
			wg.Add(1)
			go tasksorter(t, &wg)
		}
	}()

	go func() {
		wg.Wait()
		close(superChan)
	}()

	wg.Wait()

	fmt.Println("Errors:")
	for _, err := range undoneTasks {
		fmt.Println(err)
	}

	fmt.Println("Done tasks:")
	for _, t := range doneTasks {
		fmt.Println(t.id)
	}
}
