package gpool

import (
	"log"
	"sync"
)

var defaultPool = &Pool{
	MaxIdleWorker: 1000,
}

func Run(fn task) {
	defaultPool.Run(fn)
}

type Pool struct {
	idleMu     sync.Mutex
	idleWorker []*worker
	wantQueue  []*wantWorker

	MaxIdleWorker int

	mu            sync.Mutex
	runningWorker int
}

func NewPool(maxIdleWorker int) *Pool {
	p := &Pool{
		MaxIdleWorker: maxIdleWorker,
	}

	return p
}

func (p *Pool) Run(fn task) {
	w := p.getWorker()
	w.run(fn)
}

func (p *Pool) getWorker() *worker {
	want := &wantWorker{
		ready: make(chan struct{}),
	}
	w := p.queueForWorker(want)
	if w != nil {
		return w
	}
	select {
	case <-want.ready:
		return want.w
	}
}

func (p *Pool) queueForWorker(want *wantWorker) *worker {
	p.idleMu.Lock()
	defer p.idleMu.Unlock()

	for len(p.idleWorker) > 0 {
		w := p.idleWorker[len(p.idleWorker)-1]
		p.idleWorker = p.idleWorker[:len(p.idleWorker)-1]

		if w.running() {
			return w
		}
	}

	if p.getRunningWorker() < p.MaxIdleWorker {
		w := &worker{
			pool: p,
			task: make(chan task),
			done: make(chan struct{}),
		}
		w.start()
		return w
	}

	p.wantQueue = append(p.wantQueue, want)
	return nil
}

func (p *Pool) tryPutWorker(w *worker) {
	p.idleMu.Lock()
	defer p.idleMu.Unlock()

	for len(p.wantQueue) > 0 {
		want := p.wantQueue[0]
		copy(p.wantQueue[:len(p.wantQueue)-1], p.wantQueue[1:len(p.wantQueue)])
		p.wantQueue = p.wantQueue[:len(p.wantQueue)-1]

		want.w = w
		close(want.ready)
	}
}

func (p *Pool) incRunningWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.runningWorker++
}

func (p *Pool) decRunningWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.runningWorker--
}

func (p *Pool) getRunningWorker() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.runningWorker
}

type wantWorker struct {
	ready chan struct{}
	w     *worker
}

type task func() error

type worker struct {
	pool *Pool
	task chan task
	done chan struct{}
}

func (w *worker) start() {
	w.pool.incRunningWorker()
	go func() {
		defer func() {
			w.pool.decRunningWorker()
			if r := recover(); r != nil {
				log.Printf("user goroutine exit with panic: %v", r)
			}
		}()
		for {
			select {
			case fn := <-w.task:
				fn()
				w.pool.tryPutWorker(w)
			case <-w.done:
				return
			}
		}
	}()
}

func (w *worker) run(fn task) {
	if !w.running() {
		log.Fatalf("attempt to run task on closed worker")
	}
	w.task <- fn
}

func (w *worker) close() {
	close(w.done)
}

func (w *worker) running() bool {
	select {
	case <-w.done:
		return false
	default:
		return true
	}
}
