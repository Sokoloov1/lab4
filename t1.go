package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Количество горутин, которые будут запущены для каждого теста
const numGoroutines = 10

// Генерация случайного ASCII символа в диапазоне от 33 до 126
func generateRandomASCII() byte {
	return byte(rand.Intn(94) + 33)
}

// StopWatch обертка для измерения времени выполнения функции
func StopWatch(name string, f func()) {
	start := time.Now()
	f()
	duration := time.Since(start)
	fmt.Printf("%s Time: %v\n", name, duration)
}

// Тест Mutex: использует мьютекс для синхронизации доступа к общему ресурсу
func testMutex(wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()
	mu.Lock()
	fmt.Printf("Mutex: %c\n", generateRandomASCII())
	mu.Unlock()
}

// Тест Semaphore: использует канал с буфером для ограничения количества одновременно работающих горутин
func testSemaphore(wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()
	sem <- struct{}{} // Захват семафора
	fmt.Printf("Semaphore: %c\n", generateRandomASCII())
	<-sem // Освобождение семафора
}

// Тест SemaphoreSlim: использует семафор с ограниченным количеством попыток захвата
func testSemaphoreSlim(wg *sync.WaitGroup, sem chan struct{}, retries int) {
	defer wg.Done()
	for i := 0; i < retries; i++ {
		select {
		case sem <- struct{}{}: // Попытка захвата семафора
			fmt.Printf("SemaphoreSlim: %c\n", generateRandomASCII())
			<-sem // Освобождение семафора
			return
		default:
			time.Sleep(time.Millisecond * 10) // Ожидание перед следующей попыткой
		}
	}
}

// Тест Barrier: использует WaitGroup для синхронизации горутин в точке барьера
func testBarrier(wg *sync.WaitGroup, barrier *sync.WaitGroup) {
	defer wg.Done()
	barrier.Done() // Уменьшение счетчика барьера
	barrier.Wait() // Ожидание, пока все горутины достигнут барьера
	fmt.Printf("Barrier: %c\n", generateRandomASCII())
}

// Тест SpinLock: использует атомарные операции для реализации спин-лока
func testSpinLock(counter *int32) {
	for {
		if atomic.CompareAndSwapInt32(counter, 0, 1) { // Попытка захвата спин-лока
			fmt.Printf("SpinLock: %c\n", generateRandomASCII())
			atomic.StoreInt32(counter, 0) // Освобождение спин-лока
			break
		}
	}
}

// Тест SpinWait: активное ожидание с контролем количества итераций
func testSpinWait() {
	spinCount := 0
	for spinCount < 1000 { // Активное ожидание с контролем
		if spinCount%100 == 0 { // Добавление пауз через каждые 100 итераций
			randomDelay := time.Duration(rand.Intn(10)+1) * time.Microsecond
			time.Sleep(randomDelay)
		}
		spinCount++
	}
	fmt.Printf("SpinWait: %c\n", generateRandomASCII())
}

// Тест Monitor: использует мьютекс и условную переменную для синхронизации
func testMonitor(wg *sync.WaitGroup, mu *sync.Mutex, cond *sync.Cond) {
	defer wg.Done()
	mu.Lock()
	cond.Wait() // Ожидание сигнала от условной переменной
	fmt.Printf("Monitor: %c\n", generateRandomASCII())
	mu.Unlock()
}

func main() {
	rand.Seed(time.Now().UnixNano()) // Инициализация генератора случайных чисел
	var wg sync.WaitGroup

	// Тест Mutex
	mu := &sync.Mutex{}
	StopWatch("Mutex", func() {
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go testMutex(&wg, mu)
		}
		wg.Wait()
	})

	// Тест Semaphore
	sem := make(chan struct{}, 3) // Ограничение на 3 горутины
	StopWatch("Semaphore", func() {
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go testSemaphore(&wg, sem)
		}
		wg.Wait()
	})

	// Тест SemaphoreSlim
	retries := 5
	StopWatch("SemaphoreSlim", func() {
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go testSemaphoreSlim(&wg, sem, retries)
		}
		wg.Wait()
	})

	// Тест Barrier
	barrier := &sync.WaitGroup{}
	barrier.Add(numGoroutines)
	StopWatch("Barrier", func() {
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go testBarrier(&wg, barrier)
		}
		wg.Wait()
	})

	// Тест SpinLock
	var counter int32
	StopWatch("SpinLock", func() {
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				testSpinLock(&counter)
			}()
		}
		wg.Wait()
	})

	// Тест SpinWait
	StopWatch("SpinWait", func() {
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				testSpinWait()
			}()
		}
		wg.Wait()
	})

	// Тест Monitor
	mu = &sync.Mutex{}
	cond := sync.NewCond(mu)
	StopWatch("Monitor", func() {
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go testMonitor(&wg, mu, cond)
		}
		time.Sleep(time.Microsecond * 1000) // Даём время горутинам заблокироваться
		cond.Broadcast() // Сигнал всем горутинам для продолжения
		wg.Wait()
	})
}
