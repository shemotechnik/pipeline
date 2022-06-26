package main
/*
Пайплайн на каналах
*/
import (
	"container/ring"
	"fmt"
	"log"
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
			var data string
			for {
				fmt.Scanln(&data)
				if data != "" {
					if strings.EqualFold(data, "exit") {
						fmt.Println("Программа завершила работу!")
						log.Printf("%s\n", "exit from program")
						return
					}
					i, err := strconv.Atoi(data)
					if err != nil {
						fmt.Println("Программа обрабатывает только целые числа!")
						return
					}
					c <- i
					log.Printf("%s %d\n", "add new information to chanel:", i)
				}
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
						log.Printf("%s %d\n","filter value to chanel chanel:",data)
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
							log.Printf("%s %d\n","filter value that can be devided by 3 to chanel:", data)
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
								log.Printf("%s %d\n","put data to chanel:", data)
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
				log.Printf("%s %d\n","print data to chanel:",data)
			case <-done:
				return
			}
		}
	}

	source, done := dataSource()
	pipelineCh := bufferInt(done,specialFilter(done,negativeFilterInt(done,source)))
	consumer(done,pipelineCh)
}
