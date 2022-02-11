package gpool

import (
	"log"
	"math"
	"sync"
	"sync/atomic"
)

var DefaultPool = NewPool(1000, math.MaxInt)

func Run(fn Task) {
	DefaultPool.Run(fn)
}

func Status() *PoolStatus {
	return DefaultPool.Status()
}

type Pool struct {
	workerCache sync.Pool

	idleMu     sync.Mutex
	idleWorker []*worker
	wantQueue  []*wantWorker

	MaxIdleWorker    int
	MaxRunningWorker int

	mu            sync.Mutex
	runningWorker int

	//debug param
	hit   int64
	total int64
}

type PoolStatus struct {
	Running   int
	Idle      int
	Waiting   int
	ReuseRate float32
}

func NewPool(maxIdle int, maxRunning int) *Pool {
	p := &Pool{
		MaxIdleWorker:    maxIdle,
		MaxRunningWorker: maxRunning,
	}

	p.workerCache.New = func() interface{} {
		w := &worker{
			pool: p,
			task: make(chan Task),
		}
		return w
	}

	return p
}

func (p *Pool) Run(fn Task) {
	atomic.AddInt64(&p.total, 1)
	w := p.getWorker()
	w.run(fn)
}

func (p *Pool) Status() *PoolStatus {
	stat := &PoolStatus{}
	stat.Running = p.getRunningWorker()

	p.idleMu.Lock()
	stat.Idle = len(p.idleWorker)
	stat.Waiting = len(p.wantQueue)
	p.idleMu.Unlock()

	stat.ReuseRate = float32(atomic.LoadInt64(&p.hit)) / float32(atomic.LoadInt64(&p.total)+1)

	return stat
}

func (p *Pool) getWorker() *worker {
	w, want := p.queueForWorker()
	if w != nil {
		return w
	}

	select {
	case <-want.ready:
		return want.w
	}
}

func (p *Pool) queueForWorker() (*worker, *wantWorker) {
	p.idleMu.Lock()
	defer p.idleMu.Unlock()

	for len(p.idleWorker) > 0 {
		w := p.idleWorker[len(p.idleWorker)-1]
		p.idleWorker = p.idleWorker[:len(p.idleWorker)-1]

		if w.running() {
			atomic.AddInt64(&p.hit, 1)
			return w, nil
		}
	}

	if p.getRunningWorker() < p.MaxRunningWorker {
		w := p.workerCache.Get().(*worker)
		w.start()
		return w, nil
	}

	want := &wantWorker{
		ready: make(chan struct{}),
	}
	p.wantQueue = append(p.wantQueue, want)
	return nil, want
}

func (p *Pool) tryPutWorker(w *worker) {
	p.idleMu.Lock()
	defer p.idleMu.Unlock()

	for len(p.wantQueue) > 0 {
		want := p.wantQueue[len(p.wantQueue)-1]
		p.wantQueue = p.wantQueue[:len(p.wantQueue)-1]

		want.w = w
		close(want.ready)
		return
	}

	if len(p.idleWorker) >= p.MaxIdleWorker {
		w.close()
		return
	}
	p.idleWorker = append(p.idleWorker, w)
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

func (p *Pool) getIdleWorker() int {
	p.idleMu.Lock()
	defer p.idleMu.Unlock()

	return len(p.idleWorker)
}

type wantWorker struct {
	ready chan struct{}
	w     *worker
}

type Task func()

type worker struct {
	pool *Pool
	task chan Task
	done chan struct{}
}

func (w *worker) start() {
	w.done = make(chan struct{})
	w.pool.incRunningWorker()
	go func() {
		defer func() {
			w.pool.workerCache.Put(w)
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

func (w *worker) run(fn Task) {
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
