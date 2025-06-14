package application

import (
	"fmt"
	"slices"
	"sync"
	"time"
)

type CronJob func(app *App)

// Cron is a "dumb" cron-like scheduler that runs jobs at a specified frequency.
// It does not support complex scheduling like specific times or intervals.
// It simply runs all registered jobs every `frequency` seconds, and it does not care about drift or missed executions.
type Cron struct {
	app *App

	jobsMutex sync.Mutex
	jobs      []CronJob

	frequency  int // Frequency in seconds
	exitCh     chan bool
	exitDoneCh chan bool
}

func NewCron(app *App, frequency int) *Cron {
	return &Cron{
		app:        app,
		frequency:  frequency,
		exitCh:     make(chan bool, 1),
		exitDoneCh: make(chan bool, 1),
	}
}

func (c *Cron) AddJob(job CronJob) {
	c.jobsMutex.Lock()
	defer c.jobsMutex.Unlock()

	c.jobs = append(c.jobs, job)
}

func (c *Cron) Stop() {
	c.exitCh <- true
	close(c.exitCh)

	<-c.exitDoneCh

	c.jobsMutex.Lock()
	c.jobs = nil
	c.jobsMutex.Unlock()
}

func (c *Cron) Start() {
	go func() {
		fmt.Println("Cron start")
		timer := time.NewTimer(time.Second * time.Duration(c.frequency))

	loop:
		for {
			select {
			case <-c.exitCh:
				break loop
			case <-timer.C:
				fmt.Println("Cron tick")
				c.jobsMutex.Lock()
				jobsCopy := slices.Clone(c.jobs)
				c.jobsMutex.Unlock()

				for _, job := range jobsCopy {
					job(c.app)
				}

				timer.Reset(time.Second * time.Duration(c.frequency))
			}
		}

		timer.Stop()

		c.exitDoneCh <- true
		close(c.exitDoneCh)
	}()
}
