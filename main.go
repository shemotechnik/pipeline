package main

import (
	"container/ring"
	"fmt"
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

// Интервал очистки кольцевого буфера
const bufferDrainInterval time.Duration = 3 * time.Second

// Размер кольцевого буфера
const bufferSize int = 2

func main(){
	// источник данных
	dataSource := func() (<-chan int, <-chan bool) {
		c := make(chan int)
		done := make(chan bool)
		go func() {
			defer close(done)
			scanner := bufio.NewScanner(os.Stdin)
			var data string
			for {
				scanner.Scan()
				data = scanner.Text()
				if strings.EqualFold(data, "exit") {
					fmt.Println("Программа завершила работу!")
					return
				}
				i, err := strconv.Atoi(data)
				if err != nil {
					fmt.Println("Программа обрабатывает только целые числа!")
					continue
				}
				c <- i
			}
		}()
		return c, done
	}

	negativeFilterInt := func(done <-chan bool, c <-chan int) <-chan int {
		convertedIntChan := make(chan int)
		go func() {
			for {
				select {
				case data := <-c:
					if data > 0 {
						convertedIntChan <- data
					}
				case <-done:
					return
				}
			}
		}()
		return convertedIntChan
	}

	specialFilter := func(done <-chan bool, c <-chan int) <-chan int {
		filteredIntChan := make(chan int)
		go func() {
			for {
				select {
				case data := <-c:
					if data != 0 && data%3 == 0 {
						select {
						case filteredIntChan <- data:
						case <-done:
							return
						}
					}
				case <-done:
					return
				}
			}
		}()
		return filteredIntChan
	}

	bufferInt := func(done <-chan bool, c <-chan int) <-chan int {
		bufferedIntChan := make(chan int)
		buffer := ring.New(bufferSize)
		go func() {
			for{
				select {
				case data:=<-c:
					buffer.Value = data
					buffer = buffer.Next()
				case <-done:
					return
				}
			}
		}()

		go func() {
			for {
				select {
				case <-time.After(bufferDrainInterval):
					var bufferData []int
					buffer.Do(func(x interface{}) {
						if x != nil{
							bufferData = append(bufferData, x.(int))
						}
					})
					buffer = ring.New(bufferSize)
					if bufferData != nil {
						for _, data := range bufferData {
							select {
							case bufferedIntChan <- data:
							case <-done:
								return
							}
						}
					}
				case <-done:
					return
				}
			}
		}()
		return bufferedIntChan
	}

	consumer := func(done <-chan bool, c <-chan int) {
		for {
			select {
			case data := <-c:
				fmt.Printf("Обработаны данные: %d\n", data)
			case <-done:
				return
			}
		}
	}

	source, done := dataSource()
	pipelineCh := bufferInt(done,specialFilter(done,negativeFilterInt(done,source)))
	consumer(done,pipelineCh)
}
