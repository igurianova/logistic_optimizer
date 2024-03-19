package domain

// Pair Структура хранящая пару значений типа int
type Pair struct {
	Left  int
	Right int
}

// GetElements Метод возвращающий пару значений в виде массива
func (p *Pair) GetElements() []int {
	return []int{p.Left, p.Right}
}
