package main

import (
	"context"
	"sync"
)

// Group 并发错误组，基于 https://github.com/ssoor/open-bilibili/blob/master/library/sync/errgroup.v2/errgroup.go
type Group struct {
	cancel func()

	wg sync.WaitGroup

	errOnce sync.Once
	err     error
}

// WithContext 创建带上下文的错误组
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

// WithLimit 创建限制并发数的错误组
func WithLimit(ctx context.Context, limit int) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{
		cancel: cancel,
	}, ctx
}

// Go 启动一个协程执行函数
func (g *Group) Go(f func() error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()

		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}

// Wait 等待所有协程完成并返回第一个错误
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}

// LimitedGroup 限制并发数的协程组
type LimitedGroup struct {
	group *Group
	sem   chan struct{}
	ctx   context.Context
}

// NewLimitedGroup 创建限制并发数的协程组
func NewLimitedGroup(ctx context.Context, limit int) (*LimitedGroup, context.Context) {
	g, ctx := WithContext(ctx)
	return &LimitedGroup{
		group: g,
		sem:   make(chan struct{}, limit),
		ctx:   ctx,
	}, ctx
}

// Go 在限制并发数的情况下启动协程
func (lg *LimitedGroup) Go(f func() error) {
	lg.group.Go(func() error {
		select {
		case lg.sem <- struct{}{}:
			defer func() { <-lg.sem }()
			return f()
		case <-lg.ctx.Done():
			return lg.ctx.Err()
		}
	})
}

// Wait 等待所有协程完成
func (lg *LimitedGroup) Wait() error {
	return lg.group.Wait()
}
