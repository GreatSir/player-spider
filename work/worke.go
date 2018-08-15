package work

import "sync"

/*
定义worker接口
 */
type Worker interface {
	Task()
}
type Pool struct {
	work chan Worker
	wg sync.WaitGroup
}

func New(maxGo int) *Pool {
	p:=Pool{
		work:make(chan Worker),
	}
	p.wg.Add(maxGo)
	for i:=0;i<maxGo;i++{
		go func() {
			for w:=range p.work{
				w.Task()
			}
			p.wg.Done()
		}()
	}
	return &p
}
func (p *Pool)Run(w Worker)  {
	p.work <- w
}
func (p *Pool)ShutDown()  {
	close(p.work)
	p.wg.Wait()
}