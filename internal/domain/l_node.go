package domain

// Node Структура хранящая индекс маршрута (в котором есть этот склад) и кол-во связей для склада
type Node struct {
	Ri int // Route index
	Cc int // Communication Count
}
