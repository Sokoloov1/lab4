package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Количество философов за столом.
const (
	numPhilosophers = 5
)

// Структура Fork представляет вилку, которую используют философы.
// Вилка защищена мьютексом, чтобы предотвратить одновременное использование.
type Fork struct {
	sync.Mutex
}

// Структура Philosopher представляет философа.
// Каждый философ имеет идентификатор, левую и правую вилку.
type Philosopher struct {
	id                int
	leftFork, rightFork *Fork
}

// Метод dine реализует процесс "обеда" философа.
// Философ думает и ест в бесконечном цикле, пока не получит сигнал о завершении.
func (p Philosopher) dine(wg *sync.WaitGroup, done chan struct{}) {
	defer wg.Done() // Уменьшаем счетчик WaitGroup при завершении.

	for {
		select {
		case <-done:
			// Если получен сигнал о завершении, философ заканчивает обедать.
			fmt.Printf("Философ %d закончил обедать.\n", p.id)
			return
		default:
			// Философ думает, затем ест.
			p.think()
			p.eat()
		}
	}
}

// Метод think реализует процесс "размышления" философа.
// Философ думает случайное количество времени.
func (p Philosopher) think() {
	fmt.Printf("Философ %d размышляет о великом.\n", p.id)
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
}

// Метод eat реализует процесс "еды" философа.
// Философ берет вилки, ест и затем кладет вилки обратно.
func (p Philosopher) eat() {
	// Чтобы избежать deadlock, философы с четными id берут сначала левую вилку,
	// а с нечетными — правую.
	if p.id%2 == 0 {
		p.leftFork.Lock()  // Блокируем левую вилку.
		p.rightFork.Lock() // Блокируем правую вилку.
	} else {
		p.rightFork.Lock() // Блокируем правую вилку.
		p.leftFork.Lock()  // Блокируем левую вилку.
	}

	// Философ ест случайное количество времени.
	fmt.Printf("Философ %d ест спагетти.\n", p.id)
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

	// Освобождаем вилки.
	p.leftFork.Unlock()
	p.rightFork.Unlock()
}

func main() {
	// Инициализируем генератор случайных чисел.
	rand.Seed(time.Now().UnixNano())

	// Создаем массив вилок. Каждая вилка представлена мьютексом.
	forks := make([]*Fork, numPhilosophers)
	for i := 0; i < numPhilosophers; i++ {
		forks[i] = &Fork{}
	}

	// Создаем массив философов.
	philosophers := make([]*Philosopher, numPhilosophers)
	for i := 0; i < numPhilosophers; i++ {
		philosophers[i] = &Philosopher{
			id:        i,                      // Уникальный идентификатор философа.
			leftFork:  forks[i],               // Левая вилка.
			rightFork: forks[(i+1)%numPhilosophers], // Правая вилка (круговая зависимость).
		}
	}

	// Используем WaitGroup для ожидания завершения всех горутин.
	var wg sync.WaitGroup
	// Канал done используется для сигнализации о завершении работы.
	done := make(chan struct{})

	// Запускаем горутины для каждого философа.
	for _, philosopher := range philosophers {
		wg.Add(1) // Увеличиваем счетчик WaitGroup.
		go philosopher.dine(&wg, done)
	}

	// Философы едят 5 секунд.
	time.Sleep(5 * time.Second)

	// Закрываем канал done, чтобы сигнализировать философам о завершении.
	close(done)

	// Ожидаем завершения всех горутин.
	wg.Wait()

	// Выводим сообщение о завершении.
	fmt.Println("Все философы закончили обедать.")
}
