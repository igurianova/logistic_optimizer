package service

import (
	"bufio"
	"fmt"
	"github.com/igurianova/logistic_optimizer/internal/domain"
	"github.com/igurianova/logistic_optimizer/internal/err"
	"mime/multipart"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// VrpCalculatorService Сервис выполянющий подбор оптимальных маршрутов
type VrpCalculatorService struct {
}

// Calculate - метод для расчета оптимального маршрута с помощью алгоритма Кларка-Райта
func (s *VrpCalculatorService) Calculate(r *http.Request) ([]domain.LRoute, error) {
	// Обрабатывается не  более 32 MB
	if r.ParseMultipartForm(32<<20) != nil {
		return nil, err.HTTPError{
			Msg:    "Multipart data must be less than 32 MB",
			Status: http.StatusBadRequest,
		}
	}

	// Объём перевозимого груза одной машиной
	cc, _ := strconv.ParseFloat(
		strings.ReplaceAll(r.FormValue("cargoCapacity"), ",", "."),
		64,
	) // cargo capacity

	// Парсим матрицу расстояний между пунктами
	m, mpe := parseMatrix(r) // matrix
	if mpe != nil {
		return nil, mpe
	}

	// Парсим данные об объёме продуктов, который необходимо доставить на склад
	wpc, wpe := parseWPC(r) // warehouse product capacity
	if wpe != nil {
		return nil, wpe
	}

	// Рассчитываем километровые выигрыши
	ps := countSavingForPairs(m) // pair savings

	// Сортируем пары по убыванию километрового выигрыша для составления оптимальных маршрутов
	pairs := sortPairsBySaving(ps)

	// Оптимальные маршруты
	var routes []domain.LRoute
	wn := make(map[int]domain.Node) // warehouses nodes

	// Обходим все пары начиная с той, у которой самый большой километровый выигрышь
	for _, pair := range pairs {
		fmt.Printf("Pair {%v, %v} with saving - %.5f\n", pair.Left, pair.Right, ps[pair])

		ln, lf := wn[pair.Left]  // left node, left found
		rn, rf := wn[pair.Right] // right node, right found

		if (lf && ln.Cc == 2) ||
			(rf && rn.Cc == 2) {
			// Один из складов включен в маршрут и уже имеет 2 связи (полное включение в маршрут)
			// Ищем другую пару
			continue
		}

		if !lf && !rf {
			// Связей нет у каждого склада, значит создаём новый маршрут
			createNewRoute(pair, m, ps, wpc, &routes, &wn)
		} else if lf && rf {
			// Оба склада имеют связь

			if ln.Ri == rn.Ri {
				continue
			} else {
				lr := routes[ln.Ri] // left route
				rr := routes[rn.Ri] // right route

				// Склады в паре являются граничными пунктами на маршруте
				if (lr.AtBorder(pair.Left) && rr.AtBorder(pair.Right)) ||
					(lr.AtBorder(pair.Right) && rr.AtBorder(pair.Left)) {

					if lr.TotalCapacity+rr.TotalCapacity < cc {
						// Соединяем маршруты
						mergeRoutes(&routes, &wn, ln, rn, pair, m, ps)
					} else {
						// Ищем другую пару
						continue
					}
				} else {
					// Ищем другую пару
					continue
				}
			}
		} else {
			addPointInRoute(&routes, &wn, pair, m, ps, wpc, cc)
		}
	}

	// Удаление пустых маршрутов
	oi := 0
	for _, route := range routes {
		if route.Points != nil && len(route.Points) != 0 {
			routes[oi] = route
			oi++
		}
	}
	routes = routes[:oi]

	fmt.Printf("\nPair count - %v\n", len(pairs))
	// Выводим получившиеся маршруты
	for i, route := range routes {
		if route.Points != nil && len(route.Points) != 0 {
			points := route.Points
			routeStr := "0"
			for _, point := range points {
				routeStr += fmt.Sprintf("-%v", point)
			}
			fmt.Printf(
				"№%v Summary distance: %0.3f, Summary saving: %0.3f, Total Quantity: %0.3f, Route length: %v, Route: %v\n",
				i+1, route.TotalDistance, route.TotalSave, route.TotalCapacity, len(points), routeStr,
			)
		}
	}

	return routes, nil
}

// createNewRoute Метод создающий новый маршрут
func createNewRoute(
	pair domain.Pair,
	m [][]float64,
	ps map[domain.Pair]float64,
	wpc map[int]float64,
	routes *[]domain.LRoute,
	wn *map[int]domain.Node,
) {
	elements := getSortedPairElements(pair.Left, pair.Right)
	distance := m[elements[0]][elements[1]]
	saving := ps[domain.Pair{Left: elements[0], Right: elements[1]}]

	route := domain.LRoute{
		Points:        pair.GetElements(),
		Last:          pair.Right,
		First:         pair.Left,
		TotalCapacity: wpc[pair.Left] + wpc[pair.Right],
		TotalDistance: distance,
		TotalSave:     saving,
	}
	*routes = append(*routes, route)
	(*wn)[pair.Left] = domain.Node{Ri: len(*routes) - 1, Cc: 1}
	(*wn)[pair.Right] = domain.Node{Ri: len(*routes) - 1, Cc: 1}
}

// mergeRoutes Метод выполняющий слияение маршрутов
func mergeRoutes(
	routes *[]domain.LRoute,
	wn *map[int]domain.Node,
	ln domain.Node,
	rn domain.Node,
	pair domain.Pair,
	matrix [][]float64,
	ps map[domain.Pair]float64,
) {
	lr := (*routes)[ln.Ri] // left route
	rr := (*routes)[rn.Ri] // right route
	domain.SmartMerge(&lr, &rr, routes, wn)

	// Добавляем связь для склада и обнуляем старый маршрут
	processRoutePoint(&(*routes)[ln.Ri], wn, pair.Left)
	processRoutePoint(&(*routes)[rn.Ri], wn, pair.Right)

	// Добавляем расстояние и выигрыши между пунктами по которым выполяли слияение
	elements := getSortedPairElements(pair.Left, pair.Right)
	routesCount := len(*routes) - 1
	(*routes)[routesCount].TotalDistance += matrix[elements[0]][elements[1]]
	(*routes)[routesCount].TotalSave += ps[domain.Pair{Left: elements[0], Right: elements[1]}]
}

// addPointInRoute Метод добавляющий пункт из пары к маршруту, где добавлен другой пункт
func addPointInRoute(
	routes *[]domain.LRoute,
	wn *map[int]domain.Node,
	pair domain.Pair,
	matrix [][]float64,
	ps map[domain.Pair]float64,
	wpc map[int]float64,
	cc float64,
) {
	// Определяем склад, включенный в маршрут
	var includedPoint int
	var includedNode domain.Node
	var excludedPoint int
	if node, ok := (*wn)[pair.Left]; ok {
		includedPoint, excludedPoint, includedNode = pair.Left, pair.Right, node
	} else if node, ok := (*wn)[pair.Right]; ok {
		includedPoint, excludedPoint, includedNode = pair.Right, pair.Left, node
	}

	// Добавляем к маршруту новый пункт
	route := &(*routes)[includedNode.Ri]
	if route.TotalCapacity+wpc[excludedPoint] < cc {
		elements := getSortedPairElements(excludedPoint, includedPoint)
		distance := matrix[elements[0]][elements[1]]
		saving := ps[domain.Pair{Left: elements[0], Right: elements[1]}]
		switch {
		case route.Last == includedPoint:
			route.AddPoint(excludedPoint, wpc[excludedPoint], distance, saving)
		case route.First == includedPoint:
			route.InsertPoint(excludedPoint, wpc[excludedPoint], distance, saving)
		default:
			return
		}
		(*wn)[excludedPoint] = domain.Node{Ri: includedNode.Ri, Cc: 1}
		includedNode.Cc += 1
		(*wn)[includedPoint] = includedNode
	}
}

// processRoutePoint Метод добавляющий связь для склада и обнуляющий старый маршрут
func processRoutePoint(route *domain.LRoute, wn *map[int]domain.Node, point int) {
	*route = domain.LRoute{}
	node := (*wn)[point]
	node.Cc += 1
	(*wn)[point] = node
}

// getSortedPairElements Метод принимает на вход пару значений и возвращает отсортированный массив из 2 значений
func getSortedPairElements(i, j int) [2]int {
	if i < j {
		return [2]int{i, j}
	} else {
		return [2]int{j, i}
	}
}

// parseMatrix Метод выполняющий парсинг файла с матрицей расстояний
func parseMatrix(r *http.Request) ([][]float64, error) {
	file, _, fe := r.FormFile("matrix")
	if fe != nil {
		return nil, err.HTTPError{
			Msg:    fmt.Sprintf("matrix is not found: %s", fe),
			Status: http.StatusBadRequest,
		}
	}

	var ce error
	defer func(file multipart.File) {
		ce = file.Close()
	}(file)
	if ce != nil {
		return nil, err.HTTPError{
			Msg:    fmt.Sprintf("reading file was failed: %s", ce),
			Status: http.StatusInternalServerError,
		}
	}

	fs := bufio.NewScanner(file)

	fs.Split(bufio.ScanLines)

	var matrix [][]float64
	for fs.Scan() {
		elements := strings.Split(fs.Text(), ";")
		distances := make([]float64, len(elements))
		for i, element := range elements {
			distances[i], _ = strconv.ParseFloat(
				strings.ReplaceAll(element, ",", "."),
				64,
			)
		}
		matrix = append(matrix, distances)
	}
	return matrix, nil
}

// parseWPC Метод выполняющий парсинг файла со спросом на продукт для каждого склада
func parseWPC(r *http.Request) (map[int]float64, error) {
	file, _, fe := r.FormFile("warehouse-product-capacity")
	if fe != nil {
		return nil, err.HTTPError{
			Msg:    fmt.Sprintf("warehouse product capacity is not found: %s", fe),
			Status: http.StatusBadRequest,
		}
	}

	var ce error
	defer func(file multipart.File) {
		ce = file.Close()
	}(file)
	if ce != nil {
		return nil, err.HTTPError{
			Msg:    fmt.Sprintf("reading file was failed: %s", ce),
			Status: http.StatusInternalServerError,
		}
	}

	fs := bufio.NewScanner(file)

	fs.Split(bufio.ScanLines)

	wpc := make(map[int]float64)

	for fs.Scan() {
		elements := strings.Split(fs.Text(), ";")
		warehouse, _ := strconv.Atoi(elements[0])
		productCapacity, _ := strconv.ParseFloat(
			strings.ReplaceAll(elements[1], ",", "."),
			64,
		)
		wpc[warehouse] = productCapacity
	}
	return wpc, nil
}

// countSavingForPairs Метод считающий километровые выигрыши для пар
func countSavingForPairs(matrix [][]float64) (pairSavings map[domain.Pair]float64) {
	pairSavings = make(map[domain.Pair]float64)

	for i := 1; i < len(matrix); i++ {
		for j := i + 1; j < len(matrix[i]); j++ {
			pairSavings[domain.Pair{Left: i, Right: j}] = matrix[0][i] + matrix[0][j] - matrix[i][j]
		}
	}
	return pairSavings
}

// sortPairsBySaving Метод выполняющий сортировку пар по километровым выигрышам
func sortPairsBySaving(pairSavings map[domain.Pair]float64) (pairs []domain.Pair) {
	pairs = make([]domain.Pair, 0, len(pairSavings))
	for pair := range pairSavings {
		pairs = append(pairs, pair)
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return pairSavings[pairs[i]] > pairSavings[pairs[j]]
	})
	return pairs
}
