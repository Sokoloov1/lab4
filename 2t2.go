package main

import (
	"fmt"
	"sync"
	"time"
	"math/rand"
)

// Структура Worker представляет информацию о работнике.
type Worker struct {
	Name      string
	Position  string
	Age       int
	Salary    float64
}

// Функция calculateAverageAge вычисляет средний возраст работников для указанной должности (position).
func calculateAverageAge(workers []Worker, position string) float64 {
	var totalAge, count int

	// Проходим по каждому работнику в списке.
	for _, worker := range workers {
		// Если должность работника совпадает с искомой, учитываем его возраст.
		if worker.Position == position {
			totalAge += worker.Age // Суммируем возраст.
			count++                // Увеличиваем счетчик работников.
		}
	}

	// Если работников с указанной должностью не найдено, возвращаем 0.
	if count == 0 {
		return 0
	}

	// Возвращаем средний возраст как отношение суммы возрастов к количеству работников.
	return float64(totalAge) / float64(count)
}

// Функция findMaxSalary находит максимальную зарплату среди работников
func findMaxSalary(workers []Worker, position string, avgAge float64) float64 {
	var maxSalary float64

	// Проходим по каждому работнику в списке.
	for _, worker := range workers {
		// Если должность работника совпадает с искомой и его возраст близок к среднему,
		// проверяем его зарплату.
		if worker.Position == position && abs(float64(worker.Age)-avgAge) <= 2 {
			// Если зарплата текущего работника больше максимальной, обновляем максимальную зарплату.
			if worker.Salary > maxSalary {
				maxSalary = worker.Salary
			}
		}
	}

	// Возвращаем максимальную зарплату.
	return maxSalary
}

// Функция abs возвращает абсолютное значение числа.
// Используется для вычисления разницы между возрастом работника и средним возрастом.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Функция processWithoutConcurrency обрабатывает данные без использования многозадачности.
func processWithoutConcurrency(workers []Worker, position string) {
	// Засекаем время начала выполнения.
	start := time.Now()

	var avgAge float64
	var maxSalary float64

	// Количество частей, на которые разбиваем данные.
	countSize := 3
	// Размер каждой части.
	subsetsSize := len(workers) / countSize

	// Срезы для хранения промежуточных результатов.
	avgAgeResults := make([]float64, countSize)
	maxSalaryResults := make([]float64, countSize)

	// Поиск среднего возраста в каждой части данных.
	for i := 0; i < countSize; i++ {
		func(i int) {
			// Определяем начальный и конечный индексы для текущей части.
			startIndex := i * subsetsSize
			endIndex := (i + 1) * subsetsSize
			// Для последней части корректируем конечный индекс, чтобы не выйти за пределы списка.
			if i == countSize-1 {
				endIndex = len(workers)
			}
			// Вычисляем средний возраст для текущей части.
			avgAgeResults[i] = calculateAverageAge(workers[startIndex:endIndex], position)
		}(i)
	}

	// Объединяем результаты среднего возраста.
	var totalAge float64
	var count int
	for _, avg := range avgAgeResults {
		if avg > 0 {
			totalAge += avg
			count++
		}
	}
	// Вычисляем общий средний возраст.
	if count > 0 {
		avgAge = totalAge / float64(count)
	}

	// Поиск максимальной зарплаты в каждой части данных.
	for i := 0; i < countSize; i++ {
		func(i int) {
			// Определяем начальный и конечный индексы для текущей части.
			startIndex := i * subsetsSize
			endIndex := (i + 1) * subsetsSize
			// Для последней части корректируем конечный индекс.
			if i == countSize-1 {
				endIndex = len(workers)
			}
			// Находим максимальную зарплату для текущей части.
			maxSalaryResults[i] = findMaxSalary(workers[startIndex:endIndex], position, avgAge)
		}(i)
	}

	// Объединяем результаты максимальной зарплаты.
	for _, max := range maxSalaryResults {
		if max > maxSalary {
			maxSalary = max
		}
	}

	// Вычисляем время выполнения.
	duration := time.Since(start)

	// Выводим результаты.
	fmt.Printf("Без многозадачности:\n")
	fmt.Printf("Средний возраст: %.2f\n", avgAge)
	fmt.Printf("Максимальная зарплата: %.2f\n", maxSalary)
	fmt.Printf("Время обработки: %v\n\n", duration)
}

// Функция processWithConcurrency обрабатывает данные с использованием многозадачности (горутин).
func processWithConcurrency(workers []Worker, position string) {
	// Засекаем время начала выполнения.
	start := time.Now()

	// Используем WaitGroup для синхронизации горутин.
	var wg sync.WaitGroup
	var avgAge float64
	var maxSalary float64

	// Количество горутин.
	numGoroutines := 3
	// Размер каждой части данных.
	chunkSize := len(workers) / numGoroutines

	// Срезы для хранения промежуточных результатов.
	avgAgeResults := make([]float64, numGoroutines)
	maxSalaryResults := make([]float64, numGoroutines)

	// Запускаем горутины для вычисления среднего возраста.
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done() // Уменьшаем счетчик WaitGroup при завершении горутины.
			// Определяем начальный и конечный индексы для текущей части.
			startIndex := i * chunkSize
			endIndex := (i + 1) * chunkSize
			// Для последней части корректируем конечный индекс.
			if i == numGoroutines-1 {
				endIndex = len(workers)
			}
			// Вычисляем средний возраст для текущей части.
			avgAgeResults[i] = calculateAverageAge(workers[startIndex:endIndex], position)
		}(i)
	}

	// Ждем завершения всех горутин.
	wg.Wait()

	// Объединяем результаты среднего возраста.
	var totalAge float64
	var count int
	for _, avg := range avgAgeResults {
		if avg > 0 {
			totalAge += avg
			count++
		}
	}
	// Вычисляем общий средний возраст.
	if count > 0 {
		avgAge = totalAge / float64(count)
	}

	// Запускаем горутины для поиска максимальной зарплаты.
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done() // Уменьшаем счетчик WaitGroup при завершении горутины.
			// Определяем начальный и конечный индексы для текущей части.
			startIndex := i * chunkSize
			endIndex := (i + 1) * chunkSize
			// Для последней части корректируем конечный индекс.
			if i == numGoroutines-1 {
				endIndex = len(workers)
			}
			// Находим максимальную зарплату для текущей части.
			maxSalaryResults[i] = findMaxSalary(workers[startIndex:endIndex], position, avgAge)
		}(i)
	}

	// Ждем завершения всех горутин.
	wg.Wait()

	// Объединяем результаты максимальной зарплаты.
	for _, max := range maxSalaryResults {
		if max > maxSalary {
			maxSalary = max
		}
	}

	// Вычисляем время выполнения.
	duration := time.Since(start)

	// Выводим результаты.
	fmt.Printf("С многозадачностью (с несколькими горутинами):\n")
	fmt.Printf("Средний возраст: %.2f\n", avgAge)
	fmt.Printf("Максимальная зарплата: %.2f\n", maxSalary)
	fmt.Printf("Время обработки: %v\n\n", duration)
}

// Функция generateWorker генерирует случайного работника.
func generateWorker(index int) Worker {
	// Генерируем имя по шаблону.
	name := fmt.Sprintf("Работник %d", index)
	// Случайным образом выбираем должность: "Д" или "С".
	position := "Д"
	if rand.Intn(2) == 0 {
		position = "С"
	}
	// Генерируем случайный возраст от 20 до 60 лет.
	age := rand.Intn(41) + 20
	// Генерируем случайную зарплату от 30 000 до 100 000.
	salary := float64(rand.Intn(70000) + 30000

	// Возвращаем структуру Worker с заполненными полями.
	return Worker{
		Name:     name,
		Position: position,
		Age:      age,
		Salary:   salary,
	}
}

// Основная функция программы.
func main() {
	// Инициализируем генератор случайных чисел.
	rand.Seed(time.Now().UnixNano())

	// Создаем массив работников размером 100 000.
	var workers []Worker
	for i := 0; i < 100000; i++ {
		// Генерируем работника и добавляем его в массив.
		workers = append(workers, generateWorker(i))
	}

	// Указываем должность для анализа.
	position := "Д"

	// Обработка данных без многозадачности.
	processWithoutConcurrency(workers, position)

	// Обработка данных с многозадачностью.
	processWithConcurrency(workers, position)
}
