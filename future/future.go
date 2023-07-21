package future

type result[T, K any] struct {
	t T
	k K
}

type Future[T, K any] struct {
	c <-chan *result[T, K]
	r *result[T, K]
}

func New[T, K any](fn func() (T, K)) *Future[T, K] {
	c := make(chan *result[T, K])
	fut := &Future[T, K]{c: c}

	go func() {
		r := result[T, K]{}
		r.t, r.k = fn()
		c <- &r
	}()

	return fut

}

func (f *Future[T, K]) Await() (T, K) {
	if f.r == nil {
		f.r = <-f.c
	}

	return f.r.t, f.r.k
}
