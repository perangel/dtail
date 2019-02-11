package metrics

// Observable is an interface for an observable metric.
type Observable interface {
	Add(other Observable)
	Multiply(other Observable)
	Less(other Observable) bool
	Float() float64
	Reset()
	Clone() Observable
}

// Observables is a collection of Observable that implements sort.Interface
type Observables []Observable

func (o Observables) Len() int {
	return len(o)
}

func (o Observables) Less(i, j int) bool {
	return o[i].Less(o[j])
}

func (o Observables) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}
