package domain

type LRoute struct {
	Points        []int   `json:"warehouses"`
	First         int     `json:"first_warehouse"`
	Last          int     `json:"last_warehouse"`
	TotalCapacity float64 `json:"total_capacity"`
	TotalDistance float64 `json:"total_distance"`
	TotalSave     float64 `json:"total_save"`
}

// AtBorder Метод определяющий является ли точка началом или концом маршрута
func (r *LRoute) AtBorder(point int) bool {
	return r.First == point || r.Last == point
}

// AddPoint Метод добавляющий склад в конец маршрута
func (r *LRoute) AddPoint(point int, capacity, distance, save float64) {
	r.Points = append(r.Points, point)
	r.Last = point
	r.TotalCapacity += capacity
	r.TotalDistance += distance
	r.TotalSave += save
}

// InsertPoint Метод добавляющий склад в начало маршрута
func (r *LRoute) InsertPoint(point int, capacity, distance, save float64) {
	r.Points = append([]int{point}, r.Points...)
	r.First = point
	r.TotalCapacity += capacity
	r.TotalDistance += distance
	r.TotalSave += save
}

// MergeRoute Метод выполняющий слияние маршрутов
func (r *LRoute) MergeRoute(route *LRoute, routes *[]LRoute, wn *map[int]Node) {
	points := route.Points
	if r.Last == route.Last {
		reversedRoute := make([]int, 0, len(points))
		for i := len(points) - 1; i >= 0; i-- {
			reversedRoute = append(reversedRoute, points[i])
		}
		r.Last = reversedRoute[len(reversedRoute)-1]
		r.Points = append(r.Points, reversedRoute...)
	} else {
		r.Points = append(r.Points, points...)
		r.Last = points[len(points)-1]
	}
	r.TotalCapacity += route.TotalCapacity
	r.TotalDistance += route.TotalDistance
	r.TotalSave += route.TotalSave

	// Меняем номер маршрутов у каждого присоединенного пункта
	*routes = append(*routes, *r)
	for _, point := range route.Points {
		node := (*wn)[point]
		node.Ri = len(*routes) - 1
		(*wn)[point] = node
	}
}

// SmartMerge Метод выполняющий слияние маршрутов (эффективная версия)
func SmartMerge(lr *LRoute, rr *LRoute, routes *[]LRoute, wn *map[int]Node) {
	if len(lr.Points) > len(rr.Points) {
		lr.MergeRoute(rr, routes, wn)
	} else {
		rr.MergeRoute(lr, routes, wn)
	}
}
